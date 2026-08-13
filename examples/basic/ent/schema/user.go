package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Time("last_login").Optional(),
	}
}
func (User) Edges() []ent.Edge {
	return nil
}
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.UserMixin{
			GroupSchemaType: PermissionGroup.Type,
		},
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			TableColumns: []string{
				"email",
				"is_staff",
				"is_superuser",
				"is_active",
				"last_login",
			},
			Permissions: []vent.Permission{
				{Name: "impersonate", Desc: "Act as another user"},
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
