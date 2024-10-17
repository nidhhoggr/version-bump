package bump

import (
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
)

const (
	Version string = "2.0.1"
	Patch   int    = 3
	Minor   int    = 2
	Major   int    = 1
)

type Bump struct {
	FS            afero.Fs
	Git           GitConfig
	Configuration Configuration
}

type GitConfig struct {
	Repository Repository
	Worktree   Worktree
	GpgEntity  *openpgp.Entity
	UserName   string
	UserEmail  string
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
