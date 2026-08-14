package vent

// BoolFilter is a list-filter selector for boolean fields.
type BoolFilter string

const (
	BoolFilterAll   BoolFilter = ""
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
