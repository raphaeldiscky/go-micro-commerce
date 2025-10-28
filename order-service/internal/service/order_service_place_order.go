package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/redisutils"
)

// PlaceOrder handles the synchronous order placement flow.
// Flow: Get checkout → Validate → Reserve products → Calculate shipping →
// Create payment intent → Create order → Publish event → Return order + payment data.
//
//nolint:funlen,gocyclo,cyclop,govet,gocognit // TODO: refactor
func (s *orderService) PlaceOrder(
	ctx context.Context,
	req *dto.PlaceOrderRequest,
) (*dto.PlaceOrderResponse, error) {
	s.logger.Infof(
		"Placing order for customer %s with checkout session %s",
		req.CustomerID,
		req.CheckoutSessionID,
	)

	// Distributed lock for idempotency
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
		),
	}

	lock, err := lockRepo.Get(ctx, lockKey, ttl, opt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	var res *dto.PlaceOrderResponse

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		outboxRepo := ds.OutboxRepository()

		// Check for existing order by idempotency key
		existingOrder, errExist := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if errExist != nil && errExist.Error() != constant.OrderNotFoundErrorMessage {
			return httperror.NewInternalServerError("failed to check existing order")
		}

		if existingOrder != nil {
			s.logger.Infof("Order already exists for idempotency key %s", req.IdempotencyKey)
			// TODO: Retrieve payment data for existing order
			res = &dto.PlaceOrderResponse{
				Order: mapper.MapToOrderResponse(existingOrder),
			}

			return nil
		}

		// Step 1: Get checkout session from cart-service
		s.logger.Infof("Fetching checkout session %s from cart-service", req.CheckoutSessionID)

		checkoutSession, errn := s.cartClient.GetCheckoutSession(ctx, req.CheckoutSessionID)
		if errn != nil {
			s.logger.Errorf("Failed to get checkout session: %v", errn)

			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to get checkout session: %v", errn),
			)
		}

		// Step 2: Validate checkout session
		if checkoutSession.Status != "STATUS_PENDING" {
			return httperror.NewBadRequestError(
				fmt.Sprintf("invalid checkout session status: %s", checkoutSession.Status),
			)
		}

		if checkoutSession.CustomerID != req.CustomerID {
			return httperror.NewForbiddenError("checkout session does not belong to customer")
		}

		if len(checkoutSession.Items) == 0 {
			return httperror.NewBadRequestError("checkout session has no items")
		}

		// Determine payment gateway (default to Stripe if not specified)
		paymentGateway := constant.PaymentGatewayStripe
		if checkoutSession.PaymentGateway != nil && *checkoutSession.PaymentGateway != "" {
			paymentGateway = constant.PaymentGateway(*checkoutSession.PaymentGateway)
		}

		// Step 3: Reserve products and get prices
		s.logger.Infof("Reserving %d products", len(checkoutSession.Items))

		reservationItems := make([]dto.ProductReservationItem, len(checkoutSession.Items))
		for i, item := range checkoutSession.Items {
			reservationItems[i] = dto.ProductReservationItem{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				ExpectedVersion: 0, // Will be validated by product-service
			}
		}

		reservedProducts, err := s.productClient.ReserveProducts(
			ctx,
			req.IdempotencyKey,
			reservationItems,
		)
		if err != nil {
			s.logger.Errorf("Failed to reserve products: %v", err)
			return httperror.NewBadRequestError(fmt.Sprintf("failed to reserve products: %v", err))
		}

		// Create order items using reserved product prices
		orderItems := make([]entity.OrderItem, len(reservedProducts))
		for i, product := range reservedProducts {
			quantity := checkoutSession.Items[i].Quantity

			orderItems[i] = entity.OrderItem{
				ProductID:     product.ID,
				Quantity:      quantity,
				UnitPrice:     product.UnitPrice,
				TaxRate:       decimal.Zero,
				TotalTax:      decimal.Zero,
				TotalDiscount: decimal.Zero,
			}
		}

		// Step 4: Calculate shipping cost
		s.logger.Infof("Calculating shipping cost with fulfillment-service")

		shippingReq := dto.CalculateShippingRequest{
			Courier: dto.Courier{
				CourierID: checkoutSession.Courier.CourierID,
			},
			Destination: dto.ToAddress{
				City:        checkoutSession.Destination.City,
				State:       checkoutSession.Destination.State,
				PostalCode:  checkoutSession.Destination.PostalCode,
				CountryCode: checkoutSession.Destination.CountryCode,
			},
			Origin: dto.FromAddress{
				City:        checkoutSession.Origin.City,
				State:       checkoutSession.Origin.State,
				PostalCode:  checkoutSession.Origin.PostalCode,
				CountryCode: checkoutSession.Origin.CountryCode,
			},
			Package: dto.Package{
				WeightKG: checkoutSession.Package.WeightKG,
				Length:   checkoutSession.Package.Length,
				Width:    checkoutSession.Package.Width,
				Height:   checkoutSession.Package.Height,
				Unit:     checkoutSession.Package.Unit,
			},
			Currency: checkoutSession.Currency,
		}

		shippingResp, err := s.fulfillmentClient.CalculateShipping(ctx, shippingReq)
		if err != nil {
			s.logger.Errorf("Failed to calculate shipping: %v", err)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to calculate shipping: %v", err),
			)
		}

		// Step 5: Create order entity using NewOrder (calculates subtotal and totals)
		order, err := entity.NewOrder(
			req.CustomerID,
			req.IdempotencyKey,
			req.CheckoutSessionID,
			paymentGateway,
			checkoutSession.Currency,
			entity.Courier{
				CourierID: checkoutSession.Courier.CourierID,
			},
			entity.Destination{
				City:        checkoutSession.Destination.City,
				State:       checkoutSession.Destination.State,
				PostalCode:  checkoutSession.Destination.PostalCode,
				CountryCode: checkoutSession.Destination.CountryCode,
			},
			entity.Origin{
				City:        checkoutSession.Origin.City,
				State:       checkoutSession.Origin.State,
				PostalCode:  checkoutSession.Origin.PostalCode,
				CountryCode: checkoutSession.Origin.CountryCode,
			},
			entity.Package{
				WeightKG: checkoutSession.Package.WeightKG,
				Length:   checkoutSession.Package.Length,
				Width:    checkoutSession.Package.Width,
				Height:   checkoutSession.Package.Height,
				Unit:     checkoutSession.Package.Unit,
			},
			orderItems,
		)
		if err != nil {
			s.logger.Errorf("Failed to create order entity: %v", err)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewBadRequestError(fmt.Sprintf("failed to create order: %v", err))
		}

		// Step 6: Update shipping cost (recalculates total price)
		if err = order.UpdateShippingCost(shippingResp.Cost); err != nil {
			s.logger.Errorf("Failed to update shipping cost: %v", err)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to update shipping cost: %v", err),
			)
		}

		s.logger.Infof(
			"Order totals: subtotal=%s, shipping=%s, total=%s",
			order.Subtotal,
			order.ShippingCost,
			order.TotalPrice,
		)

		// Step 7: Create payment intent synchronously
		s.logger.Infof(
			"Creating payment intent for order with amount %s %s",
			order.TotalPrice,
			order.Currency,
		)

		paymentIntent, err := s.paymentClientGRPC.CreatePaymentIntent(
			ctx,
			order.ID,
			order.TotalPrice,
			order.Currency,
			string(paymentGateway),
			req.CustomerID,
			req.CustomerEmail,
		)
		if err != nil {
			s.logger.Errorf("Failed to create payment intent: %v", err)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewInternalServerError(
				fmt.Sprintf("failed to create payment intent: %v", err),
			)
		}

		// Step 8: Update order status to pending_payment
		if err = order.UpdateStatus(constant.OrderStatusPaymentPending); err != nil {
			s.logger.Errorf("Failed to update order status: %v", err)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to update order status: %v", err),
			)
		}

		// Step 9: Save order to database
		savedOrder, err := orderRepo.Create(ctx, order)
		if err != nil {
			s.logger.Errorf("Failed to save order: %v", err)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewInternalServerError("failed to save order")
		}

		s.logger.Infof("Order created successfully: order_id=%s", savedOrder.ID)

		// Step 9: Publish OrderCreatedEvent via outbox
		evt := producer.NewOrderLifecycleEvent(
			savedOrder.ID,
			savedOrder.CheckoutSessionID,
			savedOrder.Status,
			savedOrder.CustomerID,
			savedOrder.TotalPrice,
			savedOrder.Currency,
			savedOrder.Items,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal order event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "order",
			AggregateID:   savedOrder.ID,
			EventType:     kafka.OrderCreatedEventType,
			Topic:         kafka.OrderLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create outbox event")
		}

		// Step 10: Build response with order + payment metadata
		res = &dto.PlaceOrderResponse{
			Order: mapper.MapToOrderResponse(savedOrder),
			PaymentMetadata: dto.PaymentMetadata{
				PaymentID:            paymentIntent.PaymentID,
				PaymentGateway:       paymentGateway,
				GatewayTransactionID: paymentIntent.GatewayTransactionID,
				GatewayMetadata:      paymentIntent.GatewayMetadata,
				Amount:               paymentIntent.Amount,
				Currency:             paymentIntent.Currency,
			},
		}

		s.logger.Infof(
			"PlaceOrder completed: order_id=%s, payment_id=%s, gateway_transaction_id=%s",
			savedOrder.ID,
			paymentIntent.PaymentID,
			paymentIntent.GatewayTransactionID,
		)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// compensateProductReservation releases reserved products in case of failure.
func (s *orderService) compensateProductReservation(
	ctx context.Context,
	items []dto.ProductReservationItem,
) error {
	restorationItems := make([]dto.ProductRestorationItem, len(items))
	for i, item := range items {
		restorationItems[i] = dto.ProductRestorationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return s.productClient.ReleaseProducts(ctx, restorationItems)
}
