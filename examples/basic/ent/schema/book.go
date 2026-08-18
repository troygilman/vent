package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Book is the main showcase: mixed field kinds, a unique FK, list filters,
// read-only fields, a custom virtual field, and an extra permission.
type Book struct {
	ent.Schema
}

func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.Int("pages").NonNegative().Default(0),
		field.Bool("published").Default(false),
		field.Time("published_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		// Sensitive fields are omitted from the default admin surface; the
		// custom "notes" field below reads/writes this value instead.
		field.String("internal_notes").Optional().Nillable().Sensitive(),
	}
}

func (Book) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("author", Author.Type).Unique().Required(),
		edge.To("reviews", Review.Type),
	}
}

func (Book) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			RouteName:           "books",
			SingularDisplayName: "Book",
			PluralDisplayName:   "Books",
			TableColumns:        []string{"title", "author", "published", "pages"},
			FilterableColumns:   []string{"title", "published", "pages"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{
					"title",
					"author",
					"pages",
					"published",
					"published_at",
					"created_at",
					"notes",
				},
			}},
			ReadOnlyFields: []string{"created_at"},
			CustomFields: []vent.Field{
				{Name: "notes", Type: "string", InputType: "string"},
			},
			Permissions: []vent.Permission{
				{Name: "publish", Desc: "Publish a book"},
			},
		},
	}
}
