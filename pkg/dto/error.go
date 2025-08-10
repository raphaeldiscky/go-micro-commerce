// Package dto provides data transfer objects for the application.
package dto

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
