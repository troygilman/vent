package utils

import (
	"encoding/json"
)

func Stringify(value any, typeName string) string {
	if typeName == "string" {
		return value.(string)
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(buf)
}
