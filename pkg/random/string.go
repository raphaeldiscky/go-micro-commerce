package random

const (
	// Character sets for string generation.
	alphaLower = "abcdefghijklmnopqrstuvwxyz"
	alphaUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits     = "0123456789"
	alphaNum   = alphaLower + alphaUpper + digits
)

// String generates a random string of specified length using alphanumeric characters.
func String(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	for i := range result {
		idx := Int(int64(len(alphaNum)))
		result[i] = alphaNum[idx]
	}

	return string(result)
}

// StringWithCharset generates a random string of specified length using the provided charset.
func StringWithCharset(length int, charset string) string {
	if length <= 0 || charset == "" {
		return ""
	}

	result := make([]byte, length)
	for i := range result {
		idx := Int(int64(len(charset)))
		result[i] = charset[idx]
	}

	return string(result)
}

// AlphaString generates a random alphabetic string (a-z, A-Z) of specified length.
func AlphaString(length int) string {
	return StringWithCharset(length, alphaLower+alphaUpper)
}

// NumericString generates a random numeric string (0-9) of specified length.
func NumericString(length int) string {
	return StringWithCharset(length, digits)
}

// Choice returns a random element from the provided slice.
// Returns zero value if slice is empty.
func Choice(items []string) string {
	if len(items) == 0 {
		return ""
	}

	idx := Int(int64(len(items)))

	return items[idx]
}

// Shuffle randomly shuffles a slice of strings in place.
func Shuffle(items []string) {
	for i := len(items) - 1; i > 0; i-- {
		j := Int(int64(i + 1))
		items[i], items[j] = items[j], items[i]
	}
}
