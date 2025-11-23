package main

import (
	"log"
	"os"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

func main() {
	os.MkdirAll("./ent/schema", 0755)

	ctx, err := schemast.Load("./ent/schema")
	if err != nil {
		log.Fatalf("failed to load schema: %v", err)
	}

	mutations := []schemast.Mutator{
		&schemast.UpsertSchema{
			Name: "User",
			Fields: []ent.Field{
				field.String("email").
					Unique(),
				field.String("password_hash"),
				field.Bool("is_staff").
					Default(false),
				field.Bool("is_active").
					Default(true),
			},
			Edges: []ent.Edge{
				newEdgeTo("groups", "Group"),
			},
			// Annotations: []schema.Annotation{
			// 	vent.VentSchemaAnnotation{
			// 		TableColumns: []string{
			// 			"email",
			// 			"is_staff",
			// 			"is_active",
			// 		},
			// 	},
			// },
		},
		&schemast.UpsertSchema{
			Name: "Group",
			Fields: []ent.Field{
				field.String("name").
					Unique(),
			},
			Edges: []ent.Edge{
				newEdgeTo("permissions", "Permission"),
				newEdgeFrom("users", "User", "groups"),
			},
			// Annotations: []schema.Annotation{
			// 	vent.VentSchemaAnnotation{
			// 		TableColumns: []string{
			// 			"name",
			// 		},
			// 	},
			// },
		},
		&schemast.UpsertSchema{
			Name: "Permission",
			Fields: []ent.Field{
				field.String("name").
					Unique(),
			},
			// Annotations: []schema.Annotation{
			// 	vent.VentSchemaAnnotation{
			// 		TableColumns: []string{
			// 			"name",
			// 		},
			// 	},
			// },
		},
	}

	err = schemast.Mutate(ctx, mutations...)
	if err := ctx.Print("./ent/schema"); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}
}

type placeholder struct {
	ent.Schema
}

func withType(e ent.Edge, typeName string) ent.Edge {
	e.Descriptor().Type = typeName
	return e
}

func newEdgeTo(edgeName, otherType string) ent.Edge {
	e := edge.To(edgeName, placeholder.Type)
	return withType(e, otherType)
}

func newEdgeFrom(edgeName, otherType, ref string) ent.Edge {
	e := edge.From(edgeName, placeholder.Type).
		Ref(ref)
	return withType(e, otherType)
}
