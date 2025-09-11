package mapper

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// MapProtoToProduct maps a proto.Product to an entity.
func MapProtoToProduct(p *pb.Product) (entity.Product, error) {
	uid, err := uuid.Parse(p.GetId())
	if err != nil {
		return entity.Product{}, fmt.Errorf("invalid product ID from product-service: %w", err)
	}

	// Convert protobuf Timestamp → time.Time safely
	var createdAt, updatedAt time.Time
	if p.GetCreatedAt() != nil {
		createdAt = p.GetCreatedAt().AsTime()
	}

	if p.GetUpdatedAt() != nil {
		updatedAt = p.GetUpdatedAt().AsTime()
	}

	return entity.Product{
		ID:               uid,
		Name:             p.GetName(),
		UnitPrice:        decimal.NewFromFloat(p.GetPrice()), // safely convert double → decimal
		Quantity:         p.GetQuantity(),
		Version:          p.GetVersion(),
		ReservedQuantity: p.GetReservedQuantity(),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}
