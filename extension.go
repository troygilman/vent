package vent

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type AdminExtension struct {
	entc.DefaultExtension
}

func (*AdminExtension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(gen.NewTemplate("admin").ParseFiles("templates/admin.tmpl")),
	}
}
