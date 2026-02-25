package vent

import (
	"fmt"

	"github.com/troygilman/vent/auth"
)

// FieldMapper transforms form data before it's sent to the database for
// create/update operations. Each mapper receives the mutable data map and
// can rename keys, transform values, add/remove entries, or return an error
// to abort the operation.
//
// Mappers are invoked sequentially — the output of one feeds into the next.
// Use [ChainFieldMappers] to compose multiple mappers into one.
type FieldMapper func(data map[string]any) error

// ChainFieldMappers composes multiple FieldMappers into a single FieldMapper.
// They execute in order; the first error short-circuits the chain.
func ChainFieldMappers(mappers ...FieldMapper) FieldMapper {
	return func(data map[string]any) error {
		for _, m := range mappers {
			if err := m(data); err != nil {
				return err
			}
		}
		return nil
	}
}

// MapField reads the value at key "from", applies the transform function,
// writes the result to key "to", and removes "from" if it differs from "to".
// If "from" is not present in the data, the mapper is a no-op.
func MapField(from, to string, transform func(any) (any, error)) FieldMapper {
	return func(data map[string]any) error {
		v, ok := data[from]
		if !ok {
			return nil
		}
		result, err := transform(v)
		if err != nil {
			return fmt.Errorf("field mapper %q -> %q: %w", from, to, err)
		}
		if from != to {
			delete(data, from)
		}
		data[to] = result
		return nil
	}
}

// TransformField applies a transformation to a field's value in-place.
// If the field is not present in the data, the mapper is a no-op.
func TransformField(field string, transform func(any) (any, error)) FieldMapper {
	return MapField(field, field, transform)
}

// RenameField moves a value from one key to another without modifying it.
// If "from" is not present in the data, the mapper is a no-op.
func RenameField(from, to string) FieldMapper {
	return MapField(from, to, func(v any) (any, error) { return v, nil })
}

// SetDefault sets a field to the given value only if the field is not already
// present in the data map.
func SetDefault(field string, value any) FieldMapper {
	return func(data map[string]any) error {
		if _, ok := data[field]; !ok {
			data[field] = value
		}
		return nil
	}
}

// RemoveFields removes one or more fields from the data map.
func RemoveFields(fields ...string) FieldMapper {
	return func(data map[string]any) error {
		for _, f := range fields {
			delete(data, f)
		}
		return nil
	}
}

// ComputeField reads values from multiple input fields, passes them to a
// compute function, and writes the result to outputField. Input fields that
// differ from outputField are removed. If any input field is missing, the
// mapper is a no-op.
func ComputeField(inputFields []string, outputField string, compute func(inputs map[string]any) (any, error)) FieldMapper {
	return func(data map[string]any) error {
		inputs := make(map[string]any, len(inputFields))
		for _, f := range inputFields {
			v, ok := data[f]
			if !ok {
				return nil
			}
			inputs[f] = v
		}
		result, err := compute(inputs)
		if err != nil {
			return fmt.Errorf("field mapper %v -> %q: %w", inputFields, outputField, err)
		}
		for _, f := range inputFields {
			if f != outputField {
				delete(data, f)
			}
		}
		data[outputField] = result
		return nil
	}
}

// HashPassword creates a mapper that hashes a plain-text password from
// inputField and writes the hash to outputField. Empty passwords are
// silently skipped (the input field is removed and no output is written),
// which allows edit forms to leave the password unchanged.
func HashPassword(inputField, outputField string, generator auth.CredentialGenerator) FieldMapper {
	return func(data map[string]any) error {
		v, ok := data[inputField]
		if !ok {
			return nil
		}
		password, ok := v.(string)
		if !ok {
			return fmt.Errorf("field mapper %q -> %q: expected string, got %T", inputField, outputField, v)
		}
		// Skip empty passwords (e.g., user didn't change password on edit form)
		if password == "" {
			delete(data, inputField)
			return nil
		}
		hash, err := generator.Generate(password)
		if err != nil {
			return fmt.Errorf("field mapper %q -> %q: %w", inputField, outputField, err)
		}
		delete(data, inputField)
		data[outputField] = hash
		return nil
	}
}
