package vent

import "entgo.io/ent/entc/gen"

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

func (a VentSchemaAnnotation) tableFields(node *gen.Type) []*gen.Field {
	fieldMap := make(map[string]*gen.Field)
	for _, f := range node.Fields {
		fieldMap[f.Name] = f
	}
	results := make([]*gen.Field, len(a.TableColumns))
	for idx, fieldName := range a.TableColumns {
		f, ok := fieldMap[fieldName]
		if !ok {
			panic("cannot find " + fieldName + " in field map")
		}
		results[idx] = f
	}
	return results
}

type VentFieldAnnotation struct {
}

func (VentFieldAnnotation) Name() string {
	return "VentField"
}
