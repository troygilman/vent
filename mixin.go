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
	GroupSchemaType any
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
	if m.GroupSchemaType == nil {
		panic("GroupSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.To("groups", m.GroupSchemaType),
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
			FieldMappings: []FieldMapping{
				{
					From:      "password",
					To:        "password_hash",
					Transform: "hash_password",
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
	UserSchemaType       any
	PermissionSchemaType any
}

func (AuthGroupMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (m AuthGroupMixin) Edges() []ent.Edge {
	if m.PermissionSchemaType == nil {
		panic("PermissionSchemaType cannot be nil")
	}
	if m.UserSchemaType == nil {
		panic("UserSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.To("permissions", m.PermissionSchemaType),
		edge.From("users", m.UserSchemaType).Ref("groups"),
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
	GroupSchemaType any
}

func (AuthPermissionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (m AuthPermissionMixin) Edges() []ent.Edge {
	if m.GroupSchemaType == nil {
		panic("GroupSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.From("groups", m.GroupSchemaType).Ref("permissions"),
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
