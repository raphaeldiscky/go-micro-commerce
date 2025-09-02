package temporal

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// ValidateProducts is the global activity function.
func ValidateProducts(ctx context.Context, order *entity.Order) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.ValidateProducts(ctx, order)
}

// ReserveProducts is the global activity function.
func ReserveProducts(ctx context.Context, order *entity.Order) ([]entity.Product, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return activities.ReserveProducts(ctx, order)
}

// CalculatePricing is the global activity function.
func CalculatePricing(ctx context.Context, order *entity.Order) (entity.OrderPricing, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return entity.OrderPricing{}, err
	}

	return activities.CalculatePricing(ctx, order)
}

// ProcessPayment is the global activity function.
func ProcessPayment(ctx context.Context, order *entity.Order) (uuid.UUID, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	return activities.ProcessPayment(ctx, order)
}

// ConfirmProductsDeduction is the global activity function.
func ConfirmProductsDeduction(ctx context.Context, order *entity.Order) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.ConfirmProductsDeduction(ctx, order)
}

// CreateShipping is the global activity function.
func CreateShipping(ctx context.Context, order *entity.Order) (dto.CreateShippingResponse, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return dto.CreateShippingResponse{}, err
	}

	return activities.CreateShipping(ctx, order)
}

// SendOrderConfirmation is the global activity function.
func SendOrderConfirmation(ctx context.Context, req dto.SendOrderConfirmationRequest) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.SendOrderConfirmation(ctx, req)
}

// ReleaseProducts is the global activity function.
func ReleaseProducts(ctx context.Context, order *entity.Order) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.ReleaseProducts(ctx, order)
}

// RefundPayment is the global activity function.
func RefundPayment(ctx context.Context, req dto.RefundPaymentRequest) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.RefundPayment(ctx, req)
}

// RestoreProducts is the global activity function.
func RestoreProducts(ctx context.Context, order *entity.Order) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.RestoreProducts(ctx, order)
}

// CancelShipping is the global activity function.
func CancelShipping(ctx context.Context, shippingID uuid.UUID) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.CancelShipping(ctx, shippingID)
}

// Helper function to get activities from context (this would be set by the worker).
func getActivitiesFromContext(ctx context.Context) (*OrderActivitiesImpl, error) {
	// In a real implementation, you would inject this via context or a global variable
	// For now, this is a placeholder that would need to be properly implemented
	// when setting up the worker
	activities := ctx.Value("temporal_activities")
	if activities == nil {
		return nil, fmt.Errorf("temporal activities not found in context")
	}

	activitiesImpl, ok := activities.(*OrderActivitiesImpl)
	if !ok {
		return nil, fmt.Errorf("temporal activities type is incorrect")
	}

	return activitiesImpl, nil
}
