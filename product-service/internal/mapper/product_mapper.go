// Package mapper provides functions for mapping entity.Product to dto.ProductResponse.
package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product/v1"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
)

// MapToProductResponse converts domain entity to DTO response.
func MapToProductResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:                product.ID,
		Name:              product.Name,
		Price:             product.Price,
		Quantity:          product.Quantity,
		Version:           product.Version,
		ReservedQuantity:  product.ReservedQuantity,
		AvailableQuantity: product.GetAvailableStock(),
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
	}
}

// MapToProtobufProduct converts domain entity to protobuf Product message.
func MapToProtobufProduct(product *entity.Product) *pb.Product {
	return &pb.Product{
		Id:               product.ID.String(),
		Name:             product.Name,
		Price:            product.Price.InexactFloat64(),
		Quantity:         product.Quantity,
		Version:          product.Version,
		ReservedQuantity: product.ReservedQuantity,
		CreatedAt:        timestamppb.New(product.CreatedAt),
		UpdatedAt:        timestamppb.New(product.UpdatedAt),
	}
}

// MapToProtobufProducts converts slice of domain entities to protobuf Product messages.
func MapToProtobufProducts(products []entity.Product) []*pb.Product {
	result := make([]*pb.Product, len(products))
	for i := range products {
		result[i] = MapToProtobufProduct(&products[i])
	}

	return result
}

// MapDTOToProtobufProduct converts DTO to protobuf Product message.
func MapDTOToProtobufProduct(product *dto.ProductResponse) *pb.Product {
	return &pb.Product{
		Id:               product.ID.String(),
		Name:             product.Name,
		Price:            product.Price.InexactFloat64(),
		Quantity:         product.Quantity,
		Version:          product.Version,
		ReservedQuantity: product.ReservedQuantity,
		CreatedAt:        timestamppb.New(product.CreatedAt),
		UpdatedAt:        timestamppb.New(product.UpdatedAt),
	}
}

// MapDTOToProtobufProducts converts slice of DTOs to protobuf Product messages.
func MapDTOToProtobufProducts(products []dto.ProductResponse) []*pb.Product {
	result := make([]*pb.Product, len(products))
	for i := range products {
		result[i] = MapDTOToProtobufProduct(&products[i])
	}

	return result
}
