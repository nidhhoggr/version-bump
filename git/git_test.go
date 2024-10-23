package git_test

import (
	"fmt"
	"github.com/go-git/go-git/v5/config"
	"github.com/joe-at-startupmedia/version-bump/v2/git"

	gogit "github.com/go-git/go-git/v5"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/joe-at-startupmedia/version-bump/v2/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGit_Save(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Version            string
		Files              []string
		MockWorktreeError  error
		MockCommitOutput   plumbing.Hash
		MockCommitError    error
		MockCreateTagError error
		ExpectedError      string
	}

	suite := map[string]test{
		"Success": {
			Version: "1.0.0",
			Files: []string{
				"file-1.txt",
				"file-2.txt",
			},
			MockWorktreeError:  nil,
			MockCommitOutput:   plumbing.NewHash("abc"),
			MockCommitError:    nil,
			MockCreateTagError: nil,
			ExpectedError:      "",
		},
		"Error Tagging Commit": {
			Version: "1.0.0",
			Files: []string{
				"file.txt",
			},
			MockWorktreeError:  nil,
			MockCommitOutput:   plumbing.NewHash("abc"),
			MockCommitError:    nil,
			MockCreateTagError: errors.New("reason"),
			ExpectedError:      "error tagging changes: reason",
		},
		"Error Committing Changes": {
			Version: "1.0.0",
			Files: []string{
				"file.txt",
			},
			MockWorktreeError:  nil,
			MockCommitOutput:   plumbing.NewHash("abc"),
			MockCommitError:    errors.New("reason"),
			MockCreateTagError: nil,
			ExpectedError:      "error committing changes: reason",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		m1 := new(mocks.Repository)
		m2 := new(mocks.Worktree)

		for _, f := range test.Files {
			m2.On("Add", f).Return(nil, nil).Once()
		}

		m2.On("Commit", test.Version, mock.AnythingOfType("*git.CommitOptions")).Return(test.MockCommitOutput, test.MockCommitError).Once()

		m1.On("CreateTag", fmt.Sprintf("v%v", test.Version), test.MockCommitOutput, mock.AnythingOfType("*git.CreateTagOptions")).Return(nil, test.MockCreateTagError).Once()

		gitConfig := &config.Config{}
		gitConfig.User.Name = git.Username
		gitConfig.User.Email = git.Email

		receiver := &git.Instance{
			Config:     gitConfig,
			Repository: m1,
			Worktree:   m2,
		}

		err := receiver.Save(test.Files, test.Version, nil)
		if test.ExpectedError != "" || err != nil {
			a.EqualError(err, test.ExpectedError)
		}
	}
}

func TestGit_Commit(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Version         string
		Files           []string
		MockAddError    error
		MockCommitHash  string
		MockCommitError error
		ExpectedError   string
	}

	suite := map[string]test{
		"Multiple Files Changed": {
			Version: "1.0.0",
			Files: []string{
				"file-1.txt",
				"file-2.txt",
			},
			MockAddError:    nil,
			MockCommitHash:  "abc",
			MockCommitError: nil,
			ExpectedError:   "",
		},
		"Stage Error": {
			Version: "1.0.0",
			Files: []string{
				"file.txt",
			},
			MockAddError:    errors.New("reason"),
			MockCommitHash:  "",
			MockCommitError: nil,
			ExpectedError:   "error staging a file file.txt: reason",
		},
		"Commit Error": {
			Version: "1.0.0",
			Files: []string{
				"file.txt",
			},
			MockAddError:    nil,
			MockCommitHash:  "abc",
			MockCommitError: errors.New("reason"),
			ExpectedError:   "error committing changes: reason",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		s := &object.Signature{
			Name:  git.Username,
			Email: git.Email,
			When:  time.Now(),
		}
		m1 := new(mocks.Repository)
		m2 := new(mocks.Worktree)

		for _, f := range test.Files {
			m2.On("Add", f).Return(nil, test.MockAddError).Once()
		}

		m2.On("Commit", test.Version, &gogit.CommitOptions{
			All:       true,
			Author:    s,
			Committer: s,
		}).Return(plumbing.NewHash(test.MockCommitHash), test.MockCommitError).Once()

		gitConfig := &config.Config{}
		gitConfig.User.Name = git.Username
		gitConfig.User.Email = git.Email
		gitInstance := &git.Instance{
			Config:     gitConfig,
			Repository: m1,
			Worktree:   m2,
		}

		h, err := gitInstance.Commit(test.Files, test.Version, s, nil)
		if test.ExpectedError != "" || err != nil {
			a.EqualError(err, test.ExpectedError)
			a.Equal(plumbing.NewHash(""), h)
		} else {
			a.Equal(plumbing.NewHash(test.MockCommitHash), h)
		}
	}
}

func TestGit_ConfigParser(t *testing.T) {
	a := assert.New(t)

	cfg := config.NewConfig()

	input := []byte(`[core]
		bare = true
		worktree = foo
		commentchar = bar
[user]
		name = John Doe
		email = john@example.com`)

	err := cfg.Unmarshal(input)
	a.Empty(err)

	cp := new(git.ConfigParser)
	cp.SetConfig(cfg)
	ok, username := cp.GetSectionOption("user", "name")
	a.True(ok)
	a.Equal("John Doe", username)
}
