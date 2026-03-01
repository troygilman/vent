package vent

// FieldTransformFunc is a function that transforms a field value before saving to the database.
// It takes the input value and returns the transformed value or an error.
type FieldTransformFunc func(value any) (any, error)
