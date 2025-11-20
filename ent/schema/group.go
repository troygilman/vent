package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Group struct {
	ent.Schema
}

func (Group) Fields() []ent.Field {
	return []ent.Field{field.String("name").Unique()}
}
func (Group) Edges() []ent.Edge {
	return []ent.Edge{edge.To("permissions", Permission.Type), edge.From("users", User.Type).Ref("groups")}
}
func (Group) Annotations() []schema.Annotation {
	return nil
}
