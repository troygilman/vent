package vent

import (
	"encoding/json"
	"net/url"
	"strings"
)

// UnmarshalFilterQuery reads dotted filter.* query parameters into dest.
// dest should be a pointer to a struct with a `json:"filter"` field.
func UnmarshalFilterQuery(values url.Values, dest any) error {
	filter := map[string]string{}
	for key, vals := range values {
		if !strings.HasPrefix(key, "filter.") || len(vals) == 0 {
			continue
		}
		filter[strings.TrimPrefix(key, "filter.")] = vals[0]
	}
	body, err := json.Marshal(map[string]any{"filter": filter})
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

// BoolFilter is a list-filter selector for boolean fields.
type BoolFilter string

const (
	BoolFilterAll   BoolFilter = "all"
	BoolFilterTrue  BoolFilter = "true"
	BoolFilterFalse BoolFilter = "false"
)

// Bool converts a boolean list-filter signal into a bool.
// ok is false for BoolFilterAll and any other unset value.
func (f BoolFilter) Bool() (v bool, ok bool) {
	switch f {
	case BoolFilterTrue:
		return true, true
	case BoolFilterFalse:
		return false, true
	default:
		return false, false
	}
}

// Normalize returns BoolFilterTrue or BoolFilterFalse unchanged,
// and BoolFilterAll for every other value.
func (f BoolFilter) Normalize() BoolFilter {
	if f == BoolFilterTrue || f == BoolFilterFalse {
		return f
	}
	return BoolFilterAll
}

func (f BoolFilter) String() string {
	return string(f)
}
