package validationutils

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// TagToMsg converts a validator.FieldError into a user-friendly error message.
func TagToMsg(fe validator.FieldError) string {
	tagMsgMap := map[string]func(fe validator.FieldError) string{
		"required": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s is required", fe.Field())
		},
		"len": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s length or value must be exactly %v", fe.Field(), fe.Param())
		},
		"max": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s length or value %v must be at most", fe.Field(), fe.Param())
		},
		"dgte": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be greater than or equal to %v", fe.Field(), fe.Param())
		},
		"dlte": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be less than or equal to %v", fe.Field(), fe.Param())
		},
		"dgt": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be greater than to %v", fe.Field(), fe.Param())
		},
		"dlt": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be less than to %v", fe.Field(), fe.Param())
		},
		"gte": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be greater than or equal to %v", fe.Field(), fe.Param())
		},
		"lte": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be lower than or equal to %v", fe.Field(), fe.Param())
		},
		"email": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s has invalid email format", fe.Field())
		},
		"eq": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be equal to %v", fe.Field(), fe.Param())
		},
		"min": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s length or value must be at least %v", fe.Field(), fe.Param())
		},
		"numeric": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be a number", fe.Field())
		},
		"boolean": func(fe validator.FieldError) string {
			return fmt.Sprintf("%s must be a boolean", fe.Field())
		},
	}

	switch fe.Tag() {
	case "gtefield":
		return fmt.Sprintf("%s must be greater than or equal to %v", fe.StructField(), fe.Param())
	case "time_format":
		return fmt.Sprintf(
			"please send time in format of %s",
			constant.ConvertGoTimeLayoutToReadable(fe.Param()),
		)
	default:
		if fn, ok := tagMsgMap[fe.Tag()]; ok {
			return fn(fe)
		}

		return "invalid input"
	}
}
