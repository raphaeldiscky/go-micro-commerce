package constant

const (
	// XUserID is the header for the user ID.
	XUserID = "X-User-ID"
	// XEmail is the header for the user email.
	XEmail = "X-Email"
	// XRoles is the header for the user roles.
	XRoles = "X-Roles"
	// XIsActive is the header for the user active status.
	XIsActive = "X-Is-Active"
	// XRequestID is the header for the request ID, automatically generated for each request from middleware.
	XRequestID = "X-Request-Id"
)

const (
	// RoleAdmin is the role for administrators.
	RoleAdmin = "admin"
	// RoleUser is the role for regular users (default).
	RoleUser = "user"
)
