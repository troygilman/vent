package vent

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type AuthUserMixin struct {
	mixin.Schema
	AuthGroupSchemaType any
}

func (AuthUserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().Unique(),
		field.String("password_hash").Sensitive(),
		field.Bool("is_staff").Default(false),
		field.Bool("is_superuser").Default(false),
		field.Bool("is_active").Default(true),
	}
}

func (m AuthUserMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", m.AuthGroupSchemaType),
	}
}

func (AuthUserMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentSchemaAnnotation{
			DisplayField: "email",
			CustomFields: []Field{
				{
					Name:      "password",
					Type:      "string",
					InputType: "password",
				},
			},
			TableColumns: []string{
				"email",
				"is_staff",
				"is_superuser",
				"is_active",
			},
			FieldSets: []FieldSet{
				{
					Fields: []string{
						"id",
						"email",
						"password",
						"is_staff",
						"is_superuser",
						"is_active",
						"groups",
					},
				},
			},
		},
	}
}

type AuthGroupMixin struct {
	mixin.Schema
	AuthUserSchemaType       any
	AuthPermissionSchemaType any
}

func (AuthGroupMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (m AuthGroupMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("permissions", m.AuthPermissionSchemaType),
		edge.From("users", m.AuthUserSchemaType).Ref("groups"),
	}
}

func (AuthGroupMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentSchemaAnnotation{
			DisplayField: "name",
			TableColumns: []string{
				"name",
			},
			FieldSets: []FieldSet{
				{
					Fields: []string{
						"name",
						"permissions",
					},
				},
			},
		},
	}
}

type AuthPermissionMixin struct {
	mixin.Schema
}

func (AuthPermissionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (AuthPermissionMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentSchemaAnnotation{
			DisplayField: "name",
			DisableAdmin: true,
		},
	}
}
