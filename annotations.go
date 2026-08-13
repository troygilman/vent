package vent

import (
	"encoding/json"
	"errors"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema"
)

type VentConfigAnnotation struct {
	VentExtensionConfig
	Configs []NodeRenderConfig
}

func (VentConfigAnnotation) Name() string {
	return "VentConfig"
}

type Permission struct {
	Name string
	Desc string
}

type AuthRole string

const (
	AuthRoleUser       AuthRole = "user"
	AuthRoleGroup      AuthRole = "group"
	AuthRolePermission AuthRole = "permission"
)

// VentAuthMixinAnnotation marks schemas that use Vent's auth mixins.
type VentAuthMixinAnnotation struct {
	Role AuthRole
}

func (VentAuthMixinAnnotation) Name() string {
	return "VentAuthMixin"
}

func (a *VentAuthMixinAnnotation) parse(node *gen.Type) error {
	annotation, ok := node.Annotations[a.Name()]
	if !ok {
		return errors.New("vent auth mixin does not exist in node annotations")
	}

	jsonBytes, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonBytes, a)
}

type VentSchemaAnnotation struct {
	DisableAdmin        bool
	RouteName           string
	SingularDisplayName string
	PluralDisplayName   string
	CustomFields        []Field
	FieldSets           []FieldSet
	TableColumns        []string
	Permissions         []Permission
}

func (VentSchemaAnnotation) Name() string {
	return "VentSchema"
}

// Merge allows schema-level VentSchema annotations to override mixin defaults.
func (VentSchemaAnnotation) Merge(annotation schema.Annotation) schema.Annotation {
	return annotation
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
