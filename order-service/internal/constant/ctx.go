package constant

// ContextKey is a type for context keys.
type ContextKey string

const (
	// CtxOrderIDKey is the context key for the order ID.
	CtxOrderIDKey ContextKey = "order_id"
	// CtxTraceIDKey is the context key for the trace ID.
	CtxTraceIDKey ContextKey = "trace_id"
)
