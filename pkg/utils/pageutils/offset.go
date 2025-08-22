package pageutils

// GetOffset calculates the offset for pagination.
func GetOffset(page, limit int64) int64 {
	return limit * (page - 1)
}
