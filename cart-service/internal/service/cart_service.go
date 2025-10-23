// Package service provides business logic for cart operations.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
)

// CartService defines the interface for cart business operations.
// Cart is lightweight - only handles item persistence and selection.
// Pricing calculations are handled by CheckoutSession.
type CartService interface {
	CreateCart(ctx context.Context, req *dto.CreateCartRequest) (*dto.CartResponse, error)
	GetCart(ctx context.Context, cartID uuid.UUID) (*dto.CartResponse, error)
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*dto.CartResponse, error)
	AddItemToActiveCart(
		ctx context.Context,
		req *dto.AddCartItemRequest,
	) (*dto.CartResponse, error)
	RemoveItemFromActiveCart(
		ctx context.Context,
		customerID uuid.UUID,
		itemID uuid.UUID,
	) (*dto.CartResponse, error)
	UpdateActiveCartItemQuantity(
		ctx context.Context,
		customerID uuid.UUID,
		itemID uuid.UUID,
		quantity int64,
	) (*dto.CartResponse, error)
	SelectActiveCartItemForCheckout(
		ctx context.Context,
		customerID uuid.UUID,
		itemID uuid.UUID,
		selected bool,
	) (*dto.CartResponse, error)
}

// cartService implements the CartService.
type cartService struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewCartService creates a new instance of cartService.
func NewCartService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) CartService {
	return &cartService{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// CreateCart creates a new cart for a customer.
func (s *cartService) CreateCart(
	ctx context.Context,
	req *dto.CreateCartRequest,
) (*dto.CartResponse, error) {
	var createdCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Create cart items
		var items []entity.CartItem

		for i := range req.Items {
			item, errItem := entity.NewCartItem(req.Items[i].ProductID, req.Items[i].Quantity)
			if errItem != nil {
				s.logger.Errorf("failed to create cart item: %v", errItem)
				return httperror.NewBadRequestError(fmt.Sprintf("invalid item: %v", errItem))
			}

			items = append(items, *item)
		}

		// Create cart entity
		cart, errCart := entity.NewCart(req.CustomerID, items)
		if errCart != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("failed to create cart: %v", errCart))
		}

		// Save cart to database
		newCart, err := cartRepo.Create(ctx, cart)
		if err != nil {
			return httperror.NewInternalServerError("failed to save cart")
		}

		createdCart = newCart

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCartResponse(createdCart), nil
}

// GetCart retrieves a cart by ID.
func (s *cartService) GetCart(
	ctx context.Context,
	cartID uuid.UUID,
) (*dto.CartResponse, error) {
	cartRepo := s.dataStore.CartRepository()

	cart, err := cartRepo.FindByID(ctx, cartID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get cart")
	}

	if cart == nil {
		return nil, httperror.NewCartNotFoundError()
	}

	return mapper.MapToCartResponse(cart), nil
}

// GetCartByUserID retrieves a cart by user ID.
func (s *cartService) GetCartByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.CartResponse, error) {
	cartRepo := s.dataStore.CartRepository()

	cart, err := cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get cart")
	}

	if cart == nil {
		return nil, httperror.NewCartNotFoundError()
	}

	return mapper.MapToCartResponse(cart), nil
}

