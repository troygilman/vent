package vent

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type VentExtensionConfig struct {
	AdminPath string
}

type AdminExtension struct {
	entc.DefaultExtension
	config *VentExtensionConfig
}

func NewAdminExtension(config *VentExtensionConfig) entc.Extension {
	if config == nil {
		config = &VentExtensionConfig{
			AdminPath: "/admin/",
		}
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
		gen.MustParse(gen.NewTemplate("admin").ParseFiles("templates/admin.tmpl")),
	}
}
