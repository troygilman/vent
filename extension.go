package vent

import (
	"embed"
	"encoding/json"
	"slices"
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
		AdminPath: "/admin/",
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
					"contains":    slices.Contains[[]any],
					"tableFields": tableFields,
				}).
				ParseFS(templates, "templates/admin.tmpl"),
		),
		gen.MustParse(
			gen.NewTemplate("migratedata").
				Funcs(template.FuncMap{
					"stringify": func(v any) (string, error) {
						buf, err := json.Marshal(v)
						if err != nil {
							return "", err
						}
						return string(buf), nil
					},
				}).
				ParseFS(templates, "templates/migratedata.tmpl"),
		),
	}
}

func tableFields(node *gen.Type) []*gen.Field {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(node); err != nil {
		return insensitiveFields(node)
	}
	return annotation.tableFields(node)
}

func insensitiveFields(node *gen.Type) []*gen.Field {
	result := []*gen.Field{}
	for _, f := range node.Fields {
		if !f.Sensitive() {
			result = append(result, f)
		}
	}
	return result
}

type VentExtensionConfig struct {
	AdminPath string
}

type VentExtensionConfigOption func(VentExtensionConfig) VentExtensionConfig

func WithAdminPath(path string) VentExtensionConfigOption {
	return func(vec VentExtensionConfig) VentExtensionConfig {
		vec.AdminPath = path
		return vec
	}
}
