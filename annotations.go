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
	TableColumns []string
	Permissions  []Permission
}

func (VentSchemaAnnotation) Name() string {
	return "VentSchema"
}

func (a VentSchemaAnnotation) MustParse(data string) VentSchemaAnnotation {
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		panic("could not unmarshal annotation: " + err.Error())
	}
	return a
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

func (a VentSchemaAnnotation) tableFields(node *gen.Type) []*gen.Field {
	if a.TableColumns == nil {
		return insensitiveFields(node)
	}
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
