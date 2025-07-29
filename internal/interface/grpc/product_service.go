// Package grpc provides the gRPC implementation for the ProductService.
package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/interfaces"
	marketplacev1 "github.com/raphaeldiscky/go-ddd-template/proto/marketplace/v1"
)

// ProductServiceServer implements the gRPC ProductService.
type ProductServiceServer struct {
	marketplacev1.UnimplementedProductServiceServer
	productService interfaces.ProductService
}

// NewProductServiceServer creates a new ProductServiceServer.
func NewProductServiceServer(productService interfaces.ProductService) *ProductServiceServer {
	return &ProductServiceServer{
		productService: productService,
	}
}

// CreateProduct creates a new product.
func (s *ProductServiceServer) CreateProduct(
	_ context.Context,
	req *marketplacev1.CreateProductRequest,
) (*marketplacev1.CreateProductResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "product name is required")
	}

	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product price must be greater than 0")
	}

	sellerID, err := uuid.Parse(req.SellerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid seller ID format")
	}

	createCmd := &command.CreateProductCommand{
		Name:     req.Name,
		Price:    req.Price,
		SellerId: sellerID,
	}

	result, err := s.productService.CreateProduct(createCmd)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create product: %v", err))
	}

	product := &marketplacev1.Product{
		Id:         result.Result.Id.String(),
		Name:       result.Result.Name,
		Price:      result.Result.Price,
		SellerId:   result.Result.Seller.Id.String(),
		SellerName: result.Result.Seller.Name,
		CreatedAt:  timestamppb.New(result.Result.CreatedAt),
		UpdatedAt:  timestamppb.New(result.Result.UpdatedAt),
	}

	return &marketplacev1.CreateProductResponse{
		Product: product,
	}, nil
}

// GetProduct retrieves a product by ID.
func (s *ProductServiceServer) GetProduct(
	_ context.Context,
	req *marketplacev1.GetProductRequest,
) (*marketplacev1.GetProductResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	result, err := s.productService.FindProductByID(productID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product: %v", err))
	}

	if result == nil || result.Result == nil {
		return nil, status.Error(codes.NotFound, "product not found")
	}

	product := &marketplacev1.Product{
		Id:         result.Result.Id.String(),
		Name:       result.Result.Name,
		Price:      result.Result.Price,
		SellerId:   result.Result.Seller.Id.String(),
		SellerName: result.Result.Seller.Name,
		CreatedAt:  timestamppb.New(result.Result.CreatedAt),
		UpdatedAt:  timestamppb.New(result.Result.UpdatedAt),
	}

	return &marketplacev1.GetProductResponse{
		Product: product,
	}, nil
}

// ListProducts lists all products with pagination.
func (s *ProductServiceServer) ListProducts(
	_ context.Context,
	req *marketplacev1.ListProductsRequest,
) (*marketplacev1.ListProductsResponse, error) {
	result, err := s.productService.FindAllProducts()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list products: %v", err))
	}

	var products []*marketplacev1.Product

	for _, productResult := range result.Result {
		product := &marketplacev1.Product{
			Id:         productResult.Id.String(),
			Name:       productResult.Name,
			Price:      productResult.Price,
			SellerId:   productResult.Seller.Id.String(),
			SellerName: productResult.Seller.Name,
			CreatedAt:  timestamppb.New(productResult.CreatedAt),
			UpdatedAt:  timestamppb.New(productResult.UpdatedAt),
		}
		products = append(products, product)
	}

	// TODO: Implement proper pagination
	total := int32(len(products))
	page := req.Page
	pageSize := req.PageSize

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	return &marketplacev1.ListProductsResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateProduct updates an existing product.
func (s *ProductServiceServer) UpdateProduct(
	_ context.Context,
	_ *marketplacev1.UpdateProductRequest,
) (*marketplacev1.UpdateProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "update product not implemented yet")
}

// DeleteProduct deletes a product by ID.
func (s *ProductServiceServer) DeleteProduct(
	_ context.Context,
	_ *marketplacev1.DeleteProductRequest,
) (*marketplacev1.DeleteProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "delete product not implemented yet")
}
