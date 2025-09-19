package constant

// ContextKey is a type for context keys.
type ContextKey string

const (
	// CtxOrderIDKey is the context key for the order ID.
	CtxOrderIDKey ContextKey = "order_id"
	// CtxTraceIDKey is the context key for the trace ID.
	CtxTraceIDKey ContextKey = "trace_id"
	// CtxUserIDKey is the context key for the user ID.
	CtxUserIDKey ContextKey = "user_id"
	// CtxUserRoleKey is the context key for the user role.
	CtxUserRoleKey ContextKey = "user_role"
)
