package bump

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
)

const (
	Version string = "2.1.0"
)

type Bump struct {
	FS            afero.Fs
	Git           GitConfig
	Configuration Configuration
}

type GitConfig struct {
	Repository Repository
	Worktree   Worktree
	Config     *config.Config
}

type Repository interface {
	Worktree() (*git.Worktree, error)
	CreateTag(string, plumbing.Hash, *git.CreateTagOptions) (*plumbing.Reference, error)
}

type Worktree interface {
	Add(string) (plumbing.Hash, error)
	Commit(string, *git.CommitOptions) (plumbing.Hash, error)
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
