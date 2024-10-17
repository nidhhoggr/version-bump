package bump

import (
	"fmt"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

func (g *GitConfig) Save(files []string, version string) error {
	tm := time.Now()
	sign := &object.Signature{
		Name:  g.UserName,
		Email: g.UserEmail,
		When:  tm,
	}

	hash, err := Commit(files, version, sign, g.Worktree, g.GpgEntity)
	if err != nil {
		return err
	}

	_, err = g.Repository.CreateTag(fmt.Sprintf("v%v", version), hash, &git.CreateTagOptions{
		Tagger:  sign,
		Message: version,
		SignKey: g.GpgEntity,
	})
	if err != nil {
		return errors.Wrap(err, "error tagging changes")
	}

	return nil
}

func Commit(files []string, version string, sign *object.Signature, worktree Worktree, entity *openpgp.Entity) (plumbing.Hash, error) {
	for _, f := range files {
		_, err := worktree.Add(f)
		if err != nil {
			return plumbing.Hash{}, errors.Wrapf(err, "error staging a file %v", f)
		}
	}

	hash, err := worktree.Commit(version, &git.CommitOptions{
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
