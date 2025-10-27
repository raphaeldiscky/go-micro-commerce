// Package service provides business logic for checkout session operations.
package service

import (
	"context"
	"fmt"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/task"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/utils/redisutils"
)

// CheckoutSessionService defines the interface for checkout session business operations.
type CheckoutSessionService interface {
	CreateCheckoutSession(
		ctx context.Context,
		req *dto.CreateCheckoutSessionRequest,
	) (*dto.CheckoutSessionResponse, error)
	GetCheckoutSession(ctx context.Context, id uuid.UUID) (*dto.CheckoutSessionResponse, error)
	UpdateCheckoutSession(
		ctx context.Context,
		sessionID uuid.UUID,
		req *dto.UpdateCheckoutSessionRequest,
	) (*dto.CheckoutSessionResponse, error)
	PlaceOrder(
		ctx context.Context,
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
	dataStore               repository.DataStore
	logger                  logger.Logger
	productClient           client.ProductClient
	asynqClient             asynq.Client
	taskCancellationService asynq.TaskCancellationService
}

// NewCheckoutSessionService creates a new instance of checkoutSessionService.
func NewCheckoutSessionService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	productClient client.ProductClient,
	asynqClient asynq.Client,
	taskCancellationService asynq.TaskCancellationService,
) CheckoutSessionService {
	return &checkoutSessionService{
		dataStore:               dataStore,
		logger:                  appLogger,
		productClient:           productClient,
		asynqClient:             asynqClient,
		taskCancellationService: taskCancellationService,
	}
}

