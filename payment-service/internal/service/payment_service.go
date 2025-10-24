// Package service provides business logic for payment operations.
package service

import (
	"context"
	"encoding/json"
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
	// CreateSetupIntent creates a SetupIntent for collecting payment method without charging
	CreateSetupIntent(
		ctx context.Context,
		req dto.CreateSetupIntentRequest,
	) (*dto.SetupIntentResponse, error)
	// ChargeStoredPaymentMethod charges a saved payment method (delayed payment flow)
	ChargeStoredPaymentMethod(
		ctx context.Context,
		orderID uuid.UUID,
	) (*dto.PaymentResponse, error)
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
		if err.Error() == constant.PaymentNotFoundErrorMessage {
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
		TransactionID:   payment.ID,
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
			if err.Error() == constant.PaymentNotFoundErrorMessage {
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

// CreateSetupIntent creates a SetupIntent for collecting payment method without charging.
// Used for delayed payment confirmation pattern.
func (s *paymentService) CreateSetupIntent(
	ctx context.Context,
	req dto.CreateSetupIntentRequest,
) (*dto.SetupIntentResponse, error) {
	s.logger.Infof(
		"Creating SetupIntent for order: %s, customer: %s",
		req.OrderID,
		req.CustomerID,
	)

	// Get Stripe gateway client (SetupIntent only supported by Stripe)
	gatewayClient, err := s.getGatewayClient(constant.PaymentGatewayStripe)
	if err != nil {
		return nil, err
	}

	// Create SetupIntent request
	setupIntentReq := &dto.SetupIntentRequest{
		CustomerID:    req.CustomerID,
		CustomerEmail: req.CustomerEmail,
		OrderID:       req.OrderID,
	}

	// Call gateway to create SetupIntent
	response, err := gatewayClient.CreateSetupIntent(ctx, setupIntentReq)
	if err != nil {
		s.logger.Errorf("Failed to create SetupIntent: %v", err)
		return nil, httperror.NewInternalServerError("failed to create setup intent")
	}

	s.logger.Infof(
		"SetupIntent created successfully: %s for order: %s",
		response.SetupIntentID,
		req.OrderID,
	)

	return response, nil
}

// ChargeStoredPaymentMethod charges a saved payment method without customer present.
// Used for delayed payment confirmation when order status changes.
//
//nolint:gocyclo,revive,cyclop,nolintlint,gocognit,funlen // ignore complexity, ChargeStoredPaymentMethod is large but intentional.
func (s *paymentService) ChargeStoredPaymentMethod(
	ctx context.Context,
	orderID uuid.UUID,
) (*dto.PaymentResponse, error) {
	s.logger.Infof("Charging stored payment method for order: %s", orderID)

	res := new(dto.PaymentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// 1. Get payment record
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

		// 2. Validate payment can be charged
		if payment.Status != constant.PaymentStatusPending {
			s.logger.Warnf(
				"Payment for order %s is in status %s, cannot charge",
				orderID,
				payment.Status,
			)

			return httperror.NewBadRequestError(
				fmt.Sprintf("payment already processed (status: %s)", payment.Status),
			)
		}

		// 3. Validate payment method info is stored
		stripeMetadata, err := payment.GetStripeMetadata()
		if err != nil {
			s.logger.Errorf("Failed to get Stripe metadata for order %s: %v", orderID, err)
			return httperror.NewInternalServerError("failed to retrieve payment metadata")
		}

		if stripeMetadata.PaymentMethodID == nil || *stripeMetadata.PaymentMethodID == "" {
			s.logger.Errorf("Payment method ID not found for order: %s", orderID)
			return httperror.NewBadRequestError("payment method not saved")
		}

		if stripeMetadata.CustomerID == nil || *stripeMetadata.CustomerID == "" {
			s.logger.Errorf("Stripe customer ID not found for order: %s", orderID)
			return httperror.NewBadRequestError("stripe customer not saved")
		}

		// 4. Update status to processing
		if err = payment.UpdateStatus(constant.PaymentStatusProcessing); err != nil {
			return httperror.NewBadRequestError("failed to update payment status")
		}

		// 5. Get gateway client
		gatewayClient, errGateway := s.getGatewayClient(payment.PaymentGateway)
		if errGateway != nil {
			return errGateway
		}

		// 6. Charge off-session
		chargeReq := &dto.ChargeOffSessionRequest{
			PaymentMethodID:  *stripeMetadata.PaymentMethodID,
			StripeCustomerID: *stripeMetadata.CustomerID,
			Amount:           payment.Amount,
			Currency:         payment.Currency,
			TransactionID:    payment.ID,
			OrderID:          orderID,
			Description:      fmt.Sprintf("Order %s", orderID),
		}

		result, errCharge := gatewayClient.ChargeOffSession(ctx, chargeReq)

		//nolint:nestif // ignore complexity
		if errCharge != nil {
			s.logger.Errorf("Off-session charge failed for order %s: %v", orderID, errCharge)

			// Mark payment as failed
			if err = payment.UpdateStatus(constant.PaymentStatusFailed); err != nil {
				s.logger.Errorf("Failed to update payment status to failed: %v", err)
			}

			_, err = paymentRepo.Update(ctx, payment)
			if err != nil {
				s.logger.Errorf("Failed to update payment: %v", err)
				return err
			}

			// Publish failed event
			evt := producer.NewPaymentLifecycleEvent(
				payment.ID,
				payment.OrderID,
				constant.PaymentStatusFailed,
				payment.Amount,
				nil, // clientSecret not needed for failed event
				nil, // expiresAt not needed for failed event
			)

			payload, errMarshal := json.Marshal(evt)
			if errMarshal != nil {
				return errMarshal
			}

			outboxEvent := &entity.OutboxEvent{
				ID:            uuid.New(),
				AggregateType: "payment",
				AggregateID:   payment.ID,
				EventType:     kafka.PaymentFailedEventType,
				Topic:         kafka.PaymentLifecycleTopic,
				Payload:       payload,
				Status:        constant.OutboxStatusPending,
				CreatedAt:     time.Now().UTC(),
				ScheduledFor:  time.Now().UTC(),
				Attempts:      0,
			}

			err = outboxRepo.Create(ctx, outboxEvent)
			if err != nil {
				return err
			}

			return httperror.NewInternalServerError("charge failed")
		}

		// 7. Update payment with gateway info
		if err = payment.SetGatewayReference(
			payment.PaymentGateway,
			result.GatewayID,
			result.GatewayResponse,
		); err != nil {
			return httperror.NewInternalServerError("failed to set gateway reference")
		}

		// 8. Determine final status based on gateway response
		var finalStatus constant.PaymentStatus
		if result.Status == constant.PaymentGatewayStatusSucceeded {
			finalStatus = constant.PaymentStatusCompleted
		} else {
			// Off-session charges should succeed or fail immediately
			// If still pending, we'll rely on webhooks to update
			finalStatus = constant.PaymentStatusProcessing
		}

		if err = payment.UpdateStatus(finalStatus); err != nil {
			return httperror.NewBadRequestError("failed to update final payment status")
		}

		// 9. Save updated payment
		updatedPayment, errUpdate := paymentRepo.Update(ctx, payment)
		if errUpdate != nil {
			return httperror.NewInternalServerError("failed to update payment")
		}

		// 10. Determine event type
		var eventType string

		switch finalStatus {
		case constant.PaymentStatusCompleted:
			eventType = kafka.PaymentCompletedEventType
		case constant.PaymentStatusFailed:
			eventType = kafka.PaymentFailedEventType
		case constant.PaymentStatusPending,
			constant.PaymentStatusProcessing,
			constant.PaymentStatusTimeout,
			constant.PaymentStatusRefunded:
			eventType = kafka.PaymentProcessingEventType
		default:
			eventType = kafka.PaymentProcessingEventType
		}

		// 11. Publish payment event
		evt := producer.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			updatedPayment.OrderID,
			updatedPayment.Status,
			updatedPayment.Amount,
			nil, // clientSecret not needed for charge event
			nil, // expiresAt not needed for charge event
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

		res = mapper.MapToPaymentResponse(updatedPayment)

		s.logger.Infof(
			"Payment charged successfully: %s, order: %s, status: %s",
			updatedPayment.ID,
			orderID,
			updatedPayment.Status,
		)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
