package bump

import (
	"fmt"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"path"
	"regexp"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/joe-at-startupmedia/version-bump/v2/gpg"
	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func New(fs afero.Fs, meta, data billy.Filesystem, dir string, shouldSignCommits bool) (*Bump, error) {
	repo, err := git.Open(
		filesystem.NewStorage(meta, cache.NewObjectLRU(cache.DefaultMaxSize)),
		data,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error opening repository")
	}
	localGitConfig, err := repo.ConfigScoped(config.GlobalScope)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving global git configuration")
	}

	var gpgEntity *openpgp.Entity

	if shouldSignCommits {

		gpgSigningKey, err := gpg.GetSigningKeyFromConfig(localGitConfig)
		if err != nil {
			return nil, errors.Wrap(err, "error retrieving gpg configuration")
		}

		if gpgSigningKey != "" {
			gpgEntity, err = gpg.PromptForPassphrase(gpgSigningKey)
			if err != nil {
				return nil, errors.Wrap(err, "could not validate gpg signing key")
			}
		}
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving git worktree")
	}

	// NOTE: default config
	dirs := []string{dir}
	o := &Bump{
		FS: fs,
		Configuration: Configuration{
			Docker: Language{
				Enabled:     true,
				Directories: dirs,
			},
			Go: Language{
				Enabled:     true,
				Directories: dirs,
			},
			JavaScript: Language{
				Enabled:     true,
				Directories: dirs,
			},
		},
		Git: GitConfig{
			UserName:   localGitConfig.User.Name,
			UserEmail:  localGitConfig.User.Email,
			Repository: repo,
			Worktree:   worktree,
			GpgEntity:  gpgEntity,
		},
	}

	// check for config file
	content, err := readFile(fs, ".bump")
	if err != nil {
		// NOTE: return default settings if config file not found
		if strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "file does not exist") {
			return o, nil
		} else {
			return nil, errors.Wrap(err, "error reading project config file")
		}
	}

	// parse config file
	userConfig := new(Configuration)
	if err := toml.Unmarshal([]byte(strings.Join(content, "\n")), userConfig); err != nil {
		return nil, errors.Wrap(err, "error parsing project config file")
	}

	o.Configuration = Configuration{
		Docker: Language{
			Enabled:     userConfig.Docker.Enabled,
			Directories: dirs,
		},
		Go: Language{
			Enabled:     userConfig.Go.Enabled,
			Directories: dirs,
		},
		JavaScript: Language{
			Enabled:     userConfig.JavaScript.Enabled,
			Directories: dirs,
		},
	}

	if len(userConfig.Docker.Directories) != 0 {
		o.Configuration.Docker.Directories = userConfig.Docker.Directories
	}

	if len(userConfig.Go.Directories) != 0 {
		o.Configuration.Go.Directories = userConfig.Go.Directories
	}

	if len(userConfig.JavaScript.Directories) != 0 {
		o.Configuration.JavaScript.Directories = userConfig.JavaScript.Directories
	}

	if len(userConfig.Docker.ExcludeFiles) != 0 {
		o.Configuration.Docker.ExcludeFiles = userConfig.Docker.ExcludeFiles
	}

	if len(userConfig.Go.ExcludeFiles) != 0 {
		o.Configuration.Go.ExcludeFiles = userConfig.Go.ExcludeFiles
	}

	if len(userConfig.JavaScript.ExcludeFiles) != 0 {
		o.Configuration.JavaScript.ExcludeFiles = userConfig.JavaScript.ExcludeFiles
	}

	return o, nil
}

func (b *Bump) Bump(versionType version.Type) error {
	console.IncrementProjectVersion()

	versions := make(map[string]int)
	var newVersionStr string
	files := make([]string, 0)

	if b.Configuration.Docker.Enabled {
		modifiedFiles, err := b.bumpComponent(langs.Docker, b.Configuration.Docker, versionType, versions, &newVersionStr)
		if err != nil {
			return errors.Wrap(err, "error incrementing version in Docker project")
		}

		files = append(files, modifiedFiles...)
	}

	if b.Configuration.Go.Enabled {
		modifiedFiles, err := b.bumpComponent(langs.Go, b.Configuration.Go, versionType, versions, &newVersionStr)
		if err != nil {
			return errors.Wrap(err, "error incrementing version in Go project")
		}

		files = append(files, modifiedFiles...)
	}

	if b.Configuration.JavaScript.Enabled {
		modifiedFiles, err := b.bumpComponent(langs.JavaScript, b.Configuration.JavaScript, versionType, versions, &newVersionStr)
		if err != nil {
			return errors.Wrap(err, "error incrementing version in JavaScript project")
		}

		files = append(files, modifiedFiles...)
	}

	if len(versions) > 1 {
		return errors.New("inconsistent versioning")
	} else if len(versions) == 0 {
		return errors.New("0 files updated")
	}

	if len(files) != 0 {
		// TODO: update changelog
		console.CommittingChanges()

		if err := b.Git.Save(files, newVersionStr); err != nil {
			return errors.Wrap(err, "error committing changes")
		}
	}

	return nil
}

