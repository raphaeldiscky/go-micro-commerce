package constant

// ContextKey is a type for context keys.
type ContextKey string

const (
	// CtxCartIDKey is the context key for the cart ID.
	CtxCartIDKey ContextKey = "cart_id"
	// CtxTraceIDKey is the context key for the trace ID.
	CtxTraceIDKey ContextKey = "trace_id"
	// CtxUserIDKey is the context key for the user ID.
	CtxUserIDKey ContextKey = "user_id"
	// CtxUserRoleKey is the context key for the user role.
	CtxUserRoleKey ContextKey = "user_role"
)
