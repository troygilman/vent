package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{field.String("email").Unique(), field.String("password_hash"), field.Bool("is_staff").Default(false), field.Bool("is_active").Default(true)}
}
func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("groups", Group.Type)}
}
func (User) Annotations() []schema.Annotation {
	return nil
}
