package utils

import (
	"encoding/json"
	"fmt"
)

func Stringify(value any, typeName string) string {
	switch typeName {
	case "string":
		return value.(string)
	default:
		buf, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return string(buf)
	}
}

func Decode(value any, typeName string) any {
	fmt.Printf("%+T\n", value)
	return value
}
