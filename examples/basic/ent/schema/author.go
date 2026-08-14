package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Author demonstrates a name-bearing schema used as a unique foreign key target.
type Author struct {
	ent.Schema
}

func (Author) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("bio").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

func (Author) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("books", Book.Type).Ref("author"),
	}
}

func (Author) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			SingularDisplayName: "Author",
			PluralDisplayName:   "Authors",
			TableColumns:        []string{"name", "active"},
			FilterableColumns:   []string{"name", "active"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"name", "bio", "active"},
			}},
		},
	}
}