func (b *Bump) bumpComponent(name string, l Language, versionType version.Type, versions map[string]int, version *string) ([]string, error) {
	console.Language(name)
	files := make([]string, 0)

	for _, dir := range l.Directories {
		f, err := getFiles(b.FS, dir, l.ExcludeFiles)
		if err != nil {
			return []string{}, errors.Wrap(err, "error listing directory files")
		}

		langSettings := langs.New(name)
		if langSettings == nil {
			return []string{}, errors.New(fmt.Sprintf("not supported language: %v", name))
		}

		modifiedFiles, err := b.incrementVersion(
			dir,
			filterFiles(langSettings.Files, f),
			*langSettings,
			versionType,
			versions,
			version,
		)
		if err != nil {
			return []string{}, err
		}

		files = append(files, modifiedFiles...)
	}

	return files, nil
}

func (b *Bump) incrementVersion(dir string, files []string, lang langs.Language, versionType version.Type, versionMap map[string]int, versionString *string) ([]string, error) {
	var identified bool
	modifiedFiles := make([]string, 0)

	for _, file := range files {
		filepath := path.Join(dir, file)
		fileContent, err := readFile(b.FS, filepath)
		if err != nil {
			return []string{}, errors.Wrapf(err, "error reading a file %v", file)
		}
		var oldVersion *version.Version
		// get current version
		if lang.Regex != nil {
		outer:
			for _, line := range fileContent {
				for _, expression := range *lang.Regex {
					regex := regexp.MustCompile(expression)
					if regex.MatchString(line) {
						oldVersion, err = version.NewFromRegex(line, regex)
						if err != nil {
							return []string{}, errors.Wrapf(err, "error parsing semantic version at file %v from version: %s", filepath, oldVersion)
						}
						break outer
					}
				}
			}
		}

		if lang.JSONFields != nil {
			for _, field := range *lang.JSONFields {
				oldVersion, err = version.New(gjson.Get(strings.Join(fileContent, ""), field).String())
				if err != nil {
					return []string{}, errors.Wrapf(err, "error parsing semantic version at file %v", filepath)
				}
				break
			}
		}

		if oldVersion != nil {

			oldVersionStr := oldVersion.String()
			err := oldVersion.Increment(versionType, version.NotAPreRelease)
			if err != nil {
				return []string{}, errors.Wrapf(err, "error bumping version %v", filepath)
			}
			*versionString = oldVersion.String()
			console.VersionUpdate(oldVersionStr, *versionString, filepath)
			identified = true
			versionMap[oldVersionStr]++

			// set future version
			if lang.Regex != nil {
				newContent := make([]string, 0)

				for _, line := range fileContent {
					var added bool
					for _, expression := range *lang.Regex {
						regex := regexp.MustCompile(expression)
						if regex.MatchString(line) {
							l := strings.ReplaceAll(line, oldVersionStr, *versionString)
							newContent = append(newContent, l)
							added = true
						}
					}

					if !added {
						newContent = append(newContent, line)
					}
				}

				newContent = append(newContent, "")
				if err := writeFile(b.FS, filepath, strings.Join(newContent, "\n")); err != nil {
					return []string{}, errors.Wrapf(err, "error writing to file %v", filepath)
				}

				modifiedFiles = append(modifiedFiles, filepath)
			}

			if lang.JSONFields != nil {
				for _, field := range *lang.JSONFields {
					if gjson.Get(strings.Join(fileContent, ""), field).Exists() {
						newContent, err := sjson.Set(strings.Join(fileContent, "\n"), field, *versionString)
						if err != nil {
							return []string{}, errors.Wrapf(err, "error setting new version on content of a file %v", file)
						}

						if err := writeFile(b.FS, filepath, newContent); err != nil {
							return []string{}, errors.Wrapf(err, "error writing to file %v", filepath)
						}

						modifiedFiles = append(modifiedFiles, filepath)
					}
				}
			}
		}
	}

	if len(files) > 0 && !identified {
		console.Error("    Version was not identified")
	}

	return modifiedFiles, nil
}
