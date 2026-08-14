package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// User extends the Vent auth user mixin with an extra field and schema-level
// overrides: custom table columns, fieldsets, and an extra permission name.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Time("last_login").Optional(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("author", Author.Type).Unique(),
		edge.To("reviews", Review.Type),
	}
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
			SingularDisplayName: "User",
			PluralDisplayName:   "Users",
			TableColumns: []string{
				"email",
				"is_staff",
				"is_superuser",
				"is_active",
				"last_login",
			},
			FilterableColumns: []string{
				"email",
				"is_staff",
				"is_active",
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
