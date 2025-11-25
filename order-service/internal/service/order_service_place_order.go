package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"golang.org/x/sync/errgroup"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/redisutils"
)

// PlaceOrder handles the synchronous order placement flow.
// Flow: Get checkout -> Validate -> Reserve products -> Calculate shipping ->
// Create payment intent -> Create order -> Publish event -> Return order + payment data.
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

		// Extract product IDs from checkout session for parallel fetch
		productIDs := make([]uuid.UUID, len(checkoutSession.Items))
		for i, item := range checkoutSession.Items {
			productIDs[i] = item.ProductID
		}

		// Step 3 & 4: Fetch products AND calculate shipping IN PARALLEL
		s.logger.Infof(
			"Starting parallel operations: fetching %d products and calculating shipping",
			len(productIDs),
		)

		var (
			products     []entity.Product
			shippingResp *dto.CalculateShippingResponse
		)

		g, gctx := errgroup.WithContext(ctx)

		// Goroutine 1: Fetch current products from product-service
		g.Go(func() error {
			s.logger.Infof("Fetching products concurrently")

			fetchedProducts, errFetch := s.productClient.GetProducts(gctx, productIDs)
			if errFetch != nil {
				s.logger.Errorf("Failed to fetch products: %v", errFetch)
				return fmt.Errorf("failed to fetch products: %w", errFetch)
			}

			products = fetchedProducts
			s.logger.Infof("Successfully fetched %d products", len(products))

			return nil
		})

		// Goroutine 2: Calculate shipping cost concurrently
		g.Go(func() error {
			s.logger.Infof("Calculating shipping cost concurrently")

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

			resp, errShipping := s.fulfillmentClient.CalculateShipping(gctx, shippingReq)
			if errShipping != nil {
				s.logger.Errorf("Failed to calculate shipping: %v", errShipping)
				return fmt.Errorf("failed to calculate shipping: %w", errShipping)
			}

			shippingResp = resp
			s.logger.Infof(
				"Successfully calculated shipping: %s %s",
				resp.Cost,
				checkoutSession.Currency,
			)

			return nil
		})

		// Wait for both parallel operations to complete
		if err = g.Wait(); err != nil {
			s.logger.Errorf("Parallel operations failed: %v", err)
			return httperror.NewBadRequestError(err.Error())
		}

		// Create product map for quick lookup
		productMap := make(map[uuid.UUID]*entity.Product)
		for i := range products {
			productMap[products[i].ID] = &products[i]
		}

		// Build reservation items with validation
		reservationItems := make([]dto.ProductReservationItem, len(checkoutSession.Items))
		for i, item := range checkoutSession.Items {
			product, exists := productMap[item.ProductID]
			if !exists {
				return httperror.NewBadRequestError(
					fmt.Sprintf("product %s not found", item.ProductID),
				)
			}

			// Validate price hasn't changed since checkout
			if !product.UnitPrice.Equal(item.UnitPrice) {
				return httperror.NewBadRequestError(
					fmt.Sprintf(
						"price changed for product %s (%s): expected %s, current %s",
						product.ID,
						product.Name,
						item.UnitPrice.String(),
						product.UnitPrice.String(),
					),
				)
			}

			// Check available stock (total quantity - reserved quantity)
			availableStock := product.Quantity - product.ReservedQuantity
			if availableStock < item.Quantity {
				return httperror.NewBadRequestError(
					fmt.Sprintf(
						"insufficient stock for product %s (%s): requested %d, available %d",
						product.ID,
						product.Name,
						item.Quantity,
						availableStock,
					),
				)
			}

			// Use fresh version from product-service
			reservationItems[i] = dto.ProductReservationItem{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				ExpectedVersion: product.Version,
			}
		}

		// Step 5: Reserve products with validated items and fresh versions
		s.logger.Infof("Reserving %d products with fresh versions", len(reservationItems))

		reservedProducts, errn := s.productClient.ReserveProducts(
			ctx,
			req.IdempotencyKey,
			reservationItems,
		)
		if errn != nil {
			s.logger.Errorf("Failed to reserve products: %v", errn)
			return httperror.NewBadRequestError(fmt.Sprintf("failed to reserve products: %v", errn))
		}

		// Create order items using reserved product prices and product names
		orderItems := make([]entity.OrderItem, len(reservedProducts))
		for i, product := range reservedProducts {
			quantity := checkoutSession.Items[i].Quantity

			orderItem, err := entity.NewOrderItem(
				product.ID,
				product.Name, // Add product name snapshot
				quantity,
				product.UnitPrice,
			)
			if err != nil {
				return httperror.NewBadRequestError(
					fmt.Sprintf("failed to create order item for product %s: %v", product.ID, err),
				)
			}

			orderItems[i] = *orderItem
		}

		// Step 6: Create order entity using NewOrder (calculates subtotal and totals)
		order, errn := entity.NewOrder(
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
		if errn != nil {
			s.logger.Errorf("Failed to create order entity: %v", errn)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewBadRequestError(fmt.Sprintf("failed to create order: %v", err))
		}

		// Step 7: Update shipping cost (recalculates total price)
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

		// Step 8: Create payment intent synchronously
		s.logger.Infof(
			"Creating payment intent for order with amount %s %s",
			order.TotalPrice,
			order.Currency,
		)

		paymentIntent, errn := s.paymentClientGRPC.CreatePaymentIntent(
			ctx,
			order.ID,
			order.TotalPrice,
			order.Currency,
			string(paymentGateway),
			req.CustomerID,
			req.CustomerEmail,
		)
		if errn != nil {
			s.logger.Errorf("Failed to create payment intent: %v", errn)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewInternalServerError(
				fmt.Sprintf("failed to create payment intent: %v", err),
			)
		}

		// Step 9: Update order status to pending_payment
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

		// Step 10: Save order to database
		savedOrder, errn := orderRepo.Create(ctx, order)
		if errn != nil {
			s.logger.Errorf("Failed to save order: %v", errn)
			// Compensate: Release reserved products
			err = s.compensateProductReservation(ctx, reservationItems)
			if err != nil {
				s.logger.Errorf("Failed to release reserved products: %v", err)
			}

			return httperror.NewInternalServerError("failed to save order")
		}

		s.logger.Infof("Order created successfully: order_id=%s", savedOrder.ID)

		// Step 11: Create saga state for post-payment workflow
		// Store metadata that will be needed when payment succeeds
		userAuth, errAuth := echoutils.GetUserAuthContexts(ctx)
		if errAuth != nil {
			s.logger.Warnf("Failed to extract user auth for saga state: %v", errAuth)
			// Continue without user auth - saga can still proceed
			userAuth = pkgdto.UserAuthInfo{}
		}

		sagaMetadata := &saga.Metadata{
			ReservedProducts: reservedProducts,
			CustomerEmail:    req.CustomerEmail,
			PaymentID:        &paymentIntent.PaymentID,
			UserAuth:         &userAuth,
		}

		sagaRepo := ds.SagaStateRepository()
		sagaState := &entity.SagaState{
			ID:               uuid.New(),
			WorkflowName:     constant.PostPaymentSagaWorkflowName,
			OrderID:          savedOrder.ID,
			CurrentStep:      0,
			Status:           constant.SagaStatusPending,
			ExecutedSteps:    []string{},
			CompensatedSteps: []string{},
			Data:             sagaMetadata.ToMap(),
			Version:          1,
			RetryCount:       0,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		if err = sagaRepo.Create(ctx, sagaState); err != nil {
			s.logger.Errorf("Failed to create post-payment saga state: %v", err)
			// Don't fail the order creation - saga can be recovered later
		} else {
			s.logger.Infof(
				"Created post-payment saga state %s for order %s",
				sagaState.ID,
				savedOrder.ID,
			)
		}

		// Step 12: Publish OrderCreatedEvent via outbox
		evt := producer.NewOrderLifecycleEvent(
			savedOrder.ID,
			savedOrder.CheckoutSessionID,
			savedOrder.Status,
			savedOrder.CustomerID,
			savedOrder.TotalPrice,
			savedOrder.Currency,
			savedOrder.Items,
		)

		payload, err := sonic.Marshal(evt)
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

		// Step 11: Build response with order + payment metadata
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

// ExecutePostPaymentSagaAsync executes post-payment saga in background.
// This is called after payment succeeds to process fulfillment, deduct inventory, and send confirmation.
func (s *orderService) ExecutePostPaymentSagaAsync(
	ctx context.Context,
	orderID uuid.UUID,
) {
	go func() {
		// Create background context with user authentication for async saga execution
		bgCtx := echoutils.PropagateUserContextToBackground(ctx)

		// Copy trace ID if present
		if traceID := ctx.Value(constant.CtxTraceIDKey); traceID != nil {
			bgCtx = context.WithValue(bgCtx, constant.CtxTraceIDKey, traceID)
		}

		s.logger.Infof("Starting post-payment saga for order %s", orderID)

		// Load order
		orderRepo := s.dataStore.OrderRepository()

		order, err := orderRepo.FindByID(bgCtx, orderID)
		if err != nil {
			s.logger.Errorf("Failed to retrieve order for post-payment saga: %v", err)

			return
		}

		// Load saga state with metadata (reserved products, customer email, etc.)
		sagaRepo := s.dataStore.SagaStateRepository()

		sagaState, err := sagaRepo.FindByOrderIDAndWorkflow(
			bgCtx,
			orderID,
			constant.PostPaymentSagaWorkflowName,
		)
		if err != nil {
			if err.Error() == constant.SagaStateNotFoundErrorMessage {
				s.logger.Warnf("No post-payment saga state found for order %s", orderID)

				return
			}

			s.logger.Errorf("Failed to retrieve saga state: %v", err)

			return
		}

		// Convert saga data to metadata
		metadata := &saga.Metadata{}
		metadata.FromMap(sagaState.Data)

		payload := &saga.Payload{Order: order}

		// Execute post-payment saga
		if sagaErr := s.sagaOrchestrator.ExecutePostPaymentSaga(bgCtx, payload, metadata); sagaErr != nil {
			s.logger.Errorf("Post-payment saga failed for order %s: %v", orderID, sagaErr)
		} else {
			s.logger.Infof("Post-payment saga completed successfully for order %s", orderID)
		}
	}()
}
