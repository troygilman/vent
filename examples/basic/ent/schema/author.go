package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Author is a 1:1 profile of a User. Authors have their own PK and a required
// unique FK to users.
type Author struct {
	ent.Schema
}

func (Author) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Bool("active").Default(true),
	}
}

func (Author) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("author").
			Unique().
			Required().
			Field("user_id"),
		edge.From("books", Book.Type).Ref("author"),
	}
}

func (Author) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			SingularDisplayName: "Author",
			PluralDisplayName:   "Authors",
			TableColumns:        []string{"user", "active"},
			FilterableColumns:   []string{"active"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"user", "active"},
			}},
		},
	}
}
