package vent

import (
	"fmt"
	"strings"
)

// FilterBool is a tri-state boolean list filter: unset (all), true, or false.
//
// Datastar coerces null signals to empty strings when they are read, so JSON
// may arrive as true/false, "true"/"false", null, or "".
type FilterBool struct {
	set bool
	val bool
}

// UnmarshalJSON accepts JSON booleans, null, and the string forms Datastar sends.
func (f *FilterBool) UnmarshalJSON(data []byte) error {
	switch strings.TrimSpace(string(data)) {
	case "null", `""`, "":
		*f = FilterBool{}
		return nil
	case "true", `"true"`:
		*f = FilterBool{set: true, val: true}
		return nil
	case "false", `"false"`:
		*f = FilterBool{set: true, val: false}
		return nil
	default:
		return fmt.Errorf("invalid boolean filter %s", data)
	}
}

// MarshalJSON writes null when unset, otherwise a JSON boolean.
func (f FilterBool) MarshalJSON() ([]byte, error) {
	if !f.set {
		return []byte("null"), nil
	}
	if f.val {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// Ptr returns nil when the filter is unset.
func (f FilterBool) Ptr() *bool {
	if !f.set {
		return nil
	}
	v := f.val
	return &v
}
