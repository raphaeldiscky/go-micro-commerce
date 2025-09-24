// Package service provides business logic for chat connection management.
package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/httperror"
)

// ConnectionService defines the interface for chat connection management.
type ConnectionService interface {
	// RequestConnection generates a connection ticket and selects optimal chat node
	RequestConnection(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) (*dto.ChatConnectionResponse, error)

	// ValidateConnectionTicket validates a connection ticket and extracts user information
	ValidateConnectionTicket(
		ctx context.Context,
		ticket string,
	) (*dto.ConnectionTicketClaims, error)

	// GetNodeHealth returns health status of available chat nodes
	GetNodeHealth(ctx context.Context) ([]dto.NodeHealthResponse, error)
}

// connectionService implements the ConnectionService interface.
type connectionService struct {
	logger       logger.Logger
	jwtSecret    string
	nodeConfig   *NodeConfig
	consulClient *api.Client
}

// NodeConfig holds configuration for chat node selection.
type NodeConfig struct {
	DefaultNodeAddress string
	TicketExpiration   time.Duration
	MaxConnections     int
	ConsulAddress      string
	ChatServiceName    string
}

// NewConnectionService creates a new instance of connectionService.
func NewConnectionService(
	appLogger logger.Logger,
	jwtSecret string,
	nodeConfig *NodeConfig,
) ConnectionService {
	// Initialize Consul client for service discovery
	var consulClient *api.Client
	if nodeConfig.ConsulAddress != "" {
		consulConfig := api.DefaultConfig()
		consulConfig.Address = nodeConfig.ConsulAddress

		client, err := api.NewClient(consulConfig)
		if err != nil {
			appLogger.Warn("Failed to initialize Consul client, using fallback", "error", err)
		} else {
			consulClient = client
		}
	}

	return &connectionService{
		logger:       appLogger,
		jwtSecret:    jwtSecret,
		nodeConfig:   nodeConfig,
		consulClient: consulClient,
	}
}

// RequestConnection generates a connection ticket and selects optimal chat node.
func (s *connectionService) RequestConnection(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) (*dto.ChatConnectionResponse, error) {
	// For now, use a simple node selection strategy
	// In production, this would query service discovery and load balance
	selectedNode := s.selectOptimalNode(ctx)

	// Generate connection ticket
	ticket, expiresAt, err := s.generateConnectionticket(userID, userType)
	if err != nil {
		s.logger.Error("Failed to generate connection ticket", "error", err, "user_id", userID)
		return nil, httperror.NewInternalServerError("failed to generate connection ticket")
	}

	s.logger.Info("Generated connection ticket",
		"user_id", userID,
		"user_type", userType,
		"node_address", selectedNode,
		"expires_at", expiresAt)

	return &dto.ChatConnectionResponse{
		NodeAddress: selectedNode,
		Ticket:      ticket,
		ExpiresAt:   expiresAt,
		UserID:      userID,
		UserType:    userType,
	}, nil
}

// ValidateConnectionTicket validates a connection ticket and extracts user information.
func (s *connectionService) ValidateConnectionTicket(
	_ context.Context,
	ticket string,
) (*dto.ConnectionTicketClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(
		ticket,
		&dto.ConnectionTicketClaims{},
		func(jwtToken *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", jwtToken.Header["alg"])
			}

			return []byte(s.jwtSecret), nil
		},
	)
	if err != nil {
		s.logger.Warn("Invalid connection ticket", "error", err)
		return nil, httperror.NewUnauthorizedError("invalid connection ticket")
	}

	claims, ok := jwtToken.Claims.(*dto.ConnectionTicketClaims)
	if !ok || !jwtToken.Valid {
		s.logger.Warn("Invalid ticket claims")
		return nil, httperror.NewUnauthorizedError("invalid ticket claims")
	}

	// Check if ticket has expired
	if time.Now().After(claims.ExpiresAt.Time) {
		s.logger.Warn(
			"Connection ticket expired",
			"user_id",
			claims.UserID,
			"expired_at",
			claims.ExpiresAt.Time,
		)

		return nil, httperror.NewUnauthorizedError("connection ticket expired")
	}

	s.logger.Info("Successfully validated connection ticket",
		"user_id", claims.UserID,
		"user_type", claims.UserType)

	return claims, nil
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
		serviceName = "chat-service-websocket" // Default service name
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

// generateConnectionticket creates a signed JWT ticket for WebSocket connection.
func (s *connectionService) generateConnectionticket(
	userID uuid.UUID,
	userType constant.UserType,
) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.nodeConfig.TicketExpiration)

	claims := &dto.ConnectionTicketClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "chat-service",
			Subject:   userID.String(),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ticketString, err := jwtToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign connection ticket: %w", err)
	}

	return ticketString, expiresAt, nil
}
