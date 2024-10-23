package bump

import (
	"github.com/joe-at-startupmedia/version-bump/v2/git"
	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
)

const (
	Version string = "2.1.0"
)

type Bump struct {
	FS            afero.Fs
	Git           *git.Instance
	Configuration Configuration
}

type ConfigDecoder struct {
	Docker     langs.Config
	Go         langs.Config
	JavaScript langs.Config
}

type Configuration []langs.Config

type RunArgs struct {
	ConfirmationPrompt func(string, string, string) (bool, error)
	PassphrasePrompt   func() (string, error)
	PreReleaseMetadata string
	VersionType        version.Type
	PreReleaseType     version.PreReleaseType
}
