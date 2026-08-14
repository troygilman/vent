package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Author is a 1:1 extension of User: Author.id is the User primary key.
type Author struct {
	ent.Schema
}

func (Author) Fields() []ent.Field {
	return []ent.Field{
		// Application-assigned so Author.id can equal User.id (shared PK / FK).
		field.Int("id"),
		field.Bool("active").Default(true),
	}
}

func (Author) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("author").
			Unique().
			Required(),
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
