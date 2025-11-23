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
					"contains":  slices.Contains[[]any],
					"showField": e.showField,
				}).
				ParseFS(templates, "templates/admin.tmpl"),
		),
	}
}

func (e *AdminExtension) showField(t *gen.Type, f gen.Field) bool {
	a, ok := t.Annotations["VentSchema"]
	if !ok {
		return !f.Sensitive()
	}

	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return !f.Sensitive()
	}

	var annotation VentSchemaAnnotation
	if err := json.Unmarshal(jsonBytes, &annotation); err != nil {
		return !f.Sensitive()
	}

	return annotation.showField(f)
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
