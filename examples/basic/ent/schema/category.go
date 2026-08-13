package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Category demonstrates custom route + display names for a simple CRUD schema.
type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
		field.String("description").Optional().Nillable(),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("books", Book.Type).Ref("category"),
	}
}

func (Category) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			RouteName:           "categories",
			SingularDisplayName: "Category",
			PluralDisplayName:   "Categories",
			TableColumns:        []string{"name", "description"},
			FieldSets: []vent.FieldSet{{
				Label:  "Category",
				Fields: []string{"name", "description"},
			}},
		},
	}
}
