package constant

// Error messages used throughout the auth service for user and token operations.
const (
	// UserAlreadyExistErrorMessage is returned when a user already exists.
	UserAlreadyExistErrorMessage = "user already exist"
	// InvalidCredentialErrorMessage is returned when token is incorrect or invalid credentials.
	//nolint:gosec // false positive
	InvalidCredentialErrorMessage = "invalid credentials, wrong token or email or password"
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
	// InternalServerErrorMessage is returned for internal server errors.
	InternalServerErrorMessage = "internal server error"
)

// Address error messages used throughout the auth service for address operations.
const (
	// AddressNotFoundErrorMessage is returned when an address is not found.
	AddressNotFoundErrorMessage = "address not found"

	// AddressAccessDeniedErrorMessage is returned when a user tries to access an address they don't own.
	AddressAccessDeniedErrorMessage = "access denied: you do not have permission to access this address"

	// CannotDeleteDefaultAddressErrorMessage is returned when trying to delete the default address.
	CannotDeleteDefaultAddressErrorMessage = "cannot delete default address, please set another address as default first"

	// CannotDeleteLastAddressErrorMessage is returned when trying to delete the only address.
	CannotDeleteLastAddressErrorMessage = "cannot delete your only address"

	// InvalidCoordinatesErrorMessage is returned when coordinates are out of valid range.
	InvalidCoordinatesErrorMessage = "invalid coordinates: latitude must be between -90 and 90, longitude must be between -180 and 180"

	// InvalidCountryCodeErrorMessage is returned when country code is not 2 characters.
	InvalidCountryCodeErrorMessage = "invalid country code: must be a 2-character ISO 3166-1 alpha-2 code"

	// MaxAddressesReachedErrorMessage is returned when user tries to create more addresses than allowed.
	MaxAddressesReachedErrorMessage = "maximum number of addresses reached"
)
