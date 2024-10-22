package git

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

const (
	Username string = "username"
	Email    string = "username@domain.com"
)

type Instance struct {
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

func New(meta billy.Filesystem, data billy.Filesystem) (*Instance, error) {
	repo, err := git.Open(
		filesystem.NewStorage(meta, cache.NewObjectLRU(cache.DefaultMaxSize)),
		data,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error opening repository")
	}

	gitConfig, err := repo.ConfigScoped(config.GlobalScope)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving global git configuration")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving git worktree")
	}

	return &Instance{
		Repository: repo,
		Worktree:   worktree,
		Config:     gitConfig,
	}, nil
}

func (i *Instance) Save(files []string, version string, gpgEntity *openpgp.Entity) error {
	tm := time.Now()
	sign := &object.Signature{
		Name:  i.Config.User.Name,
		Email: i.Config.User.Email,
		When:  tm,
	}

	hash, err := i.Commit(files, version, sign, gpgEntity)
	if err != nil {
		return err
	}

	_, err = i.Repository.CreateTag(fmt.Sprintf("v%v", version), hash, &git.CreateTagOptions{
		Tagger:  sign,
		Message: version,
		SignKey: gpgEntity,
	})
	if err != nil {
		return errors.Wrap(err, "error tagging changes")
	}

	return nil
}

func (i *Instance) Commit(files []string, version string, sign *object.Signature, entity *openpgp.Entity) (plumbing.Hash, error) {
	for _, f := range files {
		_, err := i.Worktree.Add(f)
		if err != nil {
			return plumbing.Hash{}, errors.Wrapf(err, "error staging a file %v", f)
		}
	}
	hash, err := i.Worktree.Commit(version, &git.CommitOptions{
		All:       true,
		Author:    sign,
		Committer: sign,
		SignKey:   entity,
	})
	if err != nil {
		return plumbing.Hash{}, errors.Wrap(err, "error committing changes")
	}

	return hash, nil
}

func (i *Instance) GetSigningKeyFromConfig() (string, error) {

	shouldNotSign, gpgVerificationKey := getSigningKeyFromConfig(i.Config)

	if !shouldNotSign && gpgVerificationKey == "" {
		gitConfig, err := config.LoadConfig(config.GlobalScope)
		if err != nil {
			return "", errors.Wrap(err, "error loading git configuration from global scope")
		}
		_, gpgVerificationKey = getSigningKeyFromConfig(gitConfig)
	}

	return gpgVerificationKey, nil
}

func Init(meta billy.Filesystem, data billy.Filesystem) error {
	_, err := gogit.Init(
		filesystem.NewStorage(meta, cache.NewObjectLRU(cache.DefaultMaxSize)),
		data,
	)
	return err
}

func getSigningKeyFromConfig(config *config.Config) (bool, string) {

	var gpgVerificationKey string
	shouldNotSign := false

	commitSection := config.Raw.Section("commit")
	if commitSection != nil && commitSection.Options.Get("gpgsign") == "true" {
		userSection := config.Raw.Section("user")
		if userSection != nil {
			gpgVerificationKey = userSection.Options.Get("signingkey")
		}
	} else if commitSection != nil && commitSection.Options.Get("gpgsign") == "false" {
		shouldNotSign = true
	}

	return shouldNotSign, gpgVerificationKey
}
