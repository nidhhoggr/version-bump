package cmd

import (
	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Increment a patch version",
	Run: func(cmd *cobra.Command, args []string) {
		run(bump.Patch)
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
}
