package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Permission struct {
	ent.Schema
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{field.String("name").Unique()}
}
func (Permission) Edges() []ent.Edge {
	return nil
}
func (Permission) Annotations() []schema.Annotation {
	return nil
}
