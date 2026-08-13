package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// ApiKey demonstrates DisableAdmin: the Ent model exists for the app, but Vent
// does not generate admin routes, nav entries, or CRUD permissions for it.
type ApiKey struct {
	ent.Schema
}

func (ApiKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
		field.String("token").Sensitive().NotEmpty(),
	}
}

func (ApiKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			DisableAdmin: true,
		},
	}
}
