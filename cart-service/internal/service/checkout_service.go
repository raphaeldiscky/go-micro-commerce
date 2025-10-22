// Package service provides business logic for checkout session operations.
package service

import (
	"context"
	"fmt"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/utils/redisutils"
)

// CheckoutSessionService defines the interface for checkout session business operations.
type CheckoutSessionService interface {
	CreateCheckoutSession(
		ctx context.Context,
		req *dto.CreateCheckoutSessionRequest,
	) (*dto.CheckoutSessionResponse, error)
	GetCheckoutSession(ctx context.Context, id uuid.UUID) (*dto.CheckoutSessionResponse, error)
	PlaceOrder(
		ctx context.Context,
		sessionID uuid.UUID,
		req *dto.PlaceOrderRequest,
	) (*dto.CheckoutSessionResponse, error)
}

// checkoutSessionService implements the CheckoutSessionService.
type checkoutSessionService struct {
	dataStore                          repository.DataStore
	logger                             logger.Logger
	checkoutSessionOrderPlacedProducer kafka.Producer
}

// NewCheckoutSessionService creates a new instance of checkoutSessionService.
func NewCheckoutSessionService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	checkoutSessionOrderPlacedProducer kafka.Producer,
) CheckoutSessionService {
	return &checkoutSessionService{
		dataStore:                          dataStore,
		logger:                             appLogger,
		checkoutSessionOrderPlacedProducer: checkoutSessionOrderPlacedProducer,
	}
}

// CreateCheckoutSession creates a new checkout session from a cart.
func (s *checkoutSessionService) CreateCheckoutSession(
	ctx context.Context,
	req *dto.CreateCheckoutSessionRequest,
) (*dto.CheckoutSessionResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateCheckoutSessionTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateCheckoutSessionRetryInterval),
			constant.CreateCheckoutSessionRetryLimit,
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

	var createdSession *entity.CheckoutSession

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		checkoutSessionRepo := ds.CheckoutSessionRepository()
		cartRepo := ds.CartRepository()

		// Get customer's cart with ONLY selected items (database-level filtering)
		cart, errCart := cartRepo.FindByUserIDForCheckout(ctx, req.CustomerID)
		if errCart != nil {
			return httperror.NewInternalServerError("failed to get cart")
		}

		if cart == nil {
			return httperror.NewBadRequestError("cart not found")
		}

		// Validate cart has selected items (already filtered by DB query)
		if len(cart.Items) == 0 {
			return httperror.NewBadRequestError("no items selected for checkout")
		}

		// Copy selected cart items to checkout session items (snapshot pattern)
		var sessionItems []entity.CheckoutSessionItem

		for i := range cart.Items {
			cartItem := &cart.Items[i]

			sessionItem, errItem := entity.NewCheckoutSessionItem(
				cartItem.ProductID,
				cartItem.Quantity,
			)
			if errItem != nil {
				return httperror.NewBadRequestError(fmt.Sprintf("invalid item: %v", errItem))
			}

			sessionItems = append(sessionItems, *sessionItem)
		}

		// Create checkout session entity (snapshot pattern - no cart reference)
		session, errSession := entity.NewCheckoutSession(
			req.IdempotencyKey,
			req.CustomerID,
			req.AddressID,
			req.CarrierID,
			req.PaymentGateway,
			req.PaymentMethod,
			req.Currency,
			sessionItems,
		)
		if errSession != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to create checkout session: %v", errSession),
			)
		}

		// Save checkout session
		createdSession, err = checkoutSessionRepo.Create(ctx, session)
		if err != nil {
			return httperror.NewInternalServerError("failed to create checkout session")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCheckoutSessionResponse(createdSession), nil
}

// GetCheckoutSession retrieves a checkout session by ID.
func (s *checkoutSessionService) GetCheckoutSession(
	ctx context.Context,
	id uuid.UUID,
) (*dto.CheckoutSessionResponse, error) {
	checkoutSessionRepo := s.dataStore.CheckoutSessionRepository()

	session, err := checkoutSessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get checkout session")
	}

	if session == nil {
		return nil, httperror.NewBadRequestError("checkout session not found")
	}

	return mapper.MapToCheckoutSessionResponse(session), nil
}

// PlaceOrder places an order from a checkout session.
func (s *checkoutSessionService) PlaceOrder(
	ctx context.Context,
	sessionID uuid.UUID,
	req *dto.PlaceOrderRequest,
) (*dto.CheckoutSessionResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.PlaceOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.PlaceOrderRetryInterval),
			constant.PlaceOrderRetryLimit,
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

	var updatedSession *entity.CheckoutSession

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		checkoutSessionRepo := ds.CheckoutSessionRepository()

		// Get checkout session
		session, errSession := checkoutSessionRepo.GetByID(ctx, sessionID)
		if errSession != nil {
			return httperror.NewInternalServerError("failed to get checkout session")
		}

		if session == nil {
			return httperror.NewBadRequestError("checkout session not found")
		}

		// Validate session belongs to the customer
		if session.CustomerID != req.CustomerID {
			return httperror.NewForbiddenError("checkout session does not belong to customer")
		}

		// Validate session can place order
		if !session.CanPlaceOrder() {
			return httperror.NewBadRequestError(
				fmt.Sprintf(
					"checkout session cannot place order in current status: %s",
					session.Status,
				),
			)
		}

		// Update status to order_placed
		if err = session.UpdateStatus(constant.CheckoutSessionStatusOrderPlaced); err != nil {
			return httperror.NewBadRequestError("failed to update checkout session status")
		}

		// Save updated session
		updatedSession, err = checkoutSessionRepo.Update(ctx, session)
		if err != nil {
			return httperror.NewInternalServerError("failed to update checkout session")
		}

		// Publish domain event via outbox pattern
		evt := producer.NewCheckoutSessionOrderPlacedEvent(
			updatedSession.ID,
			req.IdempotencyKey,
			constant.CheckoutSessionStatusOrderPlaced,
			updatedSession.CustomerID,
			updatedSession.Currency,
			updatedSession.PaymentGateway,
			updatedSession.PaymentMethod,
			updatedSession.Items,
			updatedSession.CreatedAt,
		)

		if err = s.checkoutSessionOrderPlacedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError(
				"failed to send checkout session order placed event",
			)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCheckoutSessionResponse(updatedSession), nil
}
