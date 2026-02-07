package vent

import (
	"embed"
	"strings"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

//go:embed templates
var templates embed.FS

type AdminExtension struct {
	entc.DefaultExtension
	config VentExtensionConfig
}

func NewAdminExtension(opts ...VentExtensionConfigOption) entc.Extension {
	config := VentExtensionConfig{
		AdminPath:  "/admin/",
		UserSchema: "AuthUser",
	}
	for _, opt := range opts {
		config = opt(config)
	}
	return &AdminExtension{
		config: config,
	}
}

func (ext *AdminExtension) Annotations() []entc.Annotation {
	return []entc.Annotation{
		VentConfigAnnotation{
			VentExtensionConfig: ext.config,
		},
	}
}

func (e *AdminExtension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(
			gen.NewTemplate("admin").
				Funcs(template.FuncMap{
					"fields":         fields,
					"tableColumns":   tableColumns,
					"fieldSets":      fieldSets,
					"passwordFields": passwordFields,
					"hasSuffix":      strings.HasSuffix,
					"trimSuffix":     strings.TrimSuffix,
				}).
				ParseFS(templates, "templates/admin.tmpl"),
		),
		gen.MustParse(
			gen.NewTemplate("getters").
				ParseFS(templates, "templates/getters.tmpl"),
		),
		gen.MustParse(
			gen.NewTemplate("migrate").
				ParseFS(templates, "templates/migrate.tmpl"),
		),
	}
}

func fields(node *gen.Type) []Field {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(node); err != nil {
		return insensitiveFields(node)
	}
	return annotation.tableFields(node)
}

func fieldSets(node *gen.Type) []FieldSet {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(node); err != nil {
		return nil
	}
	return annotation.FieldSets

}

func tableColumns(node *gen.Type) []string {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(node); err != nil {
		columns := []string{}
		for _, f := range node.Fields {
			if !f.Sensitive() {
				columns = append(columns, f.Name)
			}
		}
		return columns
	}
	return annotation.TableColumns
}

func insensitiveFields(node *gen.Type) []Field {
	result := []Field{}
	for _, f := range node.Fields {
		result = append(result, Field{
			Name: f.Name,
			Type: f.Type.Type.String(),
		})
	}
	return result
}

// passwordFields returns sensitive fields whose names end with "_hash".
// These are assumed to be password fields that need a virtual form input
// and a HashPassword field mapper in the generated admin code.
func passwordFields(node *gen.Type) []*gen.Field {
	result := []*gen.Field{}
	for _, f := range node.Fields {
		if f.Sensitive() && strings.HasSuffix(f.Name, "_hash") {
			result = append(result, f)
		}
	}
	return result
}

type VentExtensionConfig struct {
	AdminPath  string
	UserSchema string
}

type VentExtensionConfigOption func(VentExtensionConfig) VentExtensionConfig

func WithAdminPath(path string) VentExtensionConfigOption {
	return func(vec VentExtensionConfig) VentExtensionConfig {
		vec.AdminPath = path
		return vec
	}
}
