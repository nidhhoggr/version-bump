package cmd

import (
	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	"github.com/spf13/cobra"
)

var minorCmd = &cobra.Command{
	Use:   "minor",
	Short: "Increment a minor version",
	Run: func(cmd *cobra.Command, args []string) {
		run(bump.Minor)
	},
}

func init() {
	rootCmd.AddCommand(minorCmd)
}
