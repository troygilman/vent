package main

import (
	"log"
	"os"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/spf13/cobra"
)

func init() {
	initCmd.Flags().StringP("schema", "s", "./ent/schema", "The schema output directory")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the vent schemas",
	Long:  `Initialize the vent schemas into a local directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		schemaDirPath := cmd.Flag("schema").Value.String()

		os.MkdirAll(schemaDirPath, 0755)

		ctx, err := schemast.Load(schemaDirPath)
		if err != nil {
			log.Fatalf("failed to load schema: %v", err)
		}

		mutations := []schemast.Mutator{
			&schemast.UpsertSchema{
				Name: "AuthUser",
				Fields: []ent.Field{
					field.String("email").Unique(),
					field.String("password_hash").Sensitive(),
					field.Bool("is_staff").Default(false),
					field.Bool("is_superuser").Default(false),
					field.Bool("is_active").Default(true),
				},
				Edges: []ent.Edge{
					newEdgeTo("groups", "AuthGroup"),
				},
			},
			&schemast.UpsertSchema{
				Name: "AuthGroup",
				Fields: []ent.Field{
					field.String("name").Unique(),
				},
				Edges: []ent.Edge{
					newEdgeTo("permissions", "AuthPermission"),
					newEdgeFrom("users", "AuthUser", "groups"),
				},
			},
			&schemast.UpsertSchema{
				Name: "AuthPermission",
				Fields: []ent.Field{
					field.String("name").Unique(),
				},
			},
		}

		err = schemast.Mutate(ctx, mutations...)
		if err := ctx.Print(schemaDirPath); err != nil {
			log.Fatalf("failed to write schema: %v", err)
		}
	},
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
