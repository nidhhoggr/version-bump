package bump

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/joe-at-startupmedia/version-bump/v2/git"
	"github.com/joe-at-startupmedia/version-bump/v2/gpg"
	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type versionBumpData struct {
	bump       *Bump
	versionMap *map[string]int
	runArgs    *RunArgs
	versionStr string
}

var ConfigParser git.ConfigParserInterface

func init() {
	ConfigParser = new(git.ConfigParser)
}

func New(dir string) (*Bump, error) {
	fs := afero.NewOsFs()
	meta := osfs.New(path.Join(dir, ".git"))
	data := osfs.New(dir)
	return From(fs, meta, data, dir)
}

func From(fs afero.Fs, meta, data billy.Filesystem, dir string) (*Bump, error) {

	gitInstance, err := git.New(meta, data)
	if err != nil {
		return nil, err
	}

	o := &Bump{
		FS:  fs,
		Git: gitInstance,
	}

	dirs := []string{dir}

	// check for config file
	content, err := readFile(fs, ".bump")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "file does not exist") {
			//return default settings if config file not found
			return o.withConfiguration(dirs, true), nil
		} else {
			return nil, errors.Wrap(err, "error reading project config file")
		}
	}

	_, err = toml.Decode(strings.Join(content, "\n"), &o.withConfiguration(dirs, false).Configuration)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing project config file")
	}

	return o, nil
}

func (b *Bump) withConfiguration(dirs []string, enabledByDefault bool) *Bump {
	b.Configuration = Configuration{
		Docker: Language{
			Enabled:     enabledByDefault,
			Directories: dirs,
		},
		Go: Language{
			Enabled:     enabledByDefault,
			Directories: dirs,
		},
		JavaScript: Language{
			Enabled:     enabledByDefault,
			Directories: dirs,
		},
	}
	return b
}

func (b *Bump) Bump(ra *RunArgs) error {
	console.IncrementProjectVersion()

	versionMap := make(map[string]int)

	vbd := &versionBumpData{
		bump:       b,
		versionMap: &versionMap,
		runArgs:    ra,
	}

	bcr := reflect.ValueOf(b.Configuration)
	bcrType := bcr.Type()

	files := make([]string, 0)

	for i := 0; i < bcr.NumField(); i++ {
		langI := bcr.Field(i).Interface()
		lang := langI.(Language)
		langName := bcrType.Field(i).Name
		if lang.Enabled {
			//fmt.Printf("%s %-v", langName, lang)
			modifiedFiles, err := vbd.bumpComponent(langName, lang)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error incrementing version in %s project", langName))
			}
			files = append(files, modifiedFiles...)
		}
	}

	if len(*vbd.versionMap) > 1 {
		return errors.New("inconsistent versioning")
	} else if len(*vbd.versionMap) == 0 {
		return errors.New("0 files updated")
	}

	if len(files) != 0 {
		console.CommittingChanges()

		var gpgEntity *openpgp.Entity

		if ra.PassphrasePrompt != nil {
			gpgSigningKey, err := b.Git.GetSigningKeyFromConfig(ConfigParser)
			if err != nil {
				return errors.Wrap(err, "error retrieving gpg configuration")
			}
			if gpgSigningKey != "" {
				gpgEntity, err = vbd.passphrasePromptWithRetries(gpgSigningKey, 3, 0)
				if err != nil {
					return err
				}
			}
		}

		if err := b.Git.Save(files, vbd.versionStr, gpgEntity); err != nil {
			return errors.Wrap(err, "error committing changes")
		}
	}

	return nil
}

func (vbd *versionBumpData) passphrasePromptWithRetries(gpgSigningKey string, retryLimit int, retryCount int) (*openpgp.Entity, error) {
	if retryCount < retryLimit {
		keyPassphrase, err := vbd.runArgs.PassphrasePrompt()
		if err != nil {
			return nil, err
		}
		gpgEntity, err := gpg.GetGpgEntity(keyPassphrase, gpgSigningKey)
		if err != nil {
			return vbd.passphrasePromptWithRetries(gpgSigningKey, retryLimit, retryCount+1)
		} else {
			return gpgEntity, nil
		}
	} else {
		return nil, errors.New("could not validate gpg signing key")
	}
}

func (vbd *versionBumpData) bumpComponent(langName string, lang Language) ([]string, error) {

	files := make([]string, 0)

	for _, dir := range lang.Directories {
		f, err := getFiles(vbd.bump.FS, dir, lang.ExcludeFiles)
		if err != nil {
			return []string{}, errors.Wrap(err, "error listing directory files")
		}

		var langSettings = langs.Supported[langName]
		if langSettings == nil {
			return []string{}, errors.New(fmt.Sprintf("not supported language: %v", langName))
		}

		filteredFiles := filterFiles(langSettings.Files, f)

		if len(filteredFiles) > 0 {

			console.Language(langName)

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
			vbd.versionStr = oldVersion.String()

			if strings.Compare(oldVersionStr, vbd.versionStr) == 0 {
				//no changes in version
				continue
			}

			identified = true
			if vbd.runArgs.ConfirmationPrompt != nil {
				confirmed, err := vbd.runArgs.ConfirmationPrompt(oldVersionStr, vbd.versionStr, file)
				if err != nil {
					return []string{}, errors.Wrap(err, "error during confirmation prompt")
				} else if !confirmed {
					//return []string{}, errors.New("proposed version was denied")
					//continue allows scenarios where denying changes in specific file(s) is necessary
					continue
				}
			}

			(*vbd.versionMap)[oldVersionStr]++

			// set future version
			if lang.Regex != nil {
				newContent := make([]string, 0)

				for _, line := range fileContent {
					var added bool
					for _, expression := range *lang.Regex {
						regex := regexp.MustCompile(expression)
						if regex.MatchString(line) {
							l := strings.ReplaceAll(line, oldVersionStr, vbd.versionStr)
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
						newContent, err := sjson.Set(strings.Join(fileContent, "\n"), field, vbd.versionStr)
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

			if len(modifiedFiles) > 0 {
				fmt.Printf("Modified:%s\n", console.VersionUpdate(oldVersionStr, vbd.versionStr, filepath))
			}
		}
	}

	if len(files) > 0 && !identified {
		console.Error("    Version was not identified")
	}

	return modifiedFiles, nil
}
