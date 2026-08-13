package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Tag is a minimal schema with no Vent annotations — codegen uses defaults
// (route name, display names, and a surface built from all supported fields/edges).
type Tag struct {
	ent.Schema
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("books", Book.Type).Ref("tags"),
	}
}
