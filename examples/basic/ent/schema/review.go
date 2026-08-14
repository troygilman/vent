package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Review belongs to one Book and one User. Deletes are disabled in admin.
type Review struct {
	ent.Schema
}

func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.Int("rating").Min(1).Max(5),
		field.String("body").Optional().Nillable(),
	}
}

func (Review) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("book", Book.Type).
			Ref("reviews").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("reviews").
			Unique().
			Required(),
	}
}

func (Review) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			SingularDisplayName: "Review",
			PluralDisplayName:   "Reviews",
			DisableDelete:       true,
			TableColumns:        []string{"user", "rating", "book"},
			FilterableColumns:   []string{"rating"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"user", "rating", "body", "book"},
			}},
		},
	}
}
