// Package service provides business logic for payment operations.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/utils/redisutils"
)

// PaymentService defines the interface for payment business operations.
type PaymentService interface {
	// CreatePayment creates a new payment record from order information
	CreatePayment(
		ctx context.Context,
		req dto.CreatePaymentRequest,
	) (*dto.PaymentResponse, error)
	// CreatePaymentIntent creates a PaymentIntent with Stripe and stores payment record
	CreatePaymentIntent(
		ctx context.Context,
		req dto.CreatePaymentIntentRequest,
	) (*dto.CreatePaymentIntentResponse, error)
	// ProcessPayment processes a payment transaction
	ProcessPayment(
		ctx context.Context,
		orderID uuid.UUID,
		req dto.ProcessPaymentRequest,
	) (*dto.PaymentResponse, error)
	// TimeoutPayment times out a payment transaction
	TimeoutPayment(ctx context.Context, orderID uuid.UUID) error
	// GetPaymentByOrderID retrieves payment by order ID
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*dto.PaymentResponse, error)
}

// paymentService implements the PaymentService.
type paymentService struct {
	dataStore             repository.DataStore
	logger                logger.Logger
	paymentGatewayClients map[string]client.PaymentGatewayClient
}

// NewPaymentService creates a new instance of paymentService.
func NewPaymentService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	paymentGatewayClients map[string]client.PaymentGatewayClient,
) PaymentService {
	return &paymentService{
		dataStore:             dataStore,
		logger:                appLogger,
		paymentGatewayClients: paymentGatewayClients,
	}
}

// getGatewayClient retrieves the payment gateway client for the specified provider.
func (s *paymentService) getGatewayClient(
	provider constant.PaymentGateway,
) (client.PaymentGatewayClient, error) {
	gatewayClient, ok := s.paymentGatewayClients[string(provider)]
	if !ok {
		return nil, httperror.NewBadRequestError(
			fmt.Sprintf("unsupported payment gateway: %s", provider),
		)
	}

	return gatewayClient, nil
}

