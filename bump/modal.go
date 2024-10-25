package bump

import (
	"github.com/joe-at-startupmedia/version-bump/v2/git"
	"github.com/joe-at-startupmedia/version-bump/v2/gpg"
	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
	"net/http"
)

const (
	Version string = "2.1.0"
)

var GhRepoName = "joe-at-startupmedia/version-bump"
var GitConfigParser git.ConfigParserInterface
var GpgEntityAccessor gpg.EntityAccessorInterface
var GpgEntityReader gpg.EntityReaderInterface
var ReleaseGetter ReleaseGetterInterface

type ReleaseGetterInterface interface {
	Get(string) (*http.Response, error)
}

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

type versionBumpData struct {
	bump             *Bump
	versionsDetected *map[string]int
	runArgs          *RunArgs
	versionStr       string
}