// AddItemToCart adds an item to the active cart.
func (s *cartService) AddItemToActiveCart(
	ctx context.Context,
	req *dto.AddCartItemRequest,
) (*dto.CartResponse, error) {
	var resultCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Try to find active cart for customer
		cart, errCart := cartRepo.FindActiveCartByUserID(ctx, req.CustomerID)

		// If no active cart exists, create a new one with the item
		if errCart != nil {
			// Create cart item
			item, errItem := entity.NewCartItem(req.ProductID, req.Quantity)
			if errItem != nil {
				return httperror.NewBadRequestError(fmt.Sprintf("invalid item: %v", errItem))
			}

			// Create new cart with status active and the item
			newCart, errNewCart := entity.NewCart(req.CustomerID, []entity.CartItem{*item})
			if errNewCart != nil {
				return httperror.NewBadRequestError(
					fmt.Sprintf("failed to create cart: %v", errNewCart),
				)
			}

			// Save new cart to database
			createdCart, err := cartRepo.Create(ctx, newCart)
			if err != nil {
				return httperror.NewInternalServerError("failed to create cart")
			}

			resultCart = createdCart

			return nil
		}

		// Cart exists - add item to existing cart
		item, errItem := entity.NewCartItem(req.ProductID, req.Quantity)
		if errItem != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("invalid item: %v", errItem))
		}

		// Add item to cart
		err := cartRepo.AddItem(ctx, cart.ID, item)
		if err != nil {
			return httperror.NewInternalServerError("failed to add item to cart")
		}

		// Fetch updated cart
		resultCart, err = cartRepo.FindByID(ctx, cart.ID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get updated cart")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCartResponse(resultCart), nil
}

// RemoveItemFromActiveCart removes an item from the active cart.
func (s *cartService) RemoveItemFromActiveCart(
	ctx context.Context,
	customerID uuid.UUID,
	itemID uuid.UUID,
) (*dto.CartResponse, error) {
	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Find active cart by customerID
		cart, errCart := cartRepo.FindActiveCartByUserID(ctx, customerID)
		if errCart != nil {
			return httperror.NewCartNotFoundError()
		}

		// Remove item from cart (repository enforces active status)
		err := cartRepo.RemoveItem(ctx, itemID)
		if err != nil {
			return httperror.NewInternalServerError("failed to remove item from cart")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cart.ID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get updated cart")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCartResponse(updatedCart), nil
}

// UpdateActiveCartItemQuantity updates the quantity of a cart item in the active cart.
func (s *cartService) UpdateActiveCartItemQuantity(
	ctx context.Context,
	customerID uuid.UUID,
	itemID uuid.UUID,
	quantity int64,
) (*dto.CartResponse, error) {
	if quantity <= 0 {
		return nil, httperror.NewBadRequestError("quantity must be greater than 0")
	}

	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Find active cart by customerID
		cart, errCart := cartRepo.FindActiveCartByUserID(ctx, customerID)
		if errCart != nil {
			s.logger.Errorf("failed to get active cart: %v", errCart)
			return httperror.NewCartNotFoundError()
		}

		// Update item quantity (repository enforces active status)
		err := cartRepo.UpdateActiveCartItemQuantity(ctx, itemID, quantity)
		if err != nil {
			s.logger.Errorf("failed to update item quantity: %v", err)
			return httperror.NewInternalServerError("failed to update item quantity")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cart.ID)
		if err != nil {
			s.logger.Errorf("failed to get updated cart: %v", err)
			return httperror.NewInternalServerError("failed to get updated cart")
		}

		return nil
	})
	if err != nil {
		s.logger.Errorf("failed to update item quantity: %v", err)
		return nil, err
	}

	return mapper.MapToCartResponse(updatedCart), nil
}

// SelectActiveCartItemForCheckout marks an item as selected for checkout in the active cart.
func (s *cartService) SelectActiveCartItemForCheckout(
	ctx context.Context,
	customerID uuid.UUID,
	itemID uuid.UUID,
	selected bool,
) (*dto.CartResponse, error) {
	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Find active cart by customerID
		cart, errCart := cartRepo.FindActiveCartByUserID(ctx, customerID)
		if errCart != nil {
			return httperror.NewCartNotFoundError()
		}

		// Select/deselect item for checkout (repository enforces active status)
		err := cartRepo.SelectForCheckout(ctx, itemID, selected)
		if err != nil {
			return httperror.NewInternalServerError("failed to update item selection")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cart.ID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get updated cart")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mapper.MapToCartResponse(updatedCart), nil
}
