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
	marketplacev1 "github.com/raphaeldiscky/go-ddd-template/proto"
)

// SellerServiceServer implements the gRPC SellerService.
type SellerServiceServer struct {
	marketplacev1.UnimplementedSellerServiceServer
	sellerService interfaces.SellerService
}

// NewSellerServiceServer creates a new SellerServiceServer.
func NewSellerServiceServer(sellerService interfaces.SellerService) *SellerServiceServer {
	return &SellerServiceServer{
		sellerService: sellerService,
	}
}

// CreateSeller creates a new seller.
func (s *SellerServiceServer) CreateSeller(
	_ context.Context,
	req *marketplacev1.CreateSellerRequest,
) (*marketplacev1.CreateSellerResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "seller name is required")
	}

	createCmd := &command.CreateSellerCommand{
		Name: req.Name,
	}

	result, err := s.sellerService.CreateSeller(createCmd)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create seller: %v", err))
	}

	seller := &marketplacev1.Seller{
		Id:        result.Result.Id.String(),
		Name:      result.Result.Name,
		CreatedAt: timestamppb.New(result.Result.CreatedAt),
		UpdatedAt: timestamppb.New(result.Result.UpdatedAt),
	}

	return &marketplacev1.CreateSellerResponse{
		Seller: seller,
	}, nil
}

// GetSeller retrieves a seller by ID.
func (s *SellerServiceServer) GetSeller(
	_ context.Context,
	req *marketplacev1.GetSellerRequest,
) (*marketplacev1.GetSellerResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "seller ID is required")
	}

	sellerID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid seller ID format")
	}

	result, err := s.sellerService.FindSellerByID(sellerID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get seller: %v", err))
	}

	if result == nil || result.Result == nil {
		return nil, status.Error(codes.NotFound, "seller not found")
	}

	seller := &marketplacev1.Seller{
		Id:        result.Result.Id.String(),
		Name:      result.Result.Name,
		CreatedAt: timestamppb.New(result.Result.CreatedAt),
		UpdatedAt: timestamppb.New(result.Result.UpdatedAt),
	}

	return &marketplacev1.GetSellerResponse{
		Seller: seller,
	}, nil
}

// ListSellers lists all sellers with pagination.
func (s *SellerServiceServer) ListSellers(
	_ context.Context,
	req *marketplacev1.ListSellersRequest,
) (*marketplacev1.ListSellersResponse, error) {
	result, err := s.sellerService.FindAllSellers()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list sellers: %v", err))
	}

	var sellers []*marketplacev1.Seller

	for _, sellerResult := range result.Result {
		seller := &marketplacev1.Seller{
			Id:        sellerResult.Id.String(),
			Name:      sellerResult.Name,
			CreatedAt: timestamppb.New(sellerResult.CreatedAt),
			UpdatedAt: timestamppb.New(sellerResult.UpdatedAt),
		}
		sellers = append(sellers, seller)
	}

	// TODO: Implement proper pagination
	total := int32(len(sellers))
	page := req.Page
	pageSize := req.PageSize

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	return &marketplacev1.ListSellersResponse{
		Sellers:  sellers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateSeller updates an existing seller.
func (s *SellerServiceServer) UpdateSeller(
	_ context.Context,
	_ *marketplacev1.UpdateSellerRequest,
) (*marketplacev1.UpdateSellerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "update seller not implemented yet")
}

// DeleteSeller deletes a seller by ID.
func (s *SellerServiceServer) DeleteSeller(
	_ context.Context,
	_ *marketplacev1.DeleteSellerRequest,
) (*marketplacev1.DeleteSellerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "delete seller not implemented yet")
}
