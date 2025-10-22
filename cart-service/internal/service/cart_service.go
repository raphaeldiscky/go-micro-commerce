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
	AddItemToCart(
		ctx context.Context,
		cartID uuid.UUID,
		req *dto.AddCartItemRequest,
	) (*dto.CartResponse, error)
	RemoveItemFromCart(
		ctx context.Context,
		cartID uuid.UUID,
		itemID uuid.UUID,
	) (*dto.CartResponse, error)
	UpdateItemQuantity(
		ctx context.Context,
		cartID uuid.UUID,
		itemID uuid.UUID,
		quantity int64,
	) (*dto.CartResponse, error)
	SelectItemForCheckout(
		ctx context.Context,
		cartID uuid.UUID,
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

// AddItemToCart adds an item to the cart.
func (s *cartService) AddItemToCart(
	ctx context.Context,
	cartID uuid.UUID,
	req *dto.AddCartItemRequest,
) (*dto.CartResponse, error) {
	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Check if cart exists
		cart, errCart := cartRepo.FindByID(ctx, cartID)
		if errCart != nil {
			return httperror.NewInternalServerError("failed to get cart")
		}

		if cart == nil {
			return httperror.NewCartNotFoundError()
		}

		// Create cart item
		item, errItem := entity.NewCartItem(req.ProductID, req.Quantity)
		if errItem != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("invalid item: %v", errItem))
		}

		// Add item to cart
		err := cartRepo.AddItem(ctx, cartID, item)
		if err != nil {
			return httperror.NewInternalServerError("failed to add item to cart")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cartID)
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

// RemoveItemFromCart removes an item from the cart.
func (s *cartService) RemoveItemFromCart(
	ctx context.Context,
	cartID uuid.UUID,
	itemID uuid.UUID,
) (*dto.CartResponse, error) {
	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Check if cart exists
		cart, errCart := cartRepo.FindByID(ctx, cartID)
		if errCart != nil {
			return httperror.NewInternalServerError("failed to get cart")
		}

		if cart == nil {
			return httperror.NewCartNotFoundError()
		}

		// Remove item from cart
		err := cartRepo.RemoveItem(ctx, cartID, itemID)
		if err != nil {
			return httperror.NewInternalServerError("failed to remove item from cart")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cartID)
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

// UpdateItemQuantity updates the quantity of a cart item.
func (s *cartService) UpdateItemQuantity(
	ctx context.Context,
	cartID uuid.UUID,
	itemID uuid.UUID,
	quantity int64,
) (*dto.CartResponse, error) {
	if quantity <= 0 {
		return nil, httperror.NewBadRequestError("quantity must be greater than 0")
	}

	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Check if cart exists
		cart, errCart := cartRepo.FindByID(ctx, cartID)
		if errCart != nil {
			return httperror.NewInternalServerError("failed to get cart")
		}

		if cart == nil {
			return httperror.NewCartNotFoundError()
		}

		// Update item quantity
		err := cartRepo.UpdateItemQuantity(ctx, cartID, itemID, quantity)
		if err != nil {
			return httperror.NewInternalServerError("failed to update item quantity")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cartID)
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

// SelectItemForCheckout marks an item as selected for checkout.
func (s *cartService) SelectItemForCheckout(
	ctx context.Context,
	cartID uuid.UUID,
	itemID uuid.UUID,
	selected bool,
) (*dto.CartResponse, error) {
	var updatedCart *entity.Cart

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		cartRepo := ds.CartRepository()

		// Check if cart exists
		cart, errCart := cartRepo.FindByID(ctx, cartID)
		if errCart != nil {
			return httperror.NewInternalServerError("failed to get cart")
		}

		if cart == nil {
			return httperror.NewCartNotFoundError()
		}

		// Select/deselect item for checkout
		err := cartRepo.SelectForCheckout(ctx, cartID, itemID, selected)
		if err != nil {
			return httperror.NewInternalServerError("failed to update item selection")
		}

		// Fetch updated cart
		updatedCart, err = cartRepo.FindByID(ctx, cartID)
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
