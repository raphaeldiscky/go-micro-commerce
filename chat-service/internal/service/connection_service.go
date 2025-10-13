// Package service provides business logic for chat connection management.
package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"

	pkgConfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/httperror"
)

// ConnectionService defines the interface for chat connection management.
type ConnectionService interface {
	// RequestConnection selects optimal chat node for connection
	RequestConnection(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) (*dto.ChatConnectionResponse, error)

	// ValidateAuthToken validates an auth service JWT and extracts user information
	ValidateAuthToken(
		ctx context.Context,
		token string,
	) (*dto.AuthTokenClaims, error)

	// GetNodeHealth returns health status of available chat nodes
	GetNodeHealth(ctx context.Context) ([]dto.NodeHealthResponse, error)
}

// connectionService implements the ConnectionService interface.
type connectionService struct {
	logger       logger.Logger
	jwtUtils     jwtutils.JWT
	nodeConfig   *NodeConfig
	consulClient *api.Client
}

// NodeConfig holds configuration for chat node selection.
type NodeConfig struct {
	DefaultNodeAddress string
	MaxConnections     int
	ConsulAddress      string
	ChatServiceName    string
}

// NewConnectionService creates a new instance of connectionService.
func NewConnectionService(
	appLogger logger.Logger,
	publicKeyPath string,
	jwksURL string,
	jwksCacheTTL time.Duration,
	jwksRefreshInterval time.Duration,
	nodeConfig *NodeConfig,
) ConnectionService {
	// Initialize JWT utils with JWKS support
	jwtUtil := jwtutils.NewJWTUtils(&pkgConfig.JWTConfig{
		PublicKeyPath:       publicKeyPath,
		JWKSUrl:             jwksURL,
		JWKSCacheTTL:        jwksCacheTTL,
		JWKSRefreshInterval: jwksRefreshInterval,
		AllowedAlgs:         []string{"RS256"},
		SigningMethod:       "RS256",
	}, appLogger)

	// Initialize Consul client for service discovery
	var consulClient *api.Client
	if nodeConfig.ConsulAddress != "" {
		consulConfig := api.DefaultConfig()
		consulConfig.Address = nodeConfig.ConsulAddress

		var consulErr error

		consulClient, consulErr = api.NewClient(consulConfig)
		if consulErr != nil {
			appLogger.Warn("Failed to initialize Consul client, using fallback", "error", consulErr)

			consulClient = nil
		}
	}

	return &connectionService{
		logger:       appLogger,
		jwtUtils:     jwtUtil,
		nodeConfig:   nodeConfig,
		consulClient: consulClient,
	}
}

// RequestConnection selects optimal chat node for connection.
func (s *connectionService) RequestConnection(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) (*dto.ChatConnectionResponse, error) {
	// Select optimal node using service discovery and load balancing
	selectedNode := s.selectOptimalNode(ctx)

	s.logger.Info("Selected chat node for connection",
		"user_id", userID,
		"user_type", userType,
		"node_address", selectedNode)

	return &dto.ChatConnectionResponse{
		NodeAddress: selectedNode,
		UserID:      userID,
		UserType:    userType,
	}, nil
}

// ValidateAuthToken validates an auth service JWT and extracts user information.
func (s *connectionService) ValidateAuthToken(
	_ context.Context,
	token string,
) (*dto.AuthTokenClaims, error) {
	// Use jwtUtils to validate the token (supports JWKS with caching)
	claims, err := s.jwtUtils.ValidateAccessToken(token)
	if err != nil {
		s.logger.Warn("Invalid auth token", "error", err)
		return nil, httperror.NewUnauthorizedError("invalid auth token")
	}

	// Map jwtutils.AccessTokenClaims to dto.AuthTokenClaims
	authClaims := &dto.AuthTokenClaims{
		RegisteredClaims: claims.RegisteredClaims,
		UserID:           claims.UserID,
		Email:            claims.Email,
		Roles:            claims.Roles,
		IsActive:         claims.IsActive,
	}

	s.logger.Info("Successfully validated auth token",
		"user_id", authClaims.UserID,
		"email", authClaims.Email)

	return authClaims, nil
}

// GetNodeHealth returns health status of available chat nodes.
func (s *connectionService) GetNodeHealth(_ context.Context) ([]dto.NodeHealthResponse, error) {
	if s.consulClient == nil {
		// Fallback to default node when Consul is unavailable
		nodes := []dto.NodeHealthResponse{
			{
				NodeID:         "chat-node-default",
				Address:        s.nodeConfig.DefaultNodeAddress,
				Status:         "healthy",
				Connections:    0,
				MaxConnections: s.nodeConfig.MaxConnections,
				LastSeen:       time.Now(),
			},
		}

		return nodes, nil
	}

	// Query Consul for healthy chat service instances
	serviceName := s.nodeConfig.ChatServiceName
	if serviceName == "" {
		s.logger.Error("missing node service name")
	}

	healthyServices, _, err := s.consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		s.logger.Error("Failed to query Consul for chat nodes", "error", err)
		// Return fallback node
		return s.getFallbackNodes(), nil
	}

	if len(healthyServices) == 0 {
		s.logger.Warn("No healthy chat nodes found in Consul, using fallback")
		return s.getFallbackNodes(), nil
	}

	var nodes []dto.NodeHealthResponse

	for _, service := range healthyServices {
		hostPort := net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))
		nodeAddress := fmt.Sprintf("ws://%s", hostPort)
		nodes = append(nodes, dto.NodeHealthResponse{
			NodeID:         service.Service.ID,
			Address:        nodeAddress,
			Status:         "healthy",
			Connections:    0, // Would be retrieved from actual node metrics
			MaxConnections: s.nodeConfig.MaxConnections,
			LastSeen:       time.Now(),
		})
	}

	return nodes, nil
}

// selectOptimalNode selects the best available chat node for a new connection.
func (s *connectionService) selectOptimalNode(ctx context.Context) string {
	nodes, err := s.GetNodeHealth(ctx)
	if err != nil {
		s.logger.Error("Failed to get node health for selection", "error", err)
		return s.nodeConfig.DefaultNodeAddress
	}

	if len(nodes) == 0 {
		s.logger.Warn("No nodes available, using default")
		return s.nodeConfig.DefaultNodeAddress
	}

	// Simple round-robin selection (in production, this would be more sophisticated)
	// For now, just return the first healthy node
	selectedNode := nodes[0]

	s.logger.Info("Selected chat node",
		"node_id", selectedNode.NodeID,
		"address", selectedNode.Address,
		"available_nodes", len(nodes))

	return selectedNode.Address
}

// getFallbackNodes returns fallback nodes when service discovery is unavailable.
func (s *connectionService) getFallbackNodes() []dto.NodeHealthResponse {
	return []dto.NodeHealthResponse{
		{
			NodeID:         "chat-node-fallback",
			Address:        s.nodeConfig.DefaultNodeAddress,
			Status:         "healthy",
			Connections:    0,
			MaxConnections: s.nodeConfig.MaxConnections,
			LastSeen:       time.Now(),
		},
	}
}
