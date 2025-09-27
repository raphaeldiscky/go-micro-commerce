package echoutils

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// ResponseOK sends a 200 OK response.
func ResponseOK[T any](ctx echo.Context, data T) error {
	return ResponseJSON[T, any](ctx, http.StatusOK, constant.ResponseSuccessMessage, data, nil)
}

// ResponseOKPlain sends a 200 OK response with no content.
func ResponseOKPlain(ctx echo.Context) error {
	return ResponseOK[any](ctx, nil)
}

// ResponseOKOffsetPagination sends a 200 OK response with pagination metadata.
func ResponseOKOffsetPagination[T any](
	ctx echo.Context,
	data T,
	pagination *dto.OffsetPagination,
) error {
	return ResponseJSON(ctx, http.StatusOK, constant.ResponseSuccessMessage, data, pagination)
}

// ResponseOKCursorPagination sends a 200 OK response with pagination metadata.
func ResponseOKCursorPagination[T any](
	ctx echo.Context,
	data T,
	pagination *dto.CursorPagination,
) error {
	return ResponseJSON(ctx, http.StatusOK, constant.ResponseSuccessMessage, data, pagination)
}

// ResponseCreated sends a 201 Created response.
func ResponseCreated[T any](ctx echo.Context, data T) error {
	return ResponseJSON[T, any](ctx, http.StatusCreated, constant.ResponseSuccessMessage, data, nil)
}

// ResponseCreatedPlain sends a 201 Created response with no content.
func ResponseCreatedPlain(ctx echo.Context) error {
	return ResponseCreated[any](ctx, nil)
}

// ResponseJSON sends a JSON response.
func ResponseJSON[T any, P any](
	ctx echo.Context,
	statusCode int,
	message string,
	data T,
	pagination P,
) error {
	return ctx.JSON(statusCode, dto.WebResponse[T, P]{
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}
