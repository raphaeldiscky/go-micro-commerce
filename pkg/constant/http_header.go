package constant

const (
	// XUserID is the header for the user ID.
	XUserID = "X-User-ID"
	// XEmail is the header for the user email.
	XEmail = "X-Email"
	// XRoles is the header for the user roles.
	XRoles = "X-Roles"
	// XRequestID is the header for the request ID, automatically generated for each request from middleware.
	XRequestID = "X-Request-Id"
	// XClientIP is the header for the client IP address.
	XClientIP = "X-Client-IP"
	// XUserAgent is the header for the client user agent.
	XUserAgent = "X-User-Agent"
)

const (
	// RoleAdmin is the role for administrators.
	RoleAdmin = "admin"
	// RoleUser is the role for regular users (default).
	RoleUser = "user"
)
