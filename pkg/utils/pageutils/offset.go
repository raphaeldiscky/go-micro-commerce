package pageutils

// GetOffset calculates the offset for pagination.
func GetOffset(page, limit int64) int64 {
	if page < 1 {
		page = 1
	}

	return limit * (page - 1)
}

// GetPageFromOffset calculates the page number from offset and limit.
func GetPageFromOffset(offset, limit int64) int64 {
	if limit <= 0 {
		return 1
	}

	return (offset / limit) + 1
}

// HasNextPage determines if there's a next page for offset pagination.
func HasNextPage(page, limit, totalItems int64) bool {
	return (page * limit) < totalItems
}
