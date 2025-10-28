package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/proto/cart/v1/cartv1connect"
	"github.com/shopspring/decimal"

	pkgconnect "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/cart/v1"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// CheckoutSession represents a checkout session retrieved from cart-service.
type CheckoutSession struct {
	CheckoutSessionID uuid.UUID
	IdempotencyKey    uuid.UUID
	CustomerID        uuid.UUID
	CartID            uuid.UUID
	Status            string
	Destination       Destination
	Origin            Origin
	Courier           Courier
	Package           Package
	Currency          string
	Items             []CheckoutSessionItem
	PaymentGateway    *string
}

// CheckoutSessionItem represents an item in a checkout session.
type CheckoutSessionItem struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	ProductName string
	Quantity    int64
	UnitPrice   decimal.Decimal
}

// Destination represents a shipping destination.
type Destination struct {
	City        string
	State       string
	PostalCode  string
	CountryCode string
}

// Origin represents a shipping origin.
type Origin struct {
	City        string
	State       string
	PostalCode  string
	CountryCode string
}

// Courier represents a courier.
type Courier struct {
	CourierID string
}

// Package represents a package.
type Package struct {
	WeightKG decimal.Decimal
	Length   decimal.Decimal
	Width    decimal.Decimal
	Height   decimal.Decimal
	Unit     string
}

// CartClient defines methods available for interacting with cart service.
type CartClient interface {
	GetCheckoutSession(ctx context.Context, checkoutSessionID uuid.UUID) (*CheckoutSession, error)
	HealthCheck(ctx context.Context) error
}

// cartClient is a Connect-RPC client for interacting with the cart service.
type cartClient struct {
	client cartv1connect.CartServiceClient
}

// NewCartClient creates a new cartClient instance with Connect-RPC.
func NewCartClient(
	cfg *config.Config,
) (CartClient, error) {
	// Create HTTP client for Connect-RPC
	httpClient := &http.Client{
		Timeout: constant.CartClientTimeout,
	}

	// Use static configuration for now
	baseURL := "http://" + net.JoinHostPort(
		cfg.Client.CartGRPCHost,
		strconv.Itoa(cfg.Client.CartGRPCPort),
	)

	// Create Connect-RPC client
	client := cartv1connect.NewCartServiceClient(httpClient, baseURL)

	return &cartClient{
		client: client,
	}, nil
}

// GetCheckoutSession retrieves a checkout session by ID.
//
//nolint:funlen
func (cc *cartClient) GetCheckoutSession(
	ctx context.Context,
	checkoutSessionID uuid.UUID,
) (*CheckoutSession, error) {
	ctx, cancel := context.WithTimeout(ctx, constant.CartClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.GetCheckoutSessionRequest{
		CheckoutSessionId: checkoutSessionID.String(),
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := cc.client.GetCheckoutSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetCheckoutSession: %w", err)
	}

	// Parse checkout session from proto
	sessionID, err := uuid.Parse(resp.Msg.GetCheckoutSessionId())
	if err != nil {
		return nil, fmt.Errorf("invalid checkout_session_id: %w", err)
	}

	idempotencyKey, err := uuid.Parse(resp.Msg.GetIdempotencyKey())
	if err != nil {
		return nil, fmt.Errorf("invalid idempotency_key: %w", err)
	}

	customerID, err := uuid.Parse(resp.Msg.GetCustomerId())
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}

	cartID, err := uuid.Parse(resp.Msg.GetCartId())
	if err != nil {
		return nil, fmt.Errorf("invalid cart_id: %w", err)
	}

	// Parse items
	items := make([]CheckoutSessionItem, len(resp.Msg.GetItems()))
	for i, item := range resp.Msg.GetItems() {
		itemID, errParse := uuid.Parse(item.GetId())
		if errParse != nil {
			return nil, fmt.Errorf("invalid item id: %w", errParse)
		}

		productID, errParse := uuid.Parse(item.GetProductId())
		if errParse != nil {
			return nil, fmt.Errorf("invalid product_id: %w", errParse)
		}

		unitPrice, errn := decimal.NewFromString(item.GetUnitPrice())
		if errn != nil {
			return nil, fmt.Errorf("invalid unit_price: %w", errn)
		}

		items[i] = CheckoutSessionItem{
			ID:          itemID,
			ProductID:   productID,
			ProductName: item.GetProductName(),
			Quantity:    item.GetQuantity(),
			UnitPrice:   unitPrice,
		}
	}

	// Parse package dimensions
	pkg := resp.Msg.GetPackage()

	weightKG, err := decimal.NewFromString(pkg.GetWeightKg())
	if err != nil {
		return nil, fmt.Errorf("invalid weight_kg: %w", err)
	}

	length, err := decimal.NewFromString(pkg.GetLength())
	if err != nil {
		return nil, fmt.Errorf("invalid length: %w", err)
	}

	width, err := decimal.NewFromString(pkg.GetWidth())
	if err != nil {
		return nil, fmt.Errorf("invalid width: %w", err)
	}

	height, err := decimal.NewFromString(pkg.GetHeight())
	if err != nil {
		return nil, fmt.Errorf("invalid height: %w", err)
	}

	// Parse destination
	dest := resp.Msg.GetDestination()
	destination := Destination{
		City:        dest.GetCity(),
		State:       dest.GetState(),
		PostalCode:  dest.GetPostalCode(),
		CountryCode: dest.GetCountryCode(),
	}

	// Parse origin
	orig := resp.Msg.GetOrigin()
	origin := Origin{
		City:        orig.GetCity(),
		State:       orig.GetState(),
		PostalCode:  orig.GetPostalCode(),
		CountryCode: orig.GetCountryCode(),
	}

	// Parse courier
	courier := Courier{
		CourierID: resp.Msg.GetCourier().GetCourierId(),
	}

	// Parse payment gateway (optional)
	var paymentGateway *string

	if resp.Msg.PaymentGateway != nil {
		pg := resp.Msg.GetPaymentGateway()
		paymentGateway = &pg
	}

	checkoutSession := &CheckoutSession{
		CheckoutSessionID: sessionID,
		IdempotencyKey:    idempotencyKey,
		CustomerID:        customerID,
		CartID:            cartID,
		Status:            resp.Msg.GetStatus().String(),
		Destination:       destination,
		Origin:            origin,
		Courier:           courier,
		Package: Package{
			WeightKG: weightKG,
			Length:   length,
			Width:    width,
			Height:   height,
			Unit:     pkg.GetUnit(),
		},
		Currency:       resp.Msg.GetCurrency(),
		Items:          items,
		PaymentGateway: paymentGateway,
	}

	return checkoutSession, nil
}

// HealthCheck verifies the connection to cart-service.
func (cc *cartClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, constant.CartClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.HealthRequest{})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := cc.client.Health(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Msg.GetStatus() != pb.HealthStatus_HEALTH_STATUS_SERVING {
		return fmt.Errorf("service unhealthy: %s", resp.Msg.GetStatus())
	}

	return nil
}
