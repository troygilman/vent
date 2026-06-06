package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

type AuthUser struct {
	ent.Schema
}

func (AuthUser) Fields() []ent.Field {
	return []ent.Field{
		field.Time("last_login").Optional(),
	}
}
func (AuthUser) Edges() []ent.Edge {
	return nil
}
func (AuthUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.AuthUserMixin{
			GroupSchemaType: AuthGroup.Type,
		},
	}
}

func (AuthUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			DisplayField: "email",
			TableColumns: []string{
				"email",
				"is_staff",
				"is_superuser",
				"is_active",
				"last_login",
			},
			FieldSets: []vent.FieldSet{
				{
					Fields: []string{
						"id",
						"email",
						"password",
						"is_staff",
						"is_superuser",
						"is_active",
						"groups",
						"last_login",
					},
				},
			},
		},
	}
}
