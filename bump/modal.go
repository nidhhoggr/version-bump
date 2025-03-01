package bump

import (
	"github.com/nidhhoggr/version-bump/v2/git"
	"github.com/nidhhoggr/version-bump/v2/gpg"
	"github.com/nidhhoggr/version-bump/v2/langs"
	"github.com/nidhhoggr/version-bump/v2/version"
	"github.com/spf13/afero"
	"net/http"
	"sync"
)

const (
	Version string = "2.1.3"
)

var GhRepoName = "nidhhoggr/version-bump"

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
	FS                      afero.Fs
	Git                     *git.Instance
	errChanVersionGathering chan error
	errChanPostProcessing   chan error
	WaitGroup               *sync.WaitGroup
	Configuration           Configuration
	mutex                   sync.Mutex
}

type Configuration []langs.Config

type RunArgs struct {
	ConfirmationPrompt func(string, string, string) (bool, error)
	PassphrasePrompt   func() (string, error)
	PrereleaseMetadata string
	VersionType        version.Type
	PrereleaseType     version.PrereleaseType
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
