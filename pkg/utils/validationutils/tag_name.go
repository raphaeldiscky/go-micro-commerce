package validationutils

import (
	"reflect"
	"strings"
)

const (
	defaultSplitN = 2
)

// TagNameFormatter formats the field name from struct tags.
func TagNameFormatter(fld *reflect.StructField) string {
	var name string

	jsonTag := fld.Tag.Get("json")
	formTag := fld.Tag.Get("form")

	if jsonTag != "" {
		name = strings.SplitN(jsonTag, ",", defaultSplitN)[0]
	} else if formTag != "" {
		name = strings.SplitN(formTag, ",", defaultSplitN)[0]
	}

	if name == "-" {
		return ""
	}

	return name
}
