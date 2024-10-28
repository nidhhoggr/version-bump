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

var (
	ErrStrOpeningRepo                   = "opening repository"
	ErrStrRetrievingGlobalConfiguration = "retrieving global git configuration"
	ErrStrRetrievingWorkTree            = "retrieving git worktree"
	ErrStrCommittingChanges             = "committing changes"
	ErrStrTaggingChanges                = "tagging changes"
	ErrStrLoadingConfiguration          = "loading git configuration from global scope"

	ErrStrFormattedStagingAFile = "staging a file %s"
)

const (
	Username string = "username"
	Email    string = "username@domain.com"
)

type Instance struct {
	Repository RepositoryInterface
	Worktree   WorktreeInterface
	Config     *config.Config
}

type RepositoryInterface interface {
	Worktree() (*git.Worktree, error)
	CreateTag(string, plumbing.Hash, *git.CreateTagOptions) (*plumbing.Reference, error)
	ConfigScoped(config.Scope) (*config.Config, error)
}

type WorktreeInterface interface {
	Add(string) (plumbing.Hash, error)
	Commit(string, *git.CommitOptions) (plumbing.Hash, error)
}

func New(meta billy.Filesystem, data billy.Filesystem) (*Instance, error) {
	repo, err := GetRepoFromFileSystem(meta, data)
	if err != nil {
		return nil, errors.Wrap(err, ErrStrOpeningRepo)
	}
	return GetInstanceFromRepo(repo)
}

func GetRepoFromFileSystem(meta billy.Filesystem, data billy.Filesystem) (*git.Repository, error) {
	return git.Open(
		filesystem.NewStorage(meta, cache.NewObjectLRU(cache.DefaultMaxSize)),
		data,
	)
}

func GetInstanceFromRepo(repo RepositoryInterface) (*Instance, error) {
	gitConfig, err := repo.ConfigScoped(config.GlobalScope)
	if err != nil {
		return nil, errors.Wrap(err, ErrStrRetrievingGlobalConfiguration)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, ErrStrRetrievingWorkTree)
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
		return errors.Wrap(err, ErrStrTaggingChanges)
	}

	return nil
}

func (i *Instance) Commit(files []string, version string, sign *object.Signature, entity *openpgp.Entity) (plumbing.Hash, error) {
	for _, f := range files {
		_, err := i.Worktree.Add(f)
		if err != nil {
			return plumbing.Hash{}, errors.Wrapf(err, ErrStrFormattedStagingAFile, f)
		}
	}
	hash, err := i.Worktree.Commit(version, &git.CommitOptions{
		All:       true,
		Author:    sign,
		Committer: sign,
		SignKey:   entity,
	})
	if err != nil {
		return plumbing.Hash{}, errors.Wrap(err, ErrStrCommittingChanges)
	}

	return hash, nil
}

func (i *Instance) GetSigningKeyFromConfig(configParser ConfigParserInterface) (string, error) {
	configParser.SetConfig(i.Config)
	shouldNotSign, gpgVerificationKey := getSigningKeyFromConfig(configParser)

	if !shouldNotSign && gpgVerificationKey == "" {
		gitConfig, err := configParser.LoadConfig(config.GlobalScope)
		if err != nil {
			return "", errors.Wrap(err, ErrStrLoadingConfiguration)
		}
		configParser.SetConfig(gitConfig)
		_, gpgVerificationKey = getSigningKeyFromConfig(configParser)
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

func getSigningKeyFromConfig(configParser ConfigParserInterface) (bool, string) {

	var gpgVerificationKey string
	shouldNotSign := false

	shouldGpgsign := configParser.GetSectionOption("commit", "gpgsign")
	if shouldGpgsign == "true" {
		gpgVerificationKey = configParser.GetSectionOption("user", "signingkey")
	} else if shouldGpgsign == "false" {
		shouldNotSign = true
	}

	return shouldNotSign, gpgVerificationKey
}
