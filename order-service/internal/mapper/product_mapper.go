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
	uid, err := uuid.Parse(p.Id)
	if err != nil {
		return entity.Product{}, fmt.Errorf("invalid product ID from product-service: %w", err)
	}

	// Convert protobuf Timestamp → time.Time safely
	var createdAt, updatedAt time.Time
	if p.CreatedAt != nil {
		createdAt = p.CreatedAt.AsTime()
	}

	if p.UpdatedAt != nil {
		updatedAt = p.UpdatedAt.AsTime()
	}

	return entity.Product{
		ID:               uid,
		Name:             p.Name,
		Price:            decimal.NewFromFloat(p.Price), // safely convert double → decimal
		Quantity:         p.Quantity,
		Version:          p.Version,
		ReservedQuantity: p.ReservedQuantity,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}
