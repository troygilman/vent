package vent

import (
	"embed"

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

func (*AdminExtension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(gen.NewTemplate("admin").ParseFS(templates, "templates/admin.tmpl")),
	}
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
