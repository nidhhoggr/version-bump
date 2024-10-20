package main

import (
	"fmt"
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
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

var preReleaseTypeAlpha bool
var preReleaseTypeBeta bool
var preReleaseTypeRc bool
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
		hasPreRelease := preReleaseTypeAlpha || preReleaseTypeBeta || preReleaseTypeRc
		if len(args) == 1 || hasPreRelease {
			dir := "."
			b, err := bump.New(afero.NewOsFs(), osfs.New(path.Join(dir, ".git")), osfs.New(dir), dir, passphrasePrompt)
			if err != nil {
				console.Fatal(errors.Wrap(err, "error preparing project configuration"))
			}

			versionType := version.NotAVersion
			preReleaseType := version.NotAPreRelease

			if len(args) == 1 {
				versionType = version.FromString(args[0])
			}

			if hasPreRelease {
				if preReleaseTypeAlpha {
					preReleaseType = version.AlphaPreRelease
				} else if preReleaseTypeBeta {
					preReleaseType = version.BetaPreRelease
				} else if preReleaseTypeRc {
					preReleaseType = version.ReleaseCandidate
				}
			}

			err = b.Run(&bump.RunArgs{
				ConfirmationPrompt: confirmationPrompt,
				VersionType:        versionType,
				PreReleaseType:     preReleaseType,
			})
			if err != nil {
				console.Fatal(err)
			}
		} else {
			_ = cmd.Help()
		}
	},
	Version: bump.Version,
}

func main() {
	rootCmd.PersistentFlags().BoolVar(&preReleaseTypeAlpha, "alpha", false, "alpha prerelease")
	rootCmd.PersistentFlags().BoolVar(&preReleaseTypeBeta, "beta", false, "beta prerelease")
	rootCmd.PersistentFlags().BoolVar(&preReleaseTypeRc, "rc", false, "release candidate prerelease")
	cobra.CheckErr(rootCmd.Execute())
}

func passphrasePrompt() (string, error) {
	return prompt.New().Ask("Input your passphrase:").
		Input("", input.WithEchoMode(input.EchoPassword))
}

func confirmationPrompt(proposedVersion string) (bool, error) {
	s, err := prompt.New().Ask(fmt.Sprintf("continue with new version: %s", proposedVersion)).Choose([]string{"Yes", "No"})
	if err != nil {
		return false, err
	} else {
		return s == "Yes", nil
	}
}
