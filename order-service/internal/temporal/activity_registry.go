package temporal

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// ReserveProductsAndCalculate is the global activity function.
func ReserveProductsAndCalculate(
	ctx context.Context,
	req dto.ReserveProductsAndCalculateRequest,
) (dto.ReserveProductsAndCalculateResponse, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return dto.ReserveProductsAndCalculateResponse{}, err
	}

	return activities.ReserveProductsAndCalculate(ctx, req)
}

// ProcessFulfillment is the global activity function.
func ProcessFulfillment(
	ctx context.Context,
	order *entity.Order,
) (dto.ProcessFulfillmentResponse, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return dto.ProcessFulfillmentResponse{}, err
	}

	return activities.ProcessFulfillment(ctx, order)
}

// SetFinalOrderPrices is the global activity function.
func SetFinalOrderPrices(
	ctx context.Context,
	req dto.SetFinalOrderPricesRequest,
) (dto.SetFinalOrderPricesResponse, error) {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return dto.SetFinalOrderPricesResponse{}, err
	}

	return activities.SetFinalOrderPrices(ctx, req)
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
func ConfirmProductsDeduction(ctx context.Context, req *dto.ConfirmProductsDeductionRequest) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.ConfirmProductsDeduction(ctx, req)
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
func ReleaseProducts(ctx context.Context, req dto.ReleaseProductsRequest) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.ReleaseProducts(ctx, req)
}

// RefundPayment is the global activity function.
func RefundPayment(ctx context.Context, req dto.RefundPaymentGatewayRequest) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.RefundPayment(ctx, req)
}

// RestoreProducts is the global activity function.
func RestoreProducts(ctx context.Context, req dto.RestoreProductsRequest) error {
	activities, err := getActivitiesFromContext(ctx)
	if err != nil {
		return err
	}

	return activities.RestoreProducts(ctx, req)
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
