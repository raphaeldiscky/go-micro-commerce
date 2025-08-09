package constant

// Error messages used throughout the auth service for user and token operations.
const (
	// UserAlreadyExistErrorMessage is returned when a user already exists.
	UserAlreadyExistErrorMessage = "user already exist"
	// InvalidCredentialErrorMessage is returned when email or password is incorrect.
	InvalidCredentialErrorMessage = "email or password is incorrect"
	// InvalidRefreshToken is returned when the refresh token is invalid.
	InvalidRefreshToken = "invalid refresh token"
	// UserNotVerifiedErrorMessage is returned when a user has not verified their email.
	UserNotVerifiedErrorMessage = "user not verified"
	// TokenExpiredErrorErrorMessage is returned when a token has expired.
	TokenExpiredErrorErrorMessage = "token expired"
	// UserNotFoundErrorErrorMessage is returned when a user is not found.
	UserNotFoundErrorErrorMessage = "user not found"
	// TokenWrongErrorErrorMessage is returned when a token is incorrect.
	TokenWrongErrorErrorMessage = "token is wrong"
	// UserAlreadyVerifiedErrorMessage is returned when a user is already verified.
	UserAlreadyVerifiedErrorMessage = "user already verified"
	// TokenAlreadyExistErrorMessage is returned when a verification token was recently sent.
	TokenAlreadyExistErrorMessage = "token was recently sent, please wait before"
)
