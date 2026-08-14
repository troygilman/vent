package vent

// Boolean list-filter selector values sent as Datastar string signals.
const (
	BoolFilterAll   = "all"
	BoolFilterTrue  = "true"
	BoolFilterFalse = "false"
)

// ParseBoolFilter converts a boolean list-filter signal into a bool.
// ok is false for BoolFilterAll and any other unset value.
func ParseBoolFilter(value string) (v bool, ok bool) {
	switch value {
	case BoolFilterTrue:
		return true, true
	case BoolFilterFalse:
		return false, true
	default:
		return false, false
	}
}

// NormalizeBoolFilter returns BoolFilterTrue or BoolFilterFalse unchanged,
// and BoolFilterAll for every other value.
func NormalizeBoolFilter(value string) string {
	if value == BoolFilterTrue || value == BoolFilterFalse {
		return value
	}
	return BoolFilterAll
}
