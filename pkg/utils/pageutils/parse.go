package pageutils

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// ParseQueryInt64 parses a query parameter into int64 with default, min, max.
func ParseQueryInt64(c echo.Context, key string, def, minValue, maxValue int64) int64 {
	valStr := c.QueryParam(key)
	if valStr == "" {
		return def
	}

	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil || val < minValue {
		return def
	}

	if maxValue > 0 && val > maxValue {
		return maxValue
	}

	return val
}
