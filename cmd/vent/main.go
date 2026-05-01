package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "vent",
	Short: "Vent Ent admin framework tooling",
	Long:  `Vent provides code generation helpers for the opinionated Vent Ent admin framework.`,
}
