package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/dto"
	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/validationutils"
)

// ErrorHandler is a middleware function that handles errors.
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			// Log error with file + line
			log.Printf("Error occurred at %s: %+v\n",
				fileLine(3), // skip 3 levels to point to where error originated
				err,
			)

			// Handle different error types
			{
				var e validator.ValidationErrors

				var e1 *json.SyntaxError

				var e2 *json.UnmarshalTypeError

				var e3 *time.ParseError

				var e4 *httperror.ResponseError

				var e6 *echo.HTTPError

				switch {
				case errors.As(err, &e):
					return handleValidationError(c, e)
				case errors.As(err, &e1):
					return handleJSONSyntaxError(c)
				case errors.As(err, &e2):
					return handleJSONUnmarshalTypeError(c, e2)
				case errors.As(err, &e3):
					return handleParseTimeError(c, e3)
				case errors.As(err, &e4):
					return c.JSON(e4.GetCode(), dto.WebResponse[any]{Message: e4.DisplayMessage()})
				case isUUIDError(err):
					return c.JSON(http.StatusBadRequest, dto.WebResponse[any]{
						Message: "invalid UUID format",
					})
				case errors.As(err, &e6):
					code := e6.Code

					message, ok := e6.Message.(string)
					if !ok {
						message = constant.InternalServerErrorMessage
					}

					if code == http.StatusBadRequest {
						if message == "EOF" {
							message = constant.EOFErrorMessage
						} else {
							message = constant.JSONSyntaxErrorMessage
						}
					}

					return c.JSON(code, dto.WebResponse[any]{Message: message})
				default:
					if errors.Is(err, io.EOF) {
						return c.JSON(
							http.StatusBadRequest,
							dto.WebResponse[any]{Message: constant.EOFErrorMessage},
						)
					}

					log.Printf("Unexpected error at %s: %+v\n", fileLine(3), err)

					return c.JSON(
						http.StatusInternalServerError,
						dto.WebResponse[any]{Message: constant.InternalServerErrorMessage},
					)
				}
			}
		}
	}
}

// helper to get file + line + function.
func fileLine(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}

	fn := runtime.FuncForPC(pc).Name()

	return fmt.Sprintf("%s:%d (%s)", file, line, fn)
}

// handleJSONSyntaxError handles JSON syntax errors.
func handleJSONSyntaxError(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, dto.WebResponse[any]{
		Message: constant.JSONSyntaxErrorMessage,
	})
}

// handleJSONUnmarshalTypeError handles JSON unmarshal type errors.
func handleJSONUnmarshalTypeError(c echo.Context, err *json.UnmarshalTypeError) error {
	return c.JSON(http.StatusBadRequest, dto.WebResponse[any]{
		Message: "invalid type for field " + err.Field,
	})
}

// handleParseTimeError handles parse time errors.
func handleParseTimeError(c echo.Context, err *time.ParseError) error {
	return c.JSON(http.StatusBadRequest, dto.WebResponse[any]{
		Message: "please send time in format of " +
			constant.ConvertGoTimeLayoutToReadable(err.Layout) +
			", got: " + err.Value,
	})
}

// handleValidationError handles validation errors.
func handleValidationError(c echo.Context, err validator.ValidationErrors) error {
	ve := []dto.FieldError{}

	for _, fe := range err {
		ve = append(ve, dto.FieldError{
			Field:   fe.Field(),
			Message: validationutils.TagToMsg(fe),
		})
	}

	return c.JSON(http.StatusBadRequest, dto.WebResponse[any]{
		Message: constant.ValidationErrorMessage,
		Errors:  ve,
	})
}

// isUUIDError checks if the error is from uuid.Parse.
func isUUIDError(err error) bool {
	// Check if error message contains UUID-related keywords
	errMsg := err.Error()

	return strings.Contains(errMsg, "invalid UUID") ||
		strings.Contains(errMsg, "UUID") ||
		strings.Contains(errMsg, "invalid length")
}
