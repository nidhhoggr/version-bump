package bump

import (
	"github.com/joe-at-startupmedia/version-bump/v2/git"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
)

const (
	Version string = "2.1.0"
)

type Bump struct {
	FS            afero.Fs
	Git           git.Config
	Configuration Configuration
}

type Configuration struct {
	Docker     Language
	Go         Language
	JavaScript Language
}

type Language struct {
	Directories  []string
	ExcludeFiles []string `toml:"exclude_files"`
	Enabled      bool
}

type RunArgs struct {
	ConfirmationPrompt func(string, string, string) (bool, error)
	PassphrasePrompt   func() (string, error)
	PreReleaseMetadata string
	VersionType        version.Type
	PreReleaseType     version.PreReleaseType
}
