// Package grpc provides gRPC client interceptors for adding authentication headers.
package grpc

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// ClientAuthInterceptor provides client-side authentication interceptor.
type ClientAuthInterceptor struct{}

// NewClientAuthInterceptor creates a new client authentication interceptor.
func NewClientAuthInterceptor() *ClientAuthInterceptor {
	return &ClientAuthInterceptor{}
}

// ForwardUserAuth creates a unary client interceptor that forwards user authentication headers.
func (c *ClientAuthInterceptor) ForwardUserAuth() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Extract user information from context and forward as metadata
		newCtx := c.addUserInfoToMetadata(ctx)

		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// addUserInfoToMetadata extracts user info from context and adds it to gRPC metadata.
func (c *ClientAuthInterceptor) addUserInfoToMetadata(ctx context.Context) context.Context {
	mdMap := make(map[string]string)

	// Extract user information from context
	if userID, ok := ctx.Value(constant.CtxKeyUserID).(uuid.UUID); ok {
		mdMap[strings.ToLower(constant.XUserID)] = userID.String()
	}

	if email, ok := ctx.Value(constant.CtxKeyEmail).(string); ok {
		mdMap[strings.ToLower(constant.XEmail)] = email
	}

	if roles, ok := ctx.Value(constant.CtxKeyRoles).([]string); ok {
		mdMap[strings.ToLower(constant.XRoles)] = strings.Join(roles, ",")
	}

	if isActive, ok := ctx.Value(constant.CtxKeyIsActive).(bool); ok {
		mdMap[strings.ToLower(constant.XIsActive)] = strconv.FormatBool(isActive)
	}

	// Create metadata from map
	md := metadata.New(mdMap)

	// If there's existing metadata, merge it
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
