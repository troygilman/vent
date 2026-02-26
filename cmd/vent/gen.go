package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/troygilman/vent"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func init() {
	genCmd.Flags().StringP("schema", "s", "./ent/schema", "The schema directory")
	rootCmd.AddCommand(genCmd)
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate ent files",
	Long:  `Generate ent files.`,
	Run: func(cmd *cobra.Command, args []string) {
		schemaDirPath := cmd.Flag("schema").Value.String()
		err := entc.Generate(schemaDirPath,
			&gen.Config{
				Features: []gen.Feature{
					gen.FeatureVersionedMigration,
					gen.FeatureUpsert,
					gen.FeatureSnapshot,
				},
			},
			entc.Extensions(vent.NewAdminExtension()),
		)
		if err != nil {
			log.Fatal("running ent codegen:", err)
		}
	},
}
