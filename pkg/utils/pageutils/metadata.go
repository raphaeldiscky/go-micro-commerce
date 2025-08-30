package pageutils

import (
	"math"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// NewMetadata creates a new PageMetaData instance.
func NewMetadata(count, page, limit int64) *dto.PageMetaData {
	totalItems := count
	totalPage := int64(math.Ceil(float64(totalItems) / float64(limit)))

	return &dto.PageMetaData{
		Page:      page,
		Size:      limit,
		TotalItem: totalItems,
		TotalPage: totalPage,
	}
}
