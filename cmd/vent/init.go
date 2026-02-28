package main

import (
	"log"
	"os"

	"entgo.io/contrib/schemast"
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
			},
			&schemast.UpsertSchema{
				Name: "AuthGroup",
			},
			&schemast.UpsertSchema{
				Name: "AuthPermission",
			},
		}

		err = schemast.Mutate(ctx, mutations...)
		if err := ctx.Print(schemaDirPath); err != nil {
			log.Fatalf("failed to write schema: %v", err)
		}
	},
}
