package bump_test

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/joe-at-startupmedia/version-bump/v2/git"
	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/golang"
	"github.com/joe-at-startupmedia/version-bump/v2/mocks"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"path"
	"testing"

	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	"github.com/stretchr/testify/assert"
)

var runTestSuites = []testBumpTestSuite{
	{
		Version: "1.3.0",
		Configuration: bump.Configuration{
			langs.Config{
				Name:        golang.Name,
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
	},
	{
		Version: "1.3.0-beta",
		Configuration: bump.Configuration{
			langs.Config{
				Name:        golang.Name,
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
	},
}

func getBumpInstance(testSuite testBumpTestSuite) *bump.Bump {

	m1 := new(mocks.Repository)
	m2 := new(mocks.Worktree)

	gitConfig := &config.Config{}
	gitConfig.User.Name = git.Username
	gitConfig.User.Email = git.Email

	r := bump.Bump{
		FS: afero.NewMemMapFs(),
		Git: &git.Instance{
			Config:     gitConfig,
			Repository: m1,
			Worktree:   m2,
		},
		Configuration: testSuite.Configuration,
	}

	for _, dir := range testSuite.Configuration[0].Directories {
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

	b := getBumpInstance(runTestSuites[0])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.Nil(err)
}

func TestBumpRun_Fails(t *testing.T) {
	a := assert.New(t)

	b := getBumpInstance(runTestSuites[1])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.NotAVersion,
		PreReleaseType: version.AlphaPreRelease,
	})
	a.ErrorContains(err, "cannot prerelease a non-prerelease without incrementing a version type")
}

func TestBumpRun_FailingUrl(t *testing.T) {
	a := assert.New(t)

	bump.GhRepoName = "nonexistent-user/nonexistent-package"
	b := getBumpInstance(runTestSuites[0])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.ErrorContains(err, "status code was not success: 404")
}

const nonExistentReleaseUrl = "https://api.github.com/repos/nonexistent-user/nonexistent-package/releases/latest"

func TestBumpRun_GetterHasError(t *testing.T) {
	a := assert.New(t)

	rg := new(mocks.ReleaseGetterMock)
	rg.On("Get", nonExistentReleaseUrl).Return(0, "", errors.New("mock scenario 1 with error"))
	bump.ReleaseGetter = rg
	b := getBumpInstance(runTestSuites[0])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.ErrorContains(err, "mock scenario 1 with error")
	rg.AssertExpectations(t)
}

func TestBumpRun_GetterHasJunkJson(t *testing.T) {
	a := assert.New(t)

	rg := new(mocks.ReleaseGetterMock)
	rg.On("Get", nonExistentReleaseUrl).Return(200, "{\"invalid_json\":", nil)
	bump.ReleaseGetter = rg
	b := getBumpInstance(runTestSuites[0])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.ErrorContains(err, "unexpected EOF")
	rg.AssertExpectations(t)
}

func TestBumpRun_GetterHasNoTagName(t *testing.T) {
	a := assert.New(t)

	rg := new(mocks.ReleaseGetterMock)
	rg.On("Get", nonExistentReleaseUrl).Return(200, "{\"tag_name_wrong\":\"\"}", nil)
	bump.ReleaseGetter = rg
	b := getBumpInstance(runTestSuites[0])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.ErrorContains(err, "tag name from request was empty")
	rg.AssertExpectations(t)
}

func TestBumpRun_GetterShouldPresentNewVersion(t *testing.T) {
	a := assert.New(t)

	rg := new(mocks.ReleaseGetterMock)
	rg.On("Get", nonExistentReleaseUrl).Return(200, "{\"tag_name\":\"v4.0.0\"}", nil)
	bump.ReleaseGetter = rg
	b := getBumpInstance(runTestSuites[0])
	err := b.Run(&bump.RunArgs{
		VersionType:    version.Minor,
		PreReleaseType: version.NotAPreRelease,
	})
	a.Empty(err)
	rg.AssertExpectations(t)
}
