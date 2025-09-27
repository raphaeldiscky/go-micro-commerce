package pageutils

import (
	"math"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// NewOffsetPagination creates a new OffsetPagination instance.
func NewOffsetPagination(count, page, size int64) *dto.OffsetPagination {
	totalItems := count
	totalPage := int64(math.Ceil(float64(totalItems) / float64(size)))

	return &dto.OffsetPagination{
		Page:      page,
		Size:      size,
		TotalItem: totalItems,
		TotalPage: totalPage,
	}
}

// NewCursorPagination creates a new CursorPagination instance.
func NewCursorPagination(
	nextCursor, prevCursor string,
	hasNext, hasPrev bool,
	limit int64,
) *dto.CursorPagination {
	return &dto.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      limit,
	}
}
