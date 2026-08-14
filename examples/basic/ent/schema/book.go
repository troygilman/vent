package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// Book is the main feature showcase: mixed field kinds, edge shapes, table
// columns (including edges), fieldsets, read-only fields, custom permissions,
// a custom virtual field, and a custom admin route name.
type Book struct {
	ent.Schema
}

func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.String("isbn").Optional().Nillable(),
		field.Int("pages").NonNegative().Default(0),
		field.Float("price").Min(0).Default(0),
		field.Bool("published").Default(false),
		field.Time("published_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Int("view_count").NonNegative().Default(0),
		// Sensitive fields are omitted from the default admin surface; the
		// custom "notes" field below reads/writes this value instead.
		field.String("internal_notes").Optional().Nillable().Sensitive(),
	}
}

func (Book) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("author", Author.Type).Unique().Required(),
		edge.To("category", Category.Type).Unique(),
		edge.To("tags", Tag.Type),
		edge.To("reviews", Review.Type),
	}
}

func (Book) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			RouteName:           "books",
			SingularDisplayName: "Book",
			PluralDisplayName:   "Books",
			// List columns may include edges (author/category) as well as fields.
			TableColumns: []string{
				"title",
				"author",
				"category",
				"published",
				"price",
				"published_at",
			},
			FilterableColumns: []string{
				"title",
				"published",
				"pages",
			},
			FieldSets: []vent.FieldSet{{
				Label: "Book",
				Fields: []string{
					"title",
					"isbn",
					"author",
					"category",
					"tags",
					"pages",
					"price",
					"published",
					"published_at",
					"view_count",
					"created_at",
					"notes",
				},
			}},
			ReadOnlyFields: []string{"view_count", "created_at"},
			CustomFields: []vent.Field{
				{Name: "notes", Type: "string", InputType: "string"},
			},
			Permissions: []vent.Permission{
				{Name: "publish", Desc: "Publish a book"},
			},
		},
	}
}
