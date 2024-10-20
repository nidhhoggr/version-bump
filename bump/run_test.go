package bump_test

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/joe-at-startupmedia/version-bump/v2/mocks"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"path"
	"testing"

	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	"github.com/stretchr/testify/assert"
)

func getBumpInstance() *bump.Bump {

	var testSuite = testBumpTestSuite{
		Version: "1.3.0",
		Configuration: bump.Configuration{
			Go: bump.Language{
				Enabled:     true,
				Directories: []string{"."},
			},
		},
		Files: allFiles{
			Go: map[string][]file{
				".": {
					{
						Name:                "main.go",
						ExpectedToBeChanged: true,
						Content: `package main

import "fmt"

const Version string = "1.2.4"

func main() {
	fmt.Println(Version)
}`,
					},
				},
			},
		},
		VersionType:        version.Minor,
		MockAddError:       nil,
		MockCommitError:    nil,
		MockCreateTagError: nil,
		ExpectedError:      "",
	}

	m1 := new(mocks.Repository)
	m2 := new(mocks.Worktree)

	r := bump.Bump{
		FS: afero.NewMemMapFs(),
		Git: bump.GitConfig{
			UserName:   username,
			UserEmail:  email,
			Repository: m1,
			Worktree:   m2,
		},
		Configuration: testSuite.Configuration,
	}

	for _, dir := range testSuite.Configuration.Go.Directories {
		for tgtDir, tgtFiles := range testSuite.Files.Go {
			if dir == tgtDir {
				for _, tgtFile := range tgtFiles {
					f, _ := r.FS.Create(path.Join(dir, tgtFile.Name))
					_, _ = f.WriteString(tgtFile.Content)
				}
			}
		}
	}

	for dir, files := range testSuite.Files.Go {
		for _, file := range files {
			if file.ExpectedToBeChanged {
				var f string
				if dir == "." {
					f = file.Name
				} else {
					f = path.Join(dir, file.Name)
				}
				m2.On("Add", f).Return(nil, testSuite.MockAddError).Once()
			}
		}
	}

	hash := plumbing.NewHash("abc")

	m2.On(
		"Commit", testSuite.Version, mock.AnythingOfType("*git.CommitOptions"),
	).Return(hash, testSuite.MockCommitError).Once()

	m1.On(
		"CreateTag", fmt.Sprintf("v%v", testSuite.Version), hash, mock.AnythingOfType("*git.CreateTagOptions"),
	).Return(nil, testSuite.MockCreateTagError).Once()

	return &r
}

func TestBumpRun(t *testing.T) {
	a := assert.New(t)

	bump.GhRepoName = "anton-yurchenko/version-bump"
	b := getBumpInstance()
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.Nil(err)
}

func TestBumpRunWithFailingUrl(t *testing.T) {
	a := assert.New(t)

	bump.GhRepoName = "nonexistent-user/nonexistent-package"
	b := getBumpInstance()
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.ErrorContains(err, "status code was not success: 404")
}
