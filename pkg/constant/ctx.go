package constant

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// CtxKeyUserID is the context key for the user ID.
	CtxKeyUserID ContextKey = "user_id"
	// CtxKeyEmail is the context key for the user email.
	CtxKeyEmail ContextKey = "email"
	// CtxKeyRoles is the context key for the user roles.
	CtxKeyRoles ContextKey = "roles"
	// CtxKeyIsActive is the context key for the user active status.
	CtxKeyIsActive ContextKey = "is_active"
	// CtxKeyClientIP is the context key for the client IP.
	CtxKeyClientIP ContextKey = "client_ip"
	// CtxKeyUserAgent is the context key for the user agent.
	CtxKeyUserAgent ContextKey = "user_agent"
	// CtxKeyResponseWriter is the context key for the HTTP response writer.
	CtxKeyResponseWriter ContextKey = "response_writer"
)

const (
	// BearerPrefix is the prefix for Bearer token in Authorization header.
	BearerPrefix = "Bearer"
)
