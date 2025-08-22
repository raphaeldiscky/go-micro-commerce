package echoutils

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/dto"
)

// ResponseOK sends a 200 OK response.
func ResponseOK[T any](ctx echo.Context, data T) error {
	return ResponseJSON(ctx, http.StatusOK, constant.ResponseSuccessMessage, data, nil)
}

// ResponseOKPlain sends a 200 OK response with no content.
func ResponseOKPlain(ctx echo.Context) error {
	return ResponseOK[any](ctx, nil)
}

// ResponseOKPagination sends a 200 OK response with pagination metadata.
func ResponseOKPagination[T any](ctx echo.Context, data T, paging *dto.PageMetaData) error {
	return ResponseJSON(ctx, http.StatusOK, constant.ResponseSuccessMessage, data, paging)
}

// ResponseCreated sends a 201 Created response.
func ResponseCreated[T any](ctx echo.Context, data T) error {
	return ResponseJSON(ctx, http.StatusCreated, constant.ResponseSuccessMessage, data, nil)
}

// ResponseCreatedPlain sends a 201 Created response with no content.
func ResponseCreatedPlain(ctx echo.Context) error {
	return ResponseCreated[any](ctx, nil)
}

// ResponseJSON sends a JSON response.
func ResponseJSON[T any](
	ctx echo.Context,
	statusCode int,
	message string,
	data T,
	paging *dto.PageMetaData,
) error {
	return ctx.JSON(statusCode, dto.WebResponse[T]{
		Message:    message,
		Data:       data,
		Pagination: paging,
	})
}
