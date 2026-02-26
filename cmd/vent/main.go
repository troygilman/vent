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

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var rootCmd = &cobra.Command{
	Use:   "vent",
	Short: "A brief description of your application",
	Long:  `A longer description.`,
}
