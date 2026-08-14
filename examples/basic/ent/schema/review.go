package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Review demonstrates DisableDelete and a required unique FK back to Book.
type Review struct {
	ent.Schema
}

func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.String("reviewer").NotEmpty(),
		field.Int("rating").Min(1).Max(5),
		field.String("body").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Review) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("book", Book.Type).
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
			ReadOnlyFields:      []string{"created_at"},
			TableColumns:        []string{"reviewer", "rating", "book", "created_at"},
			FilterableColumns:   []string{"reviewer", "rating"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"reviewer", "rating", "body", "book", "created_at"},
			}},
		},
	}
}
