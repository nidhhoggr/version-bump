package bump

import (
	"github.com/joe-at-startupmedia/version-bump/v2/git"
	"github.com/joe-at-startupmedia/version-bump/v2/gpg"
	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
	"net/http"
	"sync"
)

const (
	Version string = "2.1.0"
)

var GhRepoName = "joe-at-startupmedia/version-bump"

// #do better
var (
	GitConfigParser   git.ConfigParserInterface
	GpgEntityAccessor gpg.EntityAccessorInterface
	ReleaseGetter     ReleaseGetterInterface
)

type ReleaseGetterInterface interface {
	Get(string) (*http.Response, error)
}

type Bump struct {
	FS            afero.Fs
	Git           *git.Instance
	errChan       chan error
	waitGroup     *sync.WaitGroup
	Configuration Configuration
	mutex         sync.Mutex
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
	IsDryRun           bool
}

type versionBumpData struct {
	bump             *Bump
	versionsDetected VersionsDetected
	runArgs          *RunArgs
	versionStr       string
}

type stringedMap map[string]int

type VersionsDetected stringedMap
