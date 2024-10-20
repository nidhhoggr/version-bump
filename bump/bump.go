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

type versionBumpData struct {
	bump       *Bump
	versionMap map[string]int
	versionStr *string
	runArgs    *RunArgs
}

func New(fs afero.Fs, meta, data billy.Filesystem, dir string, passphrasePrompt func() (string, error)) (*Bump, error) {
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

	if passphrasePrompt != nil {

		gpgSigningKey, err := gpg.GetSigningKeyFromConfig(localGitConfig)
		if err != nil {
			return nil, errors.Wrap(err, "error retrieving gpg configuration")
		}

		if gpgSigningKey != "" {
			keyPassphrase, err := passphrasePrompt()
			if err != nil {
				return nil, err
			}
			gpgEntity, err = gpg.GetGpgEntity(keyPassphrase, gpgSigningKey)
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

func (b *Bump) Bump(ra *RunArgs) error {
	console.IncrementProjectVersion()

	versionMap := make(map[string]int)
	var newVersionStr string
	files := make([]string, 0)

	vbd := &versionBumpData{
		b,
		versionMap,
		&newVersionStr,
		ra,
	}

	if b.Configuration.Docker.Enabled {
		modifiedFiles, err := vbd.bumpComponent(langs.Docker, b.Configuration.Docker)
		if err != nil {
			return errors.Wrap(err, "error incrementing version in Docker project")
		}
		files = append(files, modifiedFiles...)
	}

	if b.Configuration.Go.Enabled {
		modifiedFiles, err := vbd.bumpComponent(langs.Go, b.Configuration.Go)
		if err != nil {
			return errors.Wrap(err, "error incrementing version in Go project")
		}
		files = append(files, modifiedFiles...)
	}

	if b.Configuration.JavaScript.Enabled {
		modifiedFiles, err := vbd.bumpComponent(langs.JavaScript, b.Configuration.JavaScript)
		if err != nil {
			return errors.Wrap(err, "error incrementing version in JavaScript project")
		}
		files = append(files, modifiedFiles...)
	}

	if len(versionMap) > 1 {
		return errors.New("inconsistent versioning")
	} else if len(versionMap) == 0 {
		return errors.New("0 files updated")
	}

	if len(files) != 0 {
		console.CommittingChanges()
		if ra.ConfirmationPrompt != nil {
			confirmed, err := ra.ConfirmationPrompt(newVersionStr)
			if err != nil {
				return errors.Wrap(err, "error during confirmation prompt")
			} else if !confirmed {
				return errors.New("proposed version was denied")
			}
		}
		if err := b.Git.Save(files, newVersionStr); err != nil {
			return errors.Wrap(err, "error committing changes")
		}
	}

	return nil
}

func (vbd *versionBumpData) bumpComponent(langName string, lang Language) ([]string, error) {
	console.Language(langName)
	files := make([]string, 0)

	for _, dir := range lang.Directories {
		f, err := getFiles(vbd.bump.FS, dir, lang.ExcludeFiles)
		if err != nil {
			return []string{}, errors.Wrap(err, "error listing directory files")
		}

		langSettings := langs.New(langName)
		if langSettings == nil {
			return []string{}, errors.New(fmt.Sprintf("not supported language: %v", langName))
		}

		modifiedFiles, err := vbd.incrementVersion(
			dir,
			filterFiles(langSettings.Files, f),
			*langSettings,
		)
		if err != nil {
			return []string{}, err
		}

		files = append(files, modifiedFiles...)
	}

	return files, nil
}

func (vbd *versionBumpData) incrementVersion(dir string, files []string, lang langs.Language) ([]string, error) {
	var identified bool
	modifiedFiles := make([]string, 0)

	for _, file := range files {
		filepath := path.Join(dir, file)
		fileContent, err := readFile(vbd.bump.FS, filepath)
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
			err = oldVersion.Increment(vbd.runArgs.VersionType, vbd.runArgs.PreReleaseType, vbd.runArgs.PreReleaseMetadata)
			if err != nil {
				return []string{}, errors.Wrapf(err, "error bumping version %v", filepath)
			}
			*vbd.versionStr = oldVersion.String()
			console.VersionUpdate(oldVersionStr, *vbd.versionStr, filepath)
			identified = true
			vbd.versionMap[oldVersionStr]++

			// set future version
			if lang.Regex != nil {
				newContent := make([]string, 0)

				for _, line := range fileContent {
					var added bool
					for _, expression := range *lang.Regex {
						regex := regexp.MustCompile(expression)
						if regex.MatchString(line) {
							l := strings.ReplaceAll(line, oldVersionStr, *vbd.versionStr)
							newContent = append(newContent, l)
							added = true
						}
					}

					if !added {
						newContent = append(newContent, line)
					}
				}

				newContent = append(newContent, "")
				if err = writeFile(vbd.bump.FS, filepath, strings.Join(newContent, "\n")); err != nil {
					return []string{}, errors.Wrapf(err, "error writing to file %v", filepath)
				}

				modifiedFiles = append(modifiedFiles, filepath)
			}

			if lang.JSONFields != nil {
				for _, field := range *lang.JSONFields {
					if gjson.Get(strings.Join(fileContent, ""), field).Exists() {
						newContent, err := sjson.Set(strings.Join(fileContent, "\n"), field, *vbd.versionStr)
						if err != nil {
							return []string{}, errors.Wrapf(err, "error setting new version on content of a file %v", file)
						}

						if err := writeFile(vbd.bump.FS, filepath, newContent); err != nil {
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
