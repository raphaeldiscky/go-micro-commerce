package constant

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// CtxUserID is the context key for the user ID.
	CtxUserID ContextKey = "user_id"
	// CtxEmail is the context key for the user email.
	CtxEmail ContextKey = "email"
	// CtxRoles is the context key for the user roles.
	CtxRoles ContextKey = "roles"
	// CtxIsActive is the context key for the user active status.
	CtxIsActive ContextKey = "is_active"
)