// CreatePayment creates a new payment record from order information.
func (s *paymentService) CreatePayment(
	ctx context.Context,
	req dto.CreatePaymentRequest,
) (*dto.PaymentResponse, error) {
	res := new(dto.PaymentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if payment already exists for this order
		existingPayment, err := paymentRepo.FindByOrderID(ctx, req.OrderID)
		if err != nil && err.Error() != constant.PaymentNotFoundErrorMessage {
			return httperror.NewInternalServerError("failed to check existing payment")
		}

		if existingPayment != nil {
			// Payment already exists, return existing payment
			res = mapper.MapToPaymentResponse(existingPayment)

			return nil
		}

		// Create new payment entity
		payment, err := entity.NewPayment(
			req.OrderID,
			req.Amount,
			req.Currency,
			req.PaymentGateway,
		)
		if err != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("failed to create payment: %v", err))
		}

		// Store payment method info if provided (for delayed payment confirmation)
		if req.PaymentMethodID != "" && req.StripeCustomerID != "" {
			s.logger.Infof(
				"Storing payment method info for order %s: PM=%s, Customer=%s",
				req.OrderID,
				req.PaymentMethodID,
				req.StripeCustomerID,
			)

			// Create Stripe metadata with payment method and customer IDs
			stripeMetadata := &entity.StripeMetadata{
				PaymentMethodID: &req.PaymentMethodID,
				CustomerID:      &req.StripeCustomerID,
			}

			if err = payment.SetStripeMetadata(stripeMetadata); err != nil {
				return httperror.NewBadRequestError("failed to set payment method info")
			}
		}

		// Save payment
		savedPayment, err := paymentRepo.Create(ctx, payment)
		if err != nil {
			return httperror.NewInternalServerError("failed to save payment")
		}

		// Publish payment created event
		evt := producer.NewPaymentLifecycleEvent(
			savedPayment.ID,
			savedPayment.OrderID,
			constant.PaymentStatusPending,
			savedPayment.Amount,
			nil,                    // clientSecret - not created yet for direct API call
			savedPayment.ExpiresAt, // 24-hour payment window expiry
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   savedPayment.ID,
			EventType:     kafka.PaymentCreatedEventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create outbox event")
		}

		res = mapper.MapToPaymentResponse(savedPayment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CreatePaymentIntent creates a PaymentIntent with Stripe and stores payment record
// This method combines database record creation with Stripe PaymentIntent creation.
//
//nolint:gocognit,gocyclo,cyclop,funlen // ignore complexity
func (s *paymentService) CreatePaymentIntent(
	ctx context.Context,
	req dto.CreatePaymentIntentRequest,
) (*dto.CreatePaymentIntentResponse, error) {
	s.logger.Infof(
		"Creating PaymentIntent - OrderID: %s, Amount: %s %s, Customer: %s",
		req.OrderID,
		req.Amount.String(),
		req.Currency,
		req.CustomerID,
	)

	res := new(dto.CreatePaymentIntentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if payment already exists for this order
		existingPayment, err := paymentRepo.FindByOrderID(ctx, req.OrderID)
		if err != nil && err.Error() != constant.PaymentNotFoundErrorMessage {
			return httperror.NewInternalServerError("failed to check existing payment")
		}

		//nolint:nestif // ignore complexity
		if existingPayment != nil {
			// Payment already exists, return existing payment details
			if existingPayment.GatewayTransactionID != nil {
				res.PaymentIntentID = *existingPayment.GatewayTransactionID
				res.PaymentGateway = string(existingPayment.PaymentGateway)
				res.Status = string(existingPayment.Status)
				res.Amount = existingPayment.Amount.String()
				res.Currency = existingPayment.Currency
				res.OrderID = existingPayment.OrderID.String()
				res.ExpiresAt = existingPayment.ExpiresAt

				// Extract ClientSecret from Stripe metadata
				if existingPayment.PaymentGateway == constant.PaymentGatewayStripe {
					stripeMetadata, metadataErr := existingPayment.GetStripeMetadata()
					if metadataErr == nil && stripeMetadata.ClientSecret != nil {
						res.ClientSecret = *stripeMetadata.ClientSecret
					}
				}
			}

			return nil
		}

		// Create new payment entity first
		payment, err := entity.NewPayment(
			req.OrderID,
			req.Amount,
			req.Currency,
			req.PaymentGateway,
		)
		if err != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("failed to create payment: %v", err))
		}

		// Create Stripe PaymentIntent via gateway client
		gatewayClient, err := s.getGatewayClient(req.PaymentGateway)
		if err != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to get payment gateway client: %v", err),
			)
		}

		// Build gateway request for Stripe
		gatewayReq := &dto.PaymentGatewayRequest{
			PaymentID:      payment.ID,
			CustomerID:     req.CustomerID,
			CustomerEmail:  req.CustomerEmail,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Description:    fmt.Sprintf("Order %s", req.OrderID),
			IdempotencyKey: req.IdempotencyKey.String(),
			ExpiresAt:      payment.ExpiresAt,
		}

		gatewayResp, err := gatewayClient.ProcessPayment(ctx, gatewayReq)
		if err != nil {
			return httperror.NewInternalServerError(
				fmt.Sprintf("failed to create Stripe PaymentIntent: %v", err),
			)
		}

		// Create Stripe metadata from gateway response
		stripeMetadata := &entity.StripeMetadata{
			PaymentIntentID: &gatewayResp.GatewayID,
			ClientSecret:    gatewayResp.ClientSecret,
		}

		// Store payment method info from gateway metadata if available
		if gatewayResp.GatewayResponse != nil {
			if paymentMethodID, ok := gatewayResp.GatewayResponse["payment_method_id"].(string); ok &&
				paymentMethodID != "" {
				stripeMetadata.PaymentMethodID = &paymentMethodID
			}

			if customerID, ok := gatewayResp.GatewayResponse["customer_id"].(string); ok &&
				customerID != "" {
				stripeMetadata.CustomerID = &customerID
			}
		}

		// Update payment with Stripe response details
		payment.GatewayTransactionID = &gatewayResp.GatewayID
		payment.Status = constant.PaymentStatus(gatewayResp.Status)

		// Store Stripe metadata in payment entity
		if err = payment.SetStripeMetadata(stripeMetadata); err != nil {
			return httperror.NewBadRequestError("failed to set payment metadata")
		}

		// Save payment with Stripe details
		savedPayment, err := paymentRepo.Create(ctx, payment)
		if err != nil {
			return httperror.NewInternalServerError("failed to save payment")
		}

		// Publish payment created event
		evt := producer.NewPaymentLifecycleEvent(
			savedPayment.ID,
			savedPayment.OrderID,
			constant.PaymentStatusPending,
			savedPayment.Amount,
			gatewayResp.ClientSecret,
			savedPayment.ExpiresAt, // 24-hour payment window expiry
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   savedPayment.ID,
			EventType:     kafka.PaymentCreatedEventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create outbox event")
		}

		// Build response
		res.PaymentIntentID = gatewayResp.GatewayID
		res.ClientSecret = *gatewayResp.ClientSecret
		res.PaymentGateway = string(req.PaymentGateway)
		res.Status = string(gatewayResp.Status)
		res.Amount = gatewayResp.Amount.String()
		res.Currency = gatewayResp.Currency
		res.OrderID = req.OrderID.String()
		res.ExpiresAt = savedPayment.ExpiresAt

		// Convert gateway metadata from map[string]any to map[string]string
		if gatewayResp.GatewayResponse != nil {
			res.GatewayMetadata = make(map[string]string)

			for k, v := range gatewayResp.GatewayResponse {
				if v != nil {
					res.GatewayMetadata[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

//nolint:gocyclo,revive,cyclop,nolintlint // ignore complexity, ProcessPayment processes a payment transaction is large but intentional.
func (s *paymentService) ProcessPayment(
	ctx context.Context,
	orderID uuid.UUID,
	req dto.ProcessPaymentRequest,
) (*dto.PaymentResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreatePaymentTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreatePaymentRetryInterval),
			constant.CreatePaymentRetryLimit,
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

	res := new(dto.PaymentResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Get payment
		payment, errFind := paymentRepo.FindByOrderID(ctx, orderID)
		if errFind != nil {
			if errFind.Error() == constant.PaymentNotFoundErrorMessage {
				return httperror.NewNotFoundError("payment not found for order")
			}

			return httperror.NewInternalServerError("failed to get payment")
		}

		if payment == nil {
			return httperror.NewPaymentNotFoundError()
		}

		// Check if payment can be processed
		if !payment.CanBeProcessed() {
			return httperror.NewBadRequestError("payment cannot be processed in current status")
		}

		// Update status to processing
		if err = payment.UpdateStatus(constant.PaymentStatusProcessing); err != nil {
			return httperror.NewBadRequestError("failed to update payment status")
		}

		// Process payment with payment gateway
		paymentResult, errProcess := s.processWithPaymentGateway(ctx, payment, req)
		if errProcess != nil {
			// If gateway call fails, mark payment as failed
			if err = payment.UpdateStatus(constant.PaymentStatusFailed); err != nil {
				s.logger.Errorf("failed to update payment status to failed: %v", err)
			}

			return httperror.NewInternalServerError("failed to process payment with gateway")
		}

		// Set gateway reference
		if err = payment.SetGatewayReference(
			payment.PaymentGateway,
			paymentResult.GatewayID,
			paymentResult.GatewayResponse,
		); err != nil {
			return httperror.NewInternalServerError("failed to set gateway reference")
		}

		// For client-side confirmation flow (Stripe.js):
		// - Payment stays in PROCESSING status
		// - Client receives ClientSecret to complete payment
		// - Webhooks will update to COMPLETED/FAILED after client confirmation
		//
		// Exception: If gateway immediately returns succeeded (rare), complete now
		if paymentResult.Status == constant.PaymentGatewayStatusSucceeded {
			if err = payment.UpdateStatus(constant.PaymentStatusCompleted); err != nil {
				return httperror.NewBadRequestError("failed to update payment status to completed")
			}
		}

		// Save updated payment
		updatedPayment, errUpdate := paymentRepo.Update(ctx, payment)
		if errUpdate != nil {
			return httperror.NewInternalServerError("failed to update payment")
		}

		// Determine event type based on status
		var eventType string
		if updatedPayment.Status == constant.PaymentStatusCompleted {
			eventType = kafka.PaymentCompletedEventType
		} else {
			eventType = kafka.PaymentProcessingEventType
		}

		// Publish payment event
		evt := producer.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			updatedPayment.OrderID,
			updatedPayment.Status,
			updatedPayment.Amount,
			nil, // clientSecret not needed for processing event
			nil, // expiresAt not needed for processing event
		)

		payload, errMarshal := json.Marshal(evt)
		if errMarshal != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   updatedPayment.ID,
			EventType:     eventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create payment event")
		}

		// Build response with gateway data for client-side confirmation
		res = mapper.MapToPaymentResponse(updatedPayment)
		res.ClientSecret = paymentResult.ClientSecret
		res.RequiresAction = paymentResult.RequiresAction

		if paymentResult.NextAction != nil {
			actionType := string(paymentResult.NextAction.Type)
			res.NextActionType = &actionType
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetPaymentByOrderID retrieves payment by order ID.
func (s *paymentService) GetPaymentByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*dto.PaymentResponse, error) {
	paymentRepo := s.dataStore.PaymentRepository()

	payment, err := paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, constant.ErrPaymentNotFound) {
			return nil, httperror.NewNotFoundError("payment not found for order")
		}

		return nil, httperror.NewInternalServerError("failed to get payment")
	}

	if payment == nil {
		return nil, httperror.NewPaymentNotFoundError()
	}

	return mapper.MapToPaymentResponse(payment), nil
}

// processWithPaymentGateway processes payment with payment gateway client.
func (s *paymentService) processWithPaymentGateway(
	ctx context.Context,
	payment *entity.Payment,
	req dto.ProcessPaymentRequest,
) (*dto.PaymentGatewayResponse, error) {
	s.logger.Infof(
		"Processing payment %s with gateway %s, method %s and amount %s",
		payment.ID,
		payment.PaymentGateway,
		payment.Amount,
	)

	// Get the appropriate gateway client based on payment's gateway
	gatewayClient, err := s.getGatewayClient(payment.PaymentGateway)
	if err != nil {
		return nil, err
	}

	paymentRequest := &dto.PaymentGatewayRequest{
		PaymentID:       payment.ID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		PaymentMethodID: req.PaymentMethodID, // Stripe PM ID from client
		CustomerID:      req.CustomerID,
		CustomerEmail:   req.CustomerEmail,
		IdempotencyKey:  req.IdempotencyKey.String(),
		ExpiresAt:       payment.ExpiresAt, // 24-hour payment window expiry
	}

	return gatewayClient.ProcessPayment(ctx, paymentRequest)
}

// TimeoutPayment times out a payment transaction.
func (s *paymentService) TimeoutPayment(ctx context.Context, orderID uuid.UUID) error {
	s.logger.Infof("Processing payment timeout for order: %s", orderID)

	return s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		payment, err := paymentRepo.FindByOrderID(ctx, orderID)
		if err != nil {
			if errors.Is(err, constant.ErrPaymentNotFound) {
				s.logger.Warnf("Payment not found for order %s during timeout", orderID)
				return nil // Don't error if payment doesn't exist
			}

			return httperror.NewInternalServerError("failed to get payment")
		}

		if payment == nil {
			s.logger.Warnf("Payment not found for order %s during timeout", orderID)
			return nil // Don't error if payment doesn't exist
		}

		// Check if payment can be timed out (only pending payments)
		if payment.Status != constant.PaymentStatusPending {
			s.logger.Infof(
				"Payment for order %s is already in status %s, skipping timeout",
				orderID,
				payment.Status,
			)

			return nil
		}

		// Update payment status to timed out
		if err = payment.UpdateStatus(constant.PaymentStatusTimeout); err != nil {
			return httperror.NewInternalServerError("failed to update payment status to timeout")
		}

		// Save updated payment
		updatedPayment, err := paymentRepo.Update(ctx, payment)
		if err != nil {
			return httperror.NewInternalServerError("failed to update payment")
		}

		// Publish payment timeout event using outbox pattern
		evt := producer.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			updatedPayment.OrderID,
			constant.PaymentStatusTimeout,
			updatedPayment.Amount,
			nil, // clientSecret not needed for timeout event
			nil, // expiresAt not needed for timeout event
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment timeout event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   updatedPayment.ID,
			EventType:     kafka.PaymentTimeoutEventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create payment timeout event")
		}

		s.logger.Infof("Successfully timed out payment for order %s", orderID)

		return nil
	})
}
