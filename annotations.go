package vent

import (
	"encoding/json"
	"errors"

	"entgo.io/ent/entc/gen"
)

type VentConfigAnnotation struct {
	VentExtensionConfig
}

func (VentConfigAnnotation) Name() string {
	return "VentConfig"
}

type Permission struct {
	Name string
	Desc string
}

type VentSchemaAnnotation struct {
	DisableAdmin  bool
	DisplayField  string
	CustomFields  []Field
	FieldMappings []FieldMapping
	FieldSets     []FieldSet
	TableColumns  []string
	Permissions   []Permission
}

func (VentSchemaAnnotation) Name() string {
	return "VentSchema"
}

func (a *VentSchemaAnnotation) parse(node *gen.Type) error {
	annotation, ok := node.Annotations[a.Name()]
	if !ok {
		return errors.New("vent schema does not exist in node annotations")
	}

	jsonBytes, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonBytes, a)
}

func (a VentSchemaAnnotation) tableFields(node *gen.Type) []Field {
	fields := make([]Field, 0, len(node.Fields)+len(a.CustomFields))
	for _, f := range node.Fields {
		fields = append(fields, Field{
			Name:      f.Name,
			Type:      f.Type.Type.String(),
			Sensitive: f.Sensitive(),
		})
	}
	fields = append(fields, a.CustomFields...)
	return fields
}

type Field struct {
	Name      string
	Type      string
	InputType string
	Sensitive bool
}

type FieldSet struct {
	Label  string
	Fields []string
}

// FieldMapping defines how a custom input field maps to a database field with an optional transform
type FieldMapping struct {
	From      string // Input field name (e.g., "password")
	To        string // Database field name (e.g., "password_hash")
	Transform string // Transform function key (e.g., "hash") - looked up in FieldTransforms registry
}
