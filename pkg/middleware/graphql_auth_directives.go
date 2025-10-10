package middleware

import (
	"context"
	"slices"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// RequiresAuthDirective ensures a user is authenticated for GraphQL operations.
// Returns a GraphQL error if user is not authenticated.
// This directive should be applied to queries/mutations that require authentication.
func RequiresAuthDirective(
	ctx context.Context,
	_ any,
	next graphql.Resolver,
) (any, error) {
	// Check if user ID exists in context
	if _, ok := ctx.Value(constant.CtxKeyUserID).(uuid.UUID); !ok {
		return nil, &gqlerror.Error{
			Message: "unauthorized: missing or invalid authentication token",
			Extensions: map[string]any{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	return next(ctx)
}

// RequiresRoleDirective ensures a user has the required role for GraphQL operations.
// Returns a GraphQL error if user doesn't have the required role.
// This directive should be applied to queries/mutations that require specific roles.
func RequiresRoleDirective(
	ctx context.Context,
	_ any,
	next graphql.Resolver,
	role string,
) (any, error) {
	// First, ensure user is authenticated
	userID, ok := ctx.Value(constant.CtxKeyUserID).(uuid.UUID)
	if !ok {
		return nil, &gqlerror.Error{
			Message: "unauthorized: missing or invalid authentication token",
			Extensions: map[string]any{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	// Get user roles from context
	roles, ok := ctx.Value(constant.CtxKeyRoles).([]string)
	if !ok || len(roles) == 0 {
		return nil, &gqlerror.Error{
			Message: "forbidden: user has no roles assigned",
			Extensions: map[string]any{
				"code":    "FORBIDDEN",
				"user_id": userID.String(),
			},
		}
	}

	// Check if user has the required role
	if !slices.Contains(roles, role) {
		return nil, &gqlerror.Error{
			Message: "forbidden: insufficient permissions",
			Extensions: map[string]any{
				"code":          "FORBIDDEN",
				"required_role": role,
				"user_roles":    roles,
			},
		}
	}

	return next(ctx)
}
