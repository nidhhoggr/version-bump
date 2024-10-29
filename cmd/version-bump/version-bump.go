package main

import (
	"fmt"
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/pkg/errors"
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
	autoConfirm              bool
	disablePrompts           bool
	isDryRun                 bool
	shouldDebug              bool
	preReleaseMetadataString string
	passphrase               string
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
	rootCmd.PersistentFlags().BoolVar(&flags.autoConfirm, "auto-confirm", false, "disable confirmation prompts and automatically confirm")
	rootCmd.PersistentFlags().BoolVar(&flags.disablePrompts, "disable-prompts", false, "disable passphrase and confirmation prompts. Caution: this will result in unsigned commits, tags and releases!")
	rootCmd.PersistentFlags().BoolVar(&flags.isDryRun, "dry-run", false, "perform a dry run without modifying any files or interacting with git")
	rootCmd.PersistentFlags().BoolVar(&flags.shouldDebug, "debug", false, "output debug information to the console")
	rootCmd.PersistentFlags().StringVar(&flags.preReleaseMetadataString, "metadata", "", "provide metadata for the prerelease")
	rootCmd.PersistentFlags().StringVar(&flags.passphrase, "passphrase", "", "provide gpg passphrase as a flag instead of a secure prompt. Caution!")
	cobra.CheckErr(rootCmd.Execute())
}

func runPromptMode(cmd *cobra.Command, args []string) {
	hasPreRelease := flags.preReleaseTypeAlpha || flags.preReleaseTypeBeta || flags.preReleaseTypeRc
	if len(args) == 1 || hasPreRelease {
		b, err := bump.New(currentDir)
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
			IsDryRun:           flags.isDryRun,
			ShouldDebug:        flags.shouldDebug,
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
			preReleaseMetadata, err = prompt.New().Ask("enter prerelease metadata. (leave empty for none)").Input(flags.preReleaseMetadataString)
			if err != nil {
				console.Fatal(err)
			}
		}
	}

	b, err := bump.New(currentDir)
	if err != nil {
		console.Fatal(errors.Wrap(err, "error preparing project configuration"))
	}
	err = b.Run(&bump.RunArgs{
		ConfirmationPrompt: confirmationPrompt,
		PassphrasePrompt:   passphrasePrompt,
		VersionType:        versionType,
		PreReleaseType:     preReleaseType,
		PreReleaseMetadata: preReleaseMetadata,
		IsDryRun:           flags.isDryRun,
		ShouldDebug:        flags.shouldDebug,
	})
	if err != nil {
		console.Fatal(err)
	}
}

func passphrasePrompt() (string, error) {
	if len(flags.passphrase) > 0 || flags.disablePrompts {
		return flags.passphrase, nil
	}
	return prompt.New().Ask("Input your GPG passphrase:").
		Input("", input.WithEchoMode(input.EchoNone))
}

func confirmationPrompt(currentVersion string, proposedVersion string, file string) (bool, error) {
	if flags.autoConfirm || flags.disablePrompts {
		return true, nil
	}
	s, err := prompt.New().Ask(fmt.Sprintf("continue with the following change: \n%s", console.VersionUpdate(currentVersion, proposedVersion, file))).Choose([]string{"Yes", "No"})
	if err != nil {
		return false, err
	} else {
		return s == "Yes", nil
	}
}
