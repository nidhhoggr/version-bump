package main

import (
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"path"
	"strings"

	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	"github.com/spf13/cobra"
)

var acceptedArgs = []string{"major", "minor", "patch"}

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("version-bump [%s]", strings.Join(acceptedArgs, "|")),
	Short: "Bump a semantic version of the project",
	Long: `This application helps incrementing a semantic version of a project.
It can bump the version in multiple different files at once,
for example in package.json and a Dockerfile.`,
	ValidArgs: []string{"major", "minor", "patch"},
	Args:      cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			dir := "."
			b, err := bump.New(afero.NewOsFs(), osfs.New(path.Join(dir, ".git")), osfs.New(dir), dir, true)
			if err != nil {
				console.Fatal(errors.Wrap(err, "error preparing project configuration"))
			}
			_ = b.Run(version.FromString(args[0]))
		} else {
			_ = cmd.Help()
		}
	},
	Version: bump.Version,
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}
