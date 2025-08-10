package constant

// timeLayoutTranslate returns a map for translating Go time layouts to human-readable formats.
func timeLayoutTranslate() map[string]string {
	return map[string]string{
		"02-01-2006": "DD-MM-YYYY",
		"2006-01-02": "YYYY-MM-DD",
		"2006":       "YYYY",
		"15:04":      "hh:mm",
	}
}

// ConvertGoTimeLayoutToReadable converts Go time layout to a human-readable format.
func ConvertGoTimeLayoutToReadable(layout string) string {
	return timeLayoutTranslate()[layout]
}
