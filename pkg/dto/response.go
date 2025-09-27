package dto

// WebResponse represents a standard HTTP response structure.
type WebResponse[T any, P any] struct {
	Message    string       `json:"message,omitempty"`
	Data       T            `json:"data,omitempty"`
	Pagination P            `json:"pagination,omitempty"`
	Errors     []FieldError `json:"errors,omitempty"`
}
