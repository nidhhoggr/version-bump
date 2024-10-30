package main

import (
	"fmt"
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"strings"

	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	"github.com/spf13/cobra"
)

const currentDir = "."

var yesOrNo = []string{"Yes", "No"}

var flags = &struct {
	PrereleaseTypeAlpha      bool
	PrereleaseTypeBeta       bool
	PrereleaseTypeRc         bool
	interactiveMode          bool
	autoConfirm              bool
	disablePrompts           bool
	isDryRun                 bool
	shouldDebug              bool
	PrereleaseMetadataString string
	passphrase               string
}{}

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("version-bump [%s]", strings.Join(version.TypeStrings, "|")),
	Short: "Bump a semantic version of the project",
	Long: `This application increments the semantic version of a project.
It can bump semantic versions in multiple different files at once,
as well as automate prerelease versioning and promotion.`,
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
	rootCmd.PersistentFlags().BoolVar(&flags.PrereleaseTypeAlpha, "alpha", false, "alpha Prerelease")
	rootCmd.PersistentFlags().BoolVar(&flags.PrereleaseTypeBeta, "beta", false, "beta Prerelease")
	rootCmd.PersistentFlags().BoolVar(&flags.PrereleaseTypeRc, "rc", false, "release candidate Prerelease")
	rootCmd.PersistentFlags().BoolVar(&flags.interactiveMode, "interactive", false, "enable interactive mode")
	rootCmd.PersistentFlags().BoolVar(&flags.autoConfirm, "auto-confirm", false, "disable confirmation prompts and automatically confirm")
	rootCmd.PersistentFlags().BoolVar(&flags.disablePrompts, "disable-prompts", false, "disable passphrase and confirmation prompts. Caution: this will result in unsigned commits, tags and releases!")
	rootCmd.PersistentFlags().BoolVar(&flags.isDryRun, "dry-run", false, "perform a dry run without modifying any files or interacting with git")
	rootCmd.PersistentFlags().BoolVar(&flags.shouldDebug, "debug", false, "output debug information to the console")
	rootCmd.PersistentFlags().StringVar(&flags.PrereleaseMetadataString, "metadata", "", "provide metadata for the Prerelease")
	rootCmd.PersistentFlags().StringVar(&flags.passphrase, "passphrase", "", "provide gpg passphrase as a flag instead of a secure prompt. Caution!")
	cobra.CheckErr(rootCmd.Execute())
}

func runPromptMode(cmd *cobra.Command, args []string) {
	hasPrerelease := flags.PrereleaseTypeAlpha || flags.PrereleaseTypeBeta || flags.PrereleaseTypeRc
	if len(args) == 1 || hasPrerelease {
		console.DebuggingEnabled = flags.shouldDebug
		b, err := bump.New(currentDir)
		if err != nil {
			console.Fatal(err)
		}

		versionType := version.NotAVersion
		PrereleaseType := version.NotAPrerelease

		if len(args) == 1 {
			versionType = version.FromString(args[0])
		}

		if hasPrerelease {
			if flags.PrereleaseTypeAlpha {
				PrereleaseType = version.AlphaPrerelease
			} else if flags.PrereleaseTypeBeta {
				PrereleaseType = version.BetaPrerelease
			} else if flags.PrereleaseTypeRc {
				PrereleaseType = version.ReleaseCandidate
			}
		}

		err = b.Run(&bump.RunArgs{
			ConfirmationPrompt: confirmationPrompt,
			PassphrasePrompt:   passphrasePrompt,
			VersionType:        versionType,
			PrereleaseType:     PrereleaseType,
			PrereleaseMetadata: flags.PrereleaseMetadataString,
			IsDryRun:           flags.isDryRun,
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
	PrereleaseType := version.NotAPrerelease
	PrereleaseMetadata := ""

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

	s, err = prompt.New().Ask("Is this a Prerelease?").Choose(yesOrNo)
	if err != nil {
		console.Fatal(err)
	} else if s == "Yes" {
		s, err = prompt.New().Ask(fmt.Sprintf("select Prerelease type")).Choose(version.PrereleaseTypeStrings)
		if err != nil {
			console.Fatal(err)
		} else {
			PrereleaseType = version.FromPrereleaseTypeString(s)
			PrereleaseMetadata, err = prompt.New().Ask("enter Prerelease metadata. (leave empty for none)").Input(flags.PrereleaseMetadataString)
			if err != nil {
				console.Fatal(err)
			}
		}
	}

	console.DebuggingEnabled = flags.shouldDebug
	b, err := bump.New(currentDir)
	if err != nil {
		console.Fatal(err)
	}
	err = b.Run(&bump.RunArgs{
		ConfirmationPrompt: confirmationPrompt,
		PassphrasePrompt:   passphrasePrompt,
		VersionType:        versionType,
		PrereleaseType:     PrereleaseType,
		PrereleaseMetadata: PrereleaseMetadata,
		IsDryRun:           flags.isDryRun,
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
