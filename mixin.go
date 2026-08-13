package vent

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type UserMixin struct {
	mixin.Schema
	GroupSchemaType any
}

func (UserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().Unique(),
		field.String("password_hash").Optional().Nillable().Sensitive(),
		field.Bool("is_staff").Default(false),
		field.Bool("is_superuser").Default(false),
		field.Bool("is_active").Default(true),
	}
}

func (m UserMixin) Edges() []ent.Edge {
	if m.GroupSchemaType == nil {
		panic("GroupSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.To("groups", m.GroupSchemaType),
	}
}

func (UserMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentAuthMixinAnnotation{Role: AuthRoleUser},
		VentSchemaAnnotation{
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

type PermissionGroupMixin struct {
	mixin.Schema
	UserSchemaType       any
	PermissionSchemaType any
}

func (PermissionGroupMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (m PermissionGroupMixin) Edges() []ent.Edge {
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

func (PermissionGroupMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentAuthMixinAnnotation{Role: AuthRoleGroup},
		VentSchemaAnnotation{
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

type PermissionMixin struct {
	mixin.Schema
	GroupSchemaType any
}

func (PermissionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (m PermissionMixin) Edges() []ent.Edge {
	if m.GroupSchemaType == nil {
		panic("GroupSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.From("groups", m.GroupSchemaType).Ref("permissions"),
	}
}

func (PermissionMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentAuthMixinAnnotation{Role: AuthRolePermission},
		VentSchemaAnnotation{
			DisableCreate:  true,
			DisableDelete:  true,
			ReadOnlyFields: []string{"name"},
			TableColumns: []string{
				"name",
				"groups",
			},
			FieldSets: []FieldSet{
				{
					Fields: []string{
						"name",
						"groups",
					},
				},
			},
		},
	}
}
