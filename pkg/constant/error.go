// Package constant defines error messages used across the application.
package constant

const (
	// InternalServerErrorMessage indicates an unexpected server error.
	InternalServerErrorMessage = "currently our server is facing unexpected error, please try again later"
	// EOFErrorMessage indicates that the request body is missing.
	EOFErrorMessage = "missing body request"
	// JSONSyntaxErrorMessage indicates that the JSON syntax is invalid.
	JSONSyntaxErrorMessage = "invalid JSON syntax"
	// JSONUnMarshallTypeErrorMessage indicates that the JSON value could not be unmarshalled into the expected type.
	JSONUnMarshallTypeErrorMessage = "invalid value for %s"
	// UnauthorizedErrorMessage indicates that the request requires authentication.
	UnauthorizedErrorMessage = "unauthorized"
	// RequestTimeoutErrorMessage indicates that the request took too long to process.
	RequestTimeoutErrorMessage = "failed to process request in time, please try again"
	// ValidationErrorMessage indicates that the input validation failed.
	ValidationErrorMessage = "input validation error"
	// InvalidURLParamErrorMessage indicates that the URL parameter is not valid.
	InvalidURLParamErrorMessage = "expected a numeric value but got '%s'"
	// RequestDuplicateErrorMessage indicates that the request is a duplicate.
	RequestDuplicateErrorMessage = "request duplicate"
)
