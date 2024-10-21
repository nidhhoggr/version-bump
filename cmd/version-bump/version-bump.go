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

const currentDir = "."

var yesOrNo = []string{"Yes", "No"}

var flags = &struct {
	preReleaseTypeAlpha      bool
	preReleaseTypeBeta       bool
	preReleaseTypeRc         bool
	interactiveMode          bool
	preReleaseMetadataString string
}{}

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("version-bump [%s]", strings.Join(version.TypeStrings, "|")),
	Short: "Bump a semantic version of the project",
	Long: `This application helps incrementing a semantic version of a project.
It can bump the version in multiple different files at once,
for example in package.json and a Dockerfile.`,
	ValidArgs: version.TypeStrings,
	Args:      cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if flags.interactiveMode {
			runInteractiveMode()
		} else {
			runPromptMode(cmd, args)
		}
	},
	Version: bump.Version,
}

func main() {
	rootCmd.PersistentFlags().BoolVar(&flags.preReleaseTypeAlpha, "alpha", false, "alpha prerelease")
	rootCmd.PersistentFlags().BoolVar(&flags.preReleaseTypeBeta, "beta", false, "beta prerelease")
	rootCmd.PersistentFlags().BoolVar(&flags.preReleaseTypeRc, "rc", false, "release candidate prerelease")
	rootCmd.PersistentFlags().BoolVar(&flags.interactiveMode, "interactive", false, "enable interactive mode")
	rootCmd.PersistentFlags().StringVar(&flags.preReleaseMetadataString, "metadata", "", "provide metadata for the prerelease")
	cobra.CheckErr(rootCmd.Execute())
}

func runPromptMode(cmd *cobra.Command, args []string) {
	hasPreRelease := flags.preReleaseTypeAlpha || flags.preReleaseTypeBeta || flags.preReleaseTypeRc
	if len(args) == 1 || hasPreRelease {
		b, err := bump.New(afero.NewOsFs(), osfs.New(path.Join(currentDir, ".git")), osfs.New(currentDir), currentDir)
		if err != nil {
			console.Fatal(errors.Wrap(err, "error preparing project configuration"))
		}

		versionType := version.NotAVersion
		preReleaseType := version.NotAPreRelease

		if len(args) == 1 {
			versionType = version.FromString(args[0])
		}

		if hasPreRelease {
			if flags.preReleaseTypeAlpha {
				preReleaseType = version.AlphaPreRelease
			} else if flags.preReleaseTypeBeta {
				preReleaseType = version.BetaPreRelease
			} else if flags.preReleaseTypeRc {
				preReleaseType = version.ReleaseCandidate
			}
		}

		err = b.Run(&bump.RunArgs{
			ConfirmationPrompt: confirmationPrompt,
			PassphrasePrompt:   passphrasePrompt,
			VersionType:        versionType,
			PreReleaseType:     preReleaseType,
			PreReleaseMetadata: flags.preReleaseMetadataString,
		})
		if err != nil {
			console.Fatal(err)
		}
	} else {
		_ = cmd.Help()
	}
}

func runInteractiveMode() {

	versionType := version.NotAVersion
	preReleaseType := version.NotAPreRelease
	preReleaseMetadata := ""

	s, err := prompt.New().Ask("Would you like to increment the major, minor or patch version?").Choose(yesOrNo)
	if err != nil {
		console.Fatal(err)
	} else if s == "Yes" {
		s, err = prompt.New().Ask(fmt.Sprintf("select which version to increment")).Choose(version.TypeStrings)
		if err != nil {
			console.Fatal(err)
		} else {
			versionType = version.FromString(s)
		}
	}

	s, err = prompt.New().Ask("Is this a prerelease?").Choose(yesOrNo)
	if err != nil {
		console.Fatal(err)
	} else if s == "Yes" {
		s, err = prompt.New().Ask(fmt.Sprintf("select prerelease type")).Choose(version.PreReleaseTypeStrings)
		if err != nil {
			console.Fatal(err)
		} else {
			preReleaseType = version.FromPreReleaseTypeString(s)
			preReleaseMetadata, err = prompt.New().Ask("enter prerelease metadata. (leave empty for none)").Input("")
			if err != nil {
				console.Fatal(err)
			}
		}
	}

	b, err := bump.New(afero.NewOsFs(), osfs.New(path.Join(currentDir, ".git")), osfs.New(currentDir), currentDir)
	if err != nil {
		console.Fatal(errors.Wrap(err, "error preparing project configuration"))
	}
	err = b.Run(&bump.RunArgs{
		ConfirmationPrompt: confirmationPrompt,
		PassphrasePrompt:   passphrasePrompt,
		VersionType:        versionType,
		PreReleaseType:     preReleaseType,
		PreReleaseMetadata: preReleaseMetadata,
	})
	if err != nil {
		console.Fatal(err)
	}
}

func passphrasePrompt() (string, error) {
	return prompt.New().Ask("Input your GPG passphrase:").
		Input("", input.WithEchoMode(input.EchoNone))
}

func confirmationPrompt(currentVersion string, proposedVersion string, file string) (bool, error) {
	s, err := prompt.New().Ask(fmt.Sprintf("continue with the following change: \n%s", console.VersionUpdate(currentVersion, proposedVersion, file))).Choose([]string{"Yes", "No"})
	if err != nil {
		return false, err
	} else {
		return s == "Yes", nil
	}
}
