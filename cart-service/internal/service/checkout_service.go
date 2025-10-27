// Package service provides business logic for checkout session operations.
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
	"github.com/shopspring/decimal"

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
	UpdateCheckoutSession(
		ctx context.Context,
		sessionID uuid.UUID,
		req *dto.UpdateCheckoutSessionRequest,
	) (*dto.CheckoutSessionResponse, error)
	PlaceOrder(
		ctx context.Context,
		req *dto.PlaceOrderRequest,
	) (*dto.PlaceOrderResponse, error)
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
	paymentClient                      client.PaymentClient
	fulfillmentClient                  client.FulfillmentClient
	checkoutSessionOrderPlacedProducer kafka.Producer
}

// NewCheckoutSessionService creates a new instance of checkoutSessionService.
func NewCheckoutSessionService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	productClient client.ProductClient,
	paymentClient client.PaymentClient,
	fulfillmentClient client.FulfillmentClient,
	checkoutSessionOrderPlacedProducer kafka.Producer,
) CheckoutSessionService {
	return &checkoutSessionService{
		dataStore:                          dataStore,
		logger:                             appLogger,
		productClient:                      productClient,
		paymentClient:                      paymentClient,
		fulfillmentClient:                  fulfillmentClient,
		checkoutSessionOrderPlacedProducer: checkoutSessionOrderPlacedProducer,
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
//
//nolint:gocyclo,cyclop,funlen // ignore complexity
func (s *checkoutSessionService) PlaceOrder(
	ctx context.Context,
	req *dto.PlaceOrderRequest,
) (*dto.PlaceOrderResponse, error) {
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
	var paymentResponse *dto.CreatePaymentIntentResponse
	var paymentGateway string

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

		// Calculate shipping cost via fulfillment service
		shippingCost, shippingErr := s.calculateShippingCost(ctx, session)
		if shippingErr != nil {
			return httperror.NewInternalServerError(
				fmt.Sprintf("failed to calculate shipping cost: %v", shippingErr),
			)
		}

		// Create PaymentIntent immediately for frontend redirect
		totalAmount := decimal.Zero
		for _, item := range session.Items {
			totalAmount = totalAmount.Add(item.UnitPrice.Mul(decimal.NewFromInt(item.Quantity)))
		}
		// Add shipping cost to total amount
		totalAmount = totalAmount.Add(shippingCost)

		// Update session with calculated costs
		if setErr := session.SetShippingCost(shippingCost); setErr != nil {
			return httperror.NewInternalServerError(
				fmt.Sprintf("failed to set shipping cost: %v", setErr),
			)
		}

		// Prepare items for payment request
		paymentItems := make([]dto.PaymentItemDTO, len(session.Items))
		for i, item := range session.Items {
			paymentItems[i] = dto.PaymentItemDTO{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Currency:    session.Currency,
			}
		}

		// Create PaymentIntent via payment client
		paymentResp, paymentErr := s.paymentClient.CreatePaymentIntent(
			ctx,
			&dto.CreatePaymentIntentRequest{
				OrderID:           session.ID,
				CustomerID:        req.CustomerID,
				CustomerEmail:     req.CustomerEmail,
				Amount:            totalAmount,
				Currency:          session.Currency,
				PaymentGateway:    *session.PaymentGateway,
				IdempotencyKey:    req.IdempotencyKey,
				Items:             paymentItems,
				CheckoutSessionID: session.ID,
			},
		)
		if paymentErr != nil {
			s.logger.Errorf("Failed to create PaymentIntent: %v", paymentErr)
			return httperror.NewInternalServerError("failed to create payment intent")
		}

		s.logger.Infof(
			"PaymentIntent created successfully - PaymentIntentID: %s, OrderID: %s, Amount: %s",
			paymentResp.PaymentIntentID,
			session.ID,
			totalAmount.String(),
		)

		// Capture payment response data for frontend response
		paymentResponse = paymentResp
		paymentGateway = *session.PaymentGateway

		// Update status to order_placed
		if err = session.UpdateStatus(constant.CheckoutSessionStatusOrderPlaced); err != nil {
			return httperror.NewBadRequestError("failed to update checkout session status")
		}

		// Save updated session
		updatedSession, err = checkoutSessionRepo.Update(ctx, session)
		if err != nil {
			return httperror.NewInternalServerError("failed to update checkout session")
		}

		// Build gateway metadata with PaymentIntent details for Stripe
		var gatewayMetadata json.RawMessage

		if *updatedSession.PaymentGateway == "stripe" && paymentResp != nil {
			// Use same structure as payment-service entity.StripeMetadata
			stripeMetadata := struct {
				PaymentIntentID *string `json:"payment_intent_id,omitempty"`
			}{
				PaymentIntentID: &paymentResp.PaymentIntentID,
			}

			metadataBytes, errMarshal := json.Marshal(stripeMetadata)
			if errMarshal != nil {
				s.logger.Errorf("Failed to marshal gateway metadata: %v", errMarshal)
				// Use empty metadata if marshaling fails
				gatewayMetadata = json.RawMessage("{}")
			} else {
				gatewayMetadata = json.RawMessage(metadataBytes)
			}
		} else {
			// Empty metadata for non-Stripe gateways
			gatewayMetadata = json.RawMessage("{}")
		}

		// Publish domain event via outbox pattern
		s.logger.Debugf(
			"Creating checkout session order placed event - SessionID: %s, CustomerID: %s, IdempotencyKey: %s",
			updatedSession.ID,
			updatedSession.CustomerID,
			req.IdempotencyKey,
		)

		evt := producer.NewCheckoutSessionOrderPlacedEvent(
			updatedSession.ID,
			req.IdempotencyKey,
			constant.CheckoutSessionStatusOrderPlaced,
			updatedSession.CustomerID,
			updatedSession.Currency,
			*updatedSession.PaymentGateway,
			updatedSession.Items,
			updatedSession.ShippingCost,
			updatedSession.TotalAmount,
			updatedSession.Courier,
			updatedSession.Destination,
			updatedSession.Origin,
			updatedSession.Package,
			updatedSession.CreatedAt,
			gatewayMetadata, // Include PaymentIntent details for Stripe
		)

		s.logger.Debugf(
			"Created event payload - CheckoutSessionID: %s, EventType: %s",
			evt.Payload.CheckoutSessionID,
			evt.Metadata.EventType,
		)

		// Marshal event to JSON for storage in outbox
		eventJSON, errMarshal := json.Marshal(evt)
		if errMarshal != nil {
			return httperror.NewInternalServerError("failed to marshal checkout session event")
		}

		// Get outbox repository from DataStore (uses same transaction)
		outboxRepo := ds.OutboxRepository()

		// Create outbox event
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "checkout_session",
			AggregateID:   updatedSession.ID,
			EventType:     evt.Metadata.EventType,
			Topic:         s.checkoutSessionOrderPlacedProducer.Topic(),
			Payload:       eventJSON,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now(),
			ScheduledFor:  time.Now(),
			Attempts:      0,
		}

		// Save to outbox table (within same transaction)
		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError(
				fmt.Sprintf("failed to save event to outbox: %v", err),
			)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Migrate cart after successful order placement (non-critical operation)
	// This runs outside the transaction so failures won't affect order placement
	cartRepo := s.dataStore.CartRepository()
	s.migrateCartAfterOrderPlacement(ctx, cartRepo, req.CustomerID)

	// Build standardized PlaceOrderResponse for immediate frontend redirect
	// Build gateway metadata for frontend consumption
	var finalGatewayMetadata json.RawMessage
	var transactionID, redirectURL string

	if paymentResponse != nil {
		// Build gateway-specific metadata for frontend
		if paymentGateway == "stripe" {
			stripeData := struct {
				ClientSecret    string `json:"client_secret"`
				PaymentIntentID string `json:"payment_intent_id"`
			}{
				ClientSecret:    paymentResponse.ClientSecret,
				PaymentIntentID: paymentResponse.PaymentIntentID,
			}
			metadataBytes, _ := json.Marshal(stripeData)
			finalGatewayMetadata = json.RawMessage(metadataBytes)
			transactionID = paymentResponse.PaymentIntentID
			redirectURL = "" // Stripe uses client_secret, not redirect URL
		} else {
			finalGatewayMetadata = json.RawMessage("{}")
		}
	} else {
		finalGatewayMetadata = json.RawMessage("{}")
	}

	// Create standardized response
	checkoutSessionResp := mapper.MapToCheckoutSessionResponse(updatedSession)
	placeOrderResp := &dto.PlaceOrderResponse{
		CheckoutSession: *checkoutSessionResp,
		TransactionID:   transactionID,
		Amount:          updatedSession.TotalAmount.String(),
		Currency:        updatedSession.Currency,
		Status:          "pending", // PaymentIntent is created but awaiting confirmation
		RedirectURL:     redirectURL,
		GatewayMetadata: finalGatewayMetadata,
	}

	return placeOrderResp, nil
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

// calculateShippingCost calculates the shipping cost for the given checkout session.
func (s *checkoutSessionService) calculateShippingCost(
	ctx context.Context,
	session *entity.CheckoutSession,
) (decimal.Decimal, error) {
	// Create shipping cost request
	req := &dto.GetShippingCostRequest{
		Currency:               session.Currency,
		CourierID:              session.Courier.CourierID,
		DestinationCity:        session.Destination.City,
		DestinationState:       session.Destination.State,
		DestinationPostalCode:  session.Destination.PostalCode,
		DestinationCountryCode: session.Destination.CountryCode,
		OriginCity:             session.Origin.City,
		OriginState:            session.Origin.State,
		OriginPostalCode:       session.Origin.PostalCode,
		OriginCountryCode:      session.Origin.CountryCode,
		WeightKG:               session.Package.WeightKG.String(),
		Width:                  session.Package.Width.String(),
		Height:                 session.Package.Height.String(),
		Length:                 session.Package.Length.String(),
		Unit:                   session.Package.Unit,
	}

	// Call fulfillment service
	resp, err := s.fulfillmentClient.GetShippingCost(ctx, req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get shipping cost: %w", err)
	}

	if !resp.Success {
		return decimal.Zero, fmt.Errorf("shipping cost calculation failed: %s", resp.ErrorMessage)
	}

	// Convert float64 to decimal for financial precision
	shippingCost := decimal.NewFromFloat(resp.ShippingCost)

	s.logger.Infof(
		"Shipping cost calculated successfully - SessionID: %s, Cost: %s %s",
		session.ID,
		shippingCost.String(),
		session.Currency,
	)

	return shippingCost, nil
}