const (
	mockWeightKG = 0.1
	mockHeightCM = 10
	mockLengthCM = 10
	mockWidthCM  = 10
)

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

		// Create checkout session entity with empty shipping data (to be filled later)
		session, errSession := entity.NewCheckoutSession(
			req.IdempotencyKey,
			req.CustomerID,
			req.CartID,
			"IDR",                // @TODO: Mock currency first
			entity.Courier{},     // will be added on checkout page
			entity.Destination{}, // will be added on checkout page
			entity.Package{ // @TODO: Mock package first
				WeightKG: decimal.NewFromFloat(mockWeightKG),
				Length:   decimal.NewFromFloat(mockLengthCM),
				Width:    decimal.NewFromFloat(mockWidthCM),
				Height:   decimal.NewFromFloat(mockHeightCM),
				Unit:     "cm",
			},
			entity.Origin{ // @TODO: Mock origin first
				State:       "Central Java",
				City:        "Muntilan",
				PostalCode:  "45212",
				CountryCode: "ID",
			},
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

	// Schedule checkout session reminder task
	if err = s.scheduleCheckoutSessionReminder(ctx, createdSession.ID); err != nil {
		// Log error but don't fail the checkout session creation
		s.logger.Warnf(
			"Failed to schedule checkout session reminder for session %s: %v",
			createdSession.ID,
			err,
		)
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

// UpdateCheckoutSession updates a checkout session with address, carrier, or payment gateway.
func (s *checkoutSessionService) UpdateCheckoutSession(
	ctx context.Context,
	sessionID uuid.UUID,
	req *dto.UpdateCheckoutSessionRequest,
) (*dto.CheckoutSessionResponse, error) {
	var updatedSession *entity.CheckoutSession

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
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

		// Validate session can be updated (must be in pending status)
		if session.Status != constant.CheckoutSessionStatusPending {
			return httperror.NewBadRequestError("checkout session cannot be updated")
		}

		// Update only provided fields
		if req.Courier != nil {
			session.Courier = entity.Courier{
				CourierID: req.Courier.CourierID,
			}
		}

		if req.Destination != nil {
			session.Destination = entity.Destination{
				City:        req.Destination.City,
				State:       req.Destination.State,
				PostalCode:  req.Destination.PostalCode,
				CountryCode: req.Destination.CountryCode,
			}
		}

		if req.Origin != nil {
			session.Origin = entity.Origin{
				City:        req.Origin.City,
				State:       req.Origin.State,
				PostalCode:  req.Origin.PostalCode,
				CountryCode: req.Origin.CountryCode,
			}
		}

		if req.Package != nil {
			session.Package = entity.Package{
				WeightKG: req.Package.WeightKG,
				Width:    req.Package.Width,
				Height:   req.Package.Height,
				Length:   req.Package.Length,
				Unit:     req.Package.Unit,
			}
		}

		if req.PaymentGateway != nil {
			session.PaymentGateway = req.PaymentGateway
		}

		updated, errUpdate := checkoutSessionRepo.Update(ctx, session)
		if errUpdate != nil {
			return httperror.NewInternalServerError("failed to update checkout session")
		}

		updatedSession = updated

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCheckoutSessionResponse(updatedSession), nil
}

// PlaceOrder places an order from a checkout session.
func (s *checkoutSessionService) PlaceOrder(
	ctx context.Context,
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
		session, errSession := checkoutSessionRepo.GetByID(ctx, req.CheckoutSessionID)
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

		// Update status to order_placed
		if err = session.UpdateStatus(constant.CheckoutSessionStatusOrderPlaced); err != nil {
			return httperror.NewBadRequestError("failed to update checkout session status")
		}

		// Save updated session
		updatedSession, err = checkoutSessionRepo.Update(ctx, session)
		if err != nil {
			return httperror.NewInternalServerError("failed to update checkout session")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Cancel checkout session reminder task since order was placed
	if err = s.cancelCheckoutSessionReminder(ctx, req.CheckoutSessionID); err != nil {
		// Log error but don't fail the order placement
		s.logger.Warnf(
			"Failed to cancel checkout session reminder for session %s: %v",
			req.CheckoutSessionID,
			err,
		)
	}

	// Migrate cart after successful order placement (non-critical operation)
	// This runs outside the transaction so failures won't affect order placement
	cartRepo := s.dataStore.CartRepository()
	s.migrateCartAfterOrderPlacement(ctx, cartRepo, req.CustomerID)

	return mapper.MapToCheckoutSessionResponse(updatedSession), nil
}

// scheduleCheckoutSessionReminder schedules a reminder task for a checkout session.
func (s *checkoutSessionService) scheduleCheckoutSessionReminder(
	ctx context.Context,
	checkoutSessionID uuid.UUID,
) error {
	user, err := echoutils.GetUserAuthContexts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user auth contexts: %w", err)
	}
	// Create reminder task request
	reminderRequest := &dto.CheckoutSessionReminderRequest{
		CheckoutSessionID: checkoutSessionID,
		CustomerEmail:     user.Email,
	}

	// Create the task
	reminderTask, err := task.NewCheckoutSessionReminderTask(reminderRequest)
	if err != nil {
		return fmt.Errorf("failed to create reminder task: %w", err)
	}

	// Enqueue the task with a delay
	_, err = s.asynqClient.EnqueueIn(
		constant.CheckoutSessionReminderDelay,
		reminderTask,
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue reminder task: %w", err)
	}

	s.logger.Infof(
		"Scheduled checkout session reminder for session %s in %v",
		checkoutSessionID,
		constant.CheckoutSessionReminderDelay,
	)

	return nil
}

// cancelCheckoutSessionReminder cancels a pending reminder task for a checkout session.
func (s *checkoutSessionService) cancelCheckoutSessionReminder(
	ctx context.Context,
	checkoutSessionID uuid.UUID,
) error {
	// Create cancellation helper
	cancellationHelper := task.NewCancellationHelper(
		s.taskCancellationService,
		s.logger,
	)

	// Cancel the reminder task
	err := cancellationHelper.CancelCheckoutSessionReminderTask(ctx, checkoutSessionID)
	if err != nil {
		return fmt.Errorf("failed to cancel checkout session reminder task: %w", err)
	}

	s.logger.Infof("Cancelled checkout session reminder for session: %s", checkoutSessionID)

	return nil
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

	// Cancel checkout session reminder task since session was canceled
	if err = s.cancelCheckoutSessionReminder(ctx, sessionID); err != nil {
		// Log error but don't fail the cancellation
		s.logger.Warnf(
			"Failed to cancel checkout session reminder for session %s: %v",
			sessionID,
			err,
		)
	}

	return mapper.MapToCheckoutSessionResponse(updatedSession), nil
}
