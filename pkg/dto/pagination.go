package dto

// OffsetPagination represents metadata about the current page of results.
type OffsetPagination struct {
	Page      int64 `json:"page"`
	Size      int64 `json:"size"`
	TotalItem int64 `json:"total_item"`
	TotalPage int64 `json:"total_page"`
}

// CursorPagination represents metadata about the current page of results.
type CursorPagination struct {
	NextCursor string `json:"next_cursor"`
	PrevCursor string `json:"prev_cursor"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	Limit      int64  `json:"limit"`
}
