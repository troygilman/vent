package vent

import (
	"fmt"
	"strconv"
	"time"
)

// FieldType represents the type of a field value
type FieldType int

const (
	TypeString FieldType = iota
	TypeInt
	TypeBool
	TypeTime
	TypeForeignKey
)

func (t FieldType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeBool:
		return "bool"
	case TypeTime:
		return "time"
	case TypeForeignKey:
		return "foreign_key"
	default:
		return "unknown"
	}
}

// RelationValue holds information about a foreign key relationship
type RelationValue struct {
	TargetSchema string // The name of the related schema (e.g., "User")
	TargetID     int    // The ID of the related entity
	TargetLabel  string // Display value of the related entity (e.g., "alice@example.com")
	TargetPath   string // URL path to the related entity's admin page
}

// FieldValue represents a rich field value that includes type information,
// the raw value, a display string, and optional relation metadata
type FieldValue struct {
	Type     FieldType      // The type of this field
	Raw      any            // The actual value (int, string, bool, time.Time, etc.)
	Display  string         // Human-readable display value
	Relation *RelationValue // Non-nil if this is a foreign key field
}

// NewStringFieldValue creates a FieldValue for a string
func NewStringFieldValue(v string) FieldValue {
	return FieldValue{
		Type:    TypeString,
		Raw:     v,
		Display: v,
	}
}

// NewIntFieldValue creates a FieldValue for an int
func NewIntFieldValue(v int) FieldValue {
	return FieldValue{
		Type:    TypeInt,
		Raw:     v,
		Display: strconv.Itoa(v),
	}
}

// NewBoolFieldValue creates a FieldValue for a bool
func NewBoolFieldValue(v bool) FieldValue {
	return FieldValue{
		Type:    TypeBool,
		Raw:     v,
		Display: strconv.FormatBool(v),
	}
}

// NewTimeFieldValue creates a FieldValue for a time.Time
func NewTimeFieldValue(v time.Time) FieldValue {
	return FieldValue{
		Type:    TypeTime,
		Raw:     v,
		Display: v.Format(time.RFC3339),
	}
}

// NewForeignKeyFieldValue creates a FieldValue for a foreign key relationship
func NewForeignKeyFieldValue(id int, relation RelationValue) FieldValue {
	return FieldValue{
		Type:     TypeForeignKey,
		Raw:      id,
		Display:  relation.TargetLabel,
		Relation: &relation,
	}
}

// String returns the display value of the field
func (f FieldValue) String() string {
	return f.Display
}

// IsZero returns true if the field value is unset/zero
func (f FieldValue) IsZero() bool {
	if f.Raw == nil {
		return true
	}
	switch f.Type {
	case TypeString:
		return f.Raw.(string) == ""
	case TypeInt:
		return f.Raw.(int) == 0
	case TypeBool:
		return !f.Raw.(bool)
	case TypeTime:
		return f.Raw.(time.Time).IsZero()
	case TypeForeignKey:
		return f.Raw.(int) == 0
	default:
		return true
	}
}

// StringValue returns the raw value as a string, or panics if not a string type
func (f FieldValue) StringValue() string {
	if f.Type != TypeString {
		panic(fmt.Sprintf("vent: cannot get string value from field of type %s", f.Type))
	}
	return f.Raw.(string)
}

// IntValue returns the raw value as an int, or panics if not an int type
func (f FieldValue) IntValue() int {
	if f.Type != TypeInt && f.Type != TypeForeignKey {
		panic(fmt.Sprintf("vent: cannot get int value from field of type %s", f.Type))
	}
	return f.Raw.(int)
}

// BoolValue returns the raw value as a bool, or panics if not a bool type
func (f FieldValue) BoolValue() bool {
	if f.Type != TypeBool {
		panic(fmt.Sprintf("vent: cannot get bool value from field of type %s", f.Type))
	}
	return f.Raw.(bool)
}

// TimeValue returns the raw value as a time.Time, or panics if not a time type
func (f FieldValue) TimeValue() time.Time {
	if f.Type != TypeTime {
		panic(fmt.Sprintf("vent: cannot get time value from field of type %s", f.Type))
	}
	return f.Raw.(time.Time)
}

// EntityData represents a single entity as a map of field names to their values
type EntityData map[string]FieldValue

// ID returns the "id" field value as an int, or 0 if not present
func (e EntityData) ID() int {
	if field, ok := e["id"]; ok {
		return field.IntValue()
	}
	return 0
}

// Get returns the FieldValue for the given field name and a boolean indicating if it exists
func (e EntityData) Get(name string) (FieldValue, bool) {
	field, ok := e[name]
	return field, ok
}

// GetString returns the string value of a field, or empty string if not found
func (e EntityData) GetString(name string) string {
	if field, ok := e[name]; ok && field.Type == TypeString {
		return field.StringValue()
	}
	return ""
}

// GetInt returns the int value of a field, or 0 if not found
func (e EntityData) GetInt(name string) int {
	if field, ok := e[name]; ok && (field.Type == TypeInt || field.Type == TypeForeignKey) {
		return field.IntValue()
	}
	return 0
}

// GetBool returns the bool value of a field, or false if not found
func (e EntityData) GetBool(name string) bool {
	if field, ok := e[name]; ok && field.Type == TypeBool {
		return field.BoolValue()
	}
	return false
}

// GetTime returns the time value of a field, or zero time if not found
func (e EntityData) GetTime(name string) time.Time {
	if field, ok := e[name]; ok && field.Type == TypeTime {
		return field.TimeValue()
	}
	return time.Time{}
}
