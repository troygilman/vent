package vent

// FieldKind is the supported set of admin form field render kinds.
type FieldKind string

const (
	FieldKindString           FieldKind = "string"
	FieldKindPassword         FieldKind = "password"
	FieldKindInt              FieldKind = "int"
	FieldKindFloat            FieldKind = "float"
	FieldKindBool             FieldKind = "bool"
	FieldKindForeignKey       FieldKind = "foreign_key"
	FieldKindForeignKeyUnique FieldKind = "foreign_key_unique"
	FieldKindTime             FieldKind = "time"
)

// FieldKindFromString normalizes a string into a supported FieldKind.
func FieldKindFromString(value string) (FieldKind, bool) {
	switch value {
	case string(FieldKindString):
		return FieldKindString, true
	case string(FieldKindPassword):
		return FieldKindPassword, true
	case string(FieldKindInt), "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return FieldKindInt, true
	case string(FieldKindFloat), "float32", "float64":
		return FieldKindFloat, true
	case string(FieldKindBool):
		return FieldKindBool, true
	case string(FieldKindForeignKey):
		return FieldKindForeignKey, true
	case string(FieldKindForeignKeyUnique):
		return FieldKindForeignKeyUnique, true
	case string(FieldKindTime), "time.Time":
		return FieldKindTime, true
	default:
		return "", false
	}
}
