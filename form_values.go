package vent

import (
	"errors"
	"fmt"
	"time"
)

// FormatFormValue formats an entity field value for HTML form controls.
func FormatFormValue(value any) string {
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

// ParseDateTimeLocal parses a datetime-local or RFC3339 string.
func ParseDateTimeLocal(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("value is required")
	}
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("must use datetime-local or RFC3339 format")
}

// ParseDateTimeLocalField parses a datetime-local value with a field label.
func ParseDateTimeLocalField(value string, label string) (time.Time, error) {
	parsed, err := ParseDateTimeLocal(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s: %w", label, err)
	}
	return parsed, nil
}

// PasswordHashIsSet reports whether a password hash pointer is populated.
func PasswordHashIsSet(hash *string) bool {
	return hash != nil && *hash != ""
}
