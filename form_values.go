package vent

import (
	"fmt"
	"reflect"
	"time"
)

// FormatFormValue formats an entity field value for HTML form controls.
// Optional/nillable Ent fields are pointers; they are dereferenced before
// formatting so forms show the value instead of a pointer address.
func FormatFormValue(value any) string {
	value = derefFormValue(value)
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02T15:04")
	default:
		return fmt.Sprintf("%v", value)
	}
}

func derefFormValue(value any) any {
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return nil
	}
	return rv.Interface()
}

// ParseDateTimeLocal parses a datetime-local or RFC3339 string.
func ParseDateTimeLocal(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("must use datetime-local or RFC3339 format")
}

// PasswordHashIsSet reports whether a password hash pointer is populated.
func PasswordHashIsSet(hash *string) bool {
	return hash != nil && *hash != ""
}
