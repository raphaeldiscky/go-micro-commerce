// Package service provides business logic for checkout session operations.
package service

import (
	"context"
	"fmt"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/client"
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
	CancelCheckoutSession(
		ctx context.Context,
		sessionID uuid.UUID,
		customerID uuid.UUID,
	) (*dto.CheckoutSessionResponse, error)
}

// checkoutSessionService implements the CheckoutSessionService.
type checkoutSessionService struct {
	dataStore                          repository.DataStore
	logger                             logger.Logger
	productClient                      client.ProductClient
	checkoutSessionOrderPlacedProducer kafka.Producer
}

// NewCheckoutSessionService creates a new instance of checkoutSessionService.
func NewCheckoutSessionService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	productClient client.ProductClient,
	checkoutSessionOrderPlacedProducer kafka.Producer,
) CheckoutSessionService {
	return &checkoutSessionService{
		dataStore:                          dataStore,
		logger:                             appLogger,
		productClient:                      productClient,
		checkoutSessionOrderPlacedProducer: checkoutSessionOrderPlacedProducer,
	}
}

// CreateCheckoutSession creates a new checkout session from a cart.
//
//nolint:gocyclo,cyclop // ignore for now
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
		cart, errCart := cartRepo.FindActiveCartByUserIDForCheckout(ctx, req.CustomerID)
		if errCart != nil {
			return httperror.NewInternalServerError("failed to get cart")
		}

		if cart == nil {
			return httperror.NewBadRequestError("cart not found")
		}

		// Validate cart can checkout (must be active and have selected items)
		if err = cart.CanCheckout(); err != nil {
			return httperror.NewBadRequestError(err.Error())
		}

		// Mark cart as checked out
		if err = cart.MarkAsCheckedOut(); err != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to mark cart as checked out: %v", err),
			)
		}

		// Update cart status in database
		if err = cartRepo.UpdateStatus(ctx, cart.ID, constant.CartStatusCheckedOut); err != nil {
			s.logger.Errorf("failed to update cart status: %v", err)
			return httperror.NewInternalServerError("failed to update cart status")
		}

		productIDs := make([]uuid.UUID, len(cart.Items))
		for i, item := range cart.Items {
			productIDs[i] = item.ProductID
		}

		// Fetch products
		products, errProducts := s.productClient.GetProducts(ctx, productIDs)
		if errProducts != nil {
			s.logger.Errorf("failed to get products: %v", errProducts)
			return httperror.NewInternalServerError("failed to get products")
		}

		// Build new checkout session items with prices from product-service
		// Create a map of products for quick lookup
		productMap := make(map[uuid.UUID]*entity.Product, len(products))
		for i := range products {
			productMap[products[i].ID] = &products[i]
		}

		// Loop through cart items and create checkout session items
		checkoutItems := make([]entity.CheckoutSessionItem, 0, len(cart.Items))
		for _, cartItem := range cart.Items {
			// Find product in the fetched products
			product, exists := productMap[cartItem.ProductID]
			if !exists {
				return httperror.NewBadRequestError(
					fmt.Sprintf("product %s not found", cartItem.ProductID),
				)
			}

			// Create checkout session item with product price
			checkoutItem, errItem := entity.NewCheckoutSessionItem(
				cartItem.ProductID,
				product.Name,
				cartItem.Quantity,
				product.UnitPrice,
			)
			if errItem != nil {
				return httperror.NewBadRequestError(
					fmt.Sprintf("failed to create checkout item: %v", errItem),
				)
			}

			checkoutItems = append(checkoutItems, *checkoutItem)
		}

		// Create checkout session entity with simplified signature
		session, errSession := entity.NewCheckoutSession(
			req.IdempotencyKey,
			req.CustomerID,
			req.CartID,
			"IDR", // Default currency
			checkoutItems,
		)
		if errSession != nil {
			s.logger.Errorf("failed to create checkout session entity: %v", err)

			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to create checkout session entity: %v", errSession),
			)
		}

		// Save checkout session
		createdSession, err = checkoutSessionRepo.Create(ctx, session)
		if err != nil {
			s.logger.Errorf("failed to create checkout session: %v", err)
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
		cartRepo := ds.CartRepository()

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

		// Validate products before placing order (price & stock check)
		if errValidate := s.productClient.ValidateProducts(ctx, session.Items); errValidate != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("product validation failed: %v", errValidate),
			)
		}

		// Update session with order details
		session.AddressID = &req.AddressID
		session.CarrierID = &req.CarrierID
		session.PaymentGateway = &req.PaymentGateway

		// Update status to order_placed
		if err = session.UpdateStatus(constant.CheckoutSessionStatusOrderPlaced); err != nil {
			return httperror.NewBadRequestError("failed to update checkout session status")
		}

		// Save updated session
		updatedSession, err = checkoutSessionRepo.Update(ctx, session)
		if err != nil {
			return httperror.NewInternalServerError("failed to update checkout session")
		}

		// Migrate cart: archive checked out cart and create new active cart with unselected items
		s.migrateCartAfterOrderPlacement(ctx, cartRepo, req.CustomerID)

		// Dereference nullable fields for event
		paymentGateway := ""
		if updatedSession.PaymentGateway != nil {
			paymentGateway = *updatedSession.PaymentGateway
		}

		// Publish domain event via outbox pattern
		evt := producer.NewCheckoutSessionOrderPlacedEvent(
			updatedSession.ID,
			req.IdempotencyKey,
			constant.CheckoutSessionStatusOrderPlaced,
			updatedSession.CustomerID,
			updatedSession.Currency,
			paymentGateway,
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

// migrateCartAfterOrderPlacement archives the checked out cart and creates a new active cart with unselected items.
func (s *checkoutSessionService) migrateCartAfterOrderPlacement(
	ctx context.Context,
	cartRepo repository.CartRepository,
	customerID uuid.UUID,
) {
	// Get checked out cart with unselected items for migration
	cart, errCart := cartRepo.FindCheckedOutCartWithUnselectedItems(ctx, customerID)
	if errCart != nil {
		s.logger.Warnf("failed to get cart for migration: %v", errCart)
		return
	}

	if cart == nil {
		return
	}

	// Archive the old cart
	if errArchive := cart.MarkAsArchived(); errArchive != nil {
		s.logger.Warnf("failed to mark cart as archived: %v", errArchive)
		return
	}

	if errUpdateStatus := cartRepo.UpdateStatus(ctx, cart.ID, constant.CartStatusArchived); errUpdateStatus != nil {
		s.logger.Warnf("failed to update cart status to archived: %v", errUpdateStatus)
		return
	}

	// Create new active cart only if there are unselected items
	if len(cart.Items) == 0 {
		return
	}

	newCart, errNewCart := entity.NewCart(customerID, cart.Items)
	if errNewCart != nil {
		s.logger.Warnf("failed to create new cart entity: %v", errNewCart)
		return
	}

	if _, errCreate := cartRepo.Create(ctx, newCart); errCreate != nil {
		s.logger.Warnf("failed to save new cart: %v", errCreate)
	}
}

// CancelCheckoutSession cancels a checkout session and reverts the cart to active.
func (s *checkoutSessionService) CancelCheckoutSession(
	ctx context.Context,
	sessionID uuid.UUID,
	customerID uuid.UUID,
) (*dto.CheckoutSessionResponse, error) {
	var (
		updatedSession *entity.CheckoutSession
		err            error
	)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		checkoutSessionRepo := ds.CheckoutSessionRepository()
		cartRepo := ds.CartRepository()

		// Get checkout session
		session, errSession := checkoutSessionRepo.GetByID(ctx, sessionID)
		if errSession != nil {
			return httperror.NewInternalServerError("failed to get checkout session")
		}

		if session == nil {
			return httperror.NewBadRequestError("checkout session not found")
		}

		// Validate session belongs to the customer
		if session.CustomerID != customerID {
			return httperror.NewForbiddenError("checkout session does not belong to customer")
		}

		// Validate session can be canceled
		if !session.CanBeCanceled() {
			return httperror.NewBadRequestError(
				fmt.Sprintf(
					"checkout session cannot be canceled in current status: %s",
					session.Status,
				),
			)
		}

		// Update status to canceled
		if err = session.UpdateStatus(constant.CheckoutSessionStatusCanceled); err != nil {
			return httperror.NewBadRequestError("failed to update checkout session status")
		}

		// Save updated session
		updatedSession, err = checkoutSessionRepo.Update(ctx, session)
		if err != nil {
			return httperror.NewInternalServerError("failed to update checkout session")
		}

		// Get customer's checked out cart to revert to active
		cart, errCart := cartRepo.FindByUserID(ctx, customerID)
		if errCart != nil {
			s.logger.Warnf("failed to get cart for reverting: %v", errCart)
		} else if cart != nil && cart.Status == constant.CartStatusCheckedOut {
			// Revert cart to active using domain method
			if errRevert := cart.RevertToActive(); errRevert != nil {
				s.logger.Warnf("failed to revert cart to active: %v", errRevert)
			} else if errUpdate := cartRepo.UpdateStatus(ctx, cart.ID, constant.CartStatusActive); errUpdate != nil {
				s.logger.Warnf("failed to update cart status to active: %v", errUpdate)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCheckoutSessionResponse(updatedSession), nil
}
