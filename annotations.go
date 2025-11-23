package vent

import (
	"slices"

	"entgo.io/ent/entc/gen"
)

type VentConfigAnnotation struct {
	VentExtensionConfig
}

func (VentConfigAnnotation) Name() string {
	return "VentConfig"
}

type VentSchemaAnnotation struct {
	TableColumns []string
}

func (VentSchemaAnnotation) Name() string {
	return "VentSchema"
}

func (a VentSchemaAnnotation) showField(f gen.Field) bool {
	return slices.Contains(a.TableColumns, f.Name)
}

type VentFieldAnnotation struct {
}

func (VentFieldAnnotation) Name() string {
	return "VentField"
}
