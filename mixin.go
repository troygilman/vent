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
	PermissionGroupSchemaType any
}

func (UserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").NotEmpty().Unique(),
		field.String("password_hash").Optional().Nillable().Sensitive(),
		field.Bool("is_staff").Default(false),
		field.Bool("is_superuser").Default(false),
		field.Bool("is_active").Default(true),
	}
}

func (m UserMixin) Edges() []ent.Edge {
	if m.PermissionGroupSchemaType == nil {
		panic("PermissionGroupSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.To("permission_groups", m.PermissionGroupSchemaType),
	}
}

func (UserMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentAuthMixinAnnotation{Role: AuthRoleUser},
		VentSchemaAnnotation{
			TableColumns: []string{
				"username",
				"is_staff",
				"is_superuser",
				"is_active",
			},
			FilterableColumns: []string{
				"username",
				"is_staff",
				"is_active",
			},
			FieldSets: []FieldSet{
				{
					Fields: []string{
						"id",
						"username",
						"password",
						"is_staff",
						"is_superuser",
						"is_active",
						"permission_groups",
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
		edge.From("users", m.UserSchemaType).Ref("permission_groups"),
	}
}

func (PermissionGroupMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		VentAuthMixinAnnotation{Role: AuthRolePermissionGroup},
		VentSchemaAnnotation{
			TableColumns: []string{
				"name",
			},
			FilterableColumns: []string{
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
	PermissionGroupSchemaType any
}

func (PermissionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (m PermissionMixin) Edges() []ent.Edge {
	if m.PermissionGroupSchemaType == nil {
		panic("PermissionGroupSchemaType cannot be nil")
	}
	return []ent.Edge{
		edge.From("permission_groups", m.PermissionGroupSchemaType).Ref("permissions"),
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
				"permission_groups",
			},
			FilterableColumns: []string{
				"name",
			},
			FieldSets: []FieldSet{
				{
					Fields: []string{
						"name",
						"permission_groups",
					},
				},
			},
		},
	}
}
