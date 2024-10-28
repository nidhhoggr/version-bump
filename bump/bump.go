package bump

import (
	"fmt"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/docker"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/golang"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/js"
	"path"
	"reflect"
	"regexp"
	"sort"
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

var (
	ErrStrNoSuchFileOrDirectory      = "no such file or directory"
	ErrStrFileDoesNotExist           = "file does not exist"
	ErrStrReadingConfigFile          = "reading project config file"
	ErrStrParsingConfigFile          = "parsing project config file"
	ErrStrZeroFilesUpdated           = "0 files updated"
	ErrStrRetrievingGpgConfiguration = "retrieving gpg configuration"
	ErrStrValidatingGpgSigningKey    = "validating gpg signing key"
	ErrStrListingDirectoryFiles      = "listing directory files"
	ErrStrDuringConfirmationPrompt   = "during confirmation prompt"

	ErrStrFormattedIncrementingInLangProject        = "incrementing version in %s project"
	ErrStrFormattedReadingAFile                     = "reading a file %v"
	ErrStrFormattedWritingToFile                    = "writing to file %v"
	ErrStrFormattedParsingVersionFromFileAndVersion = "parsing semantic version at file %v from version: %s"
	ErrStrFormattedBumpingVersion                   = "bumping version %v"
	ErrStrFormattedSettingVersionInFile             = "setting new version on content of a file %v"
	ErrStrFormattedInconsistentVersioning           = "inconsistent versioning %s"
)

func init() {
	GitConfigParser = new(git.ConfigParser)
	GpgEntityAccessor = &gpg.EntityAccessor{
		Reader: new(gpg.EntityReader),
	}
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
		if strings.Contains(err.Error(), ErrStrNoSuchFileOrDirectory) || strings.Contains(err.Error(), ErrStrFileDoesNotExist) {
			//return default settings if config file not found
			return o.withConfiguration(dirs, true), nil
		} else {
			return nil, errors.Wrap(err, ErrStrReadingConfigFile)
		}
	}

	cf := new(ConfigDecoder)
	_, err = toml.Decode(strings.Join(content, "\n"), cf)
	if err != nil {
		return nil, errors.Wrap(err, ErrStrParsingConfigFile)
	}

	//map ConfigDecoder to the Configuration struct
	bcr := reflect.ValueOf(cf).Elem()
	bcrType := bcr.Type()
	for i := 0; i < bcr.NumField(); i++ {
		langI := bcr.Field(i).Interface()
		lang := langI.(langs.Config)
		if lang.Enabled {
			lang.Name = bcrType.Field(i).Name
			o.Configuration = append(o.Configuration, lang)
		}
	}

	return o, nil
}

func (b *Bump) withConfiguration(dirs []string, enabledByDefault bool) *Bump {
	b.Configuration = Configuration{
		langs.Config{
			Name:        docker.Name,
			Enabled:     enabledByDefault,
			Directories: dirs,
		},
		langs.Config{
			Name:        golang.Name,
			Enabled:     enabledByDefault,
			Directories: dirs,
		},
		langs.Config{
			Name:        js.Name,
			Enabled:     enabledByDefault,
			Directories: dirs,
		},
	}
	return b
}

func (b *Bump) Bump(ra *RunArgs) error {
	console.IncrementProjectVersion(ra.IsDryRun)

	vbd := &versionBumpData{
		bump:             b,
		versionsDetected: NewVersionDetector(),
		runArgs:          ra,
	}

	files := make([]string, 0)

	for i := range b.Configuration {
		langConfig := b.Configuration[i]
		if langConfig.Enabled {
			//fmt.Printf("%s %-v", langName, lang)
			modifiedFiles, err := vbd.bumpComponent(langConfig, langs.Supported[langConfig.Name])
			if err != nil {
				return errors.Wrapf(err, ErrStrFormattedIncrementingInLangProject, langConfig.Name)
			}
			files = append(files, modifiedFiles...)
		}
	}

	versionsDetected := vbd.versionsDetected

	if len(versionsDetected) > 1 {
		return fmt.Errorf(ErrStrFormattedInconsistentVersioning, versionsDetected.String())
	} else if len(versionsDetected) == 0 {
		return errors.New(ErrStrZeroFilesUpdated)
	}

	if !ra.IsDryRun {

		if len(files) != 0 {
			console.CommittingChanges()

			var gpgEntity *openpgp.Entity

			if ra.PassphrasePrompt != nil {
				gpgSigningKey, err := b.Git.GetSigningKeyFromConfig(GitConfigParser)
				if err != nil {
					return errors.Wrap(err, ErrStrRetrievingGpgConfiguration)
				}
				if gpgSigningKey != "" {
					gpgEntity, err = vbd.passphrasePromptWithRetries(gpgSigningKey, 3, 0)
					if err != nil {
						return err
					}
				}
			}

			if err := b.Git.Save(files, vbd.versionStr, gpgEntity); err != nil {
				return err
			}
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
		gpgEntity, err := GpgEntityAccessor.GetEntity(keyPassphrase, gpgSigningKey)
		if err != nil {
			return vbd.passphrasePromptWithRetries(gpgSigningKey, retryLimit, retryCount+1)
		} else {
			return gpgEntity, nil
		}
	} else {
		return nil, errors.New(ErrStrValidatingGpgSigningKey)
	}
}

func (vbd *versionBumpData) bumpComponent(langConfig langs.Config, langSettings *langs.Settings) ([]string, error) {

	files := make([]string, 0)

	for _, dir := range langConfig.Directories {
		f, err := getFiles(vbd.bump.FS, dir, langConfig.ExcludeFiles)
		if err != nil {
			return []string{}, errors.Wrap(err, ErrStrListingDirectoryFiles)
		}

		filteredFiles := filterFiles(langSettings.Files, f)

		if len(filteredFiles) > 0 {

			console.Language(langConfig.Name, vbd.runArgs.IsDryRun)

			var versionIncrementor func(dir string, files []string, langSettings *langs.Settings) ([]string, error)

			if vbd.runArgs.IsDryRun {
				versionIncrementor = vbd.incrementVersionDryRun
			} else {
				versionIncrementor = vbd.incrementVersion
			}
			modifiedFiles, err := versionIncrementor(
				dir,
				filterFiles(langSettings.Files, f),
				langSettings,
			)
			if err != nil {
				return []string{}, err
			}

			files = append(files, modifiedFiles...)
		}
	}

	return files, nil
}

func (vbd *versionBumpData) incrementVersion(dir string, files []string, langSettings *langs.Settings) ([]string, error) {
	var identified bool
	modifiedFiles := make([]string, 0)

	for _, file := range files {
		filepath := path.Join(dir, file)
		fileContent, err := readFile(vbd.bump.FS, filepath)
		if err != nil {
			return []string{}, errors.Wrapf(err, ErrStrFormattedReadingAFile, file)
		}
		var oldVersion *version.Version
		// get current version
		if langSettings.Regex != nil {
		outer:
			for _, line := range fileContent {
				for _, expression := range *langSettings.Regex {
					regex := regexp.MustCompile(expression)
					if regex.MatchString(line) {
						oldVersion, err = version.NewFromRegex(line, regex)
						if err != nil {
							return []string{}, errors.Wrapf(err, ErrStrFormattedParsingVersionFromFileAndVersion, filepath, oldVersion)
						}
						break outer
					}
				}
			}
		}

		if langSettings.JSONFields != nil {
			for _, field := range *langSettings.JSONFields {
				oldVersion, err = version.New(gjson.Get(strings.Join(fileContent, ""), field).String())
				if err != nil {
					return []string{}, errors.Wrapf(err, ErrStrFormattedParsingVersionFromFileAndVersion, filepath, oldVersion)
				}
				break
			}
		}

		if oldVersion != nil {

			oldVersionStr := oldVersion.String()
			err = oldVersion.Increment(vbd.runArgs.VersionType, vbd.runArgs.PreReleaseType, vbd.runArgs.PreReleaseMetadata)
			if err != nil {
				return []string{}, errors.Wrapf(err, ErrStrFormattedBumpingVersion, filepath)
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
					return []string{}, errors.Wrap(err, ErrStrDuringConfirmationPrompt)
				} else if !confirmed {
					//return []string{}, errors.New("proposed version was denied")
					//continue allows scenarios where denying changes in specific file(s) is necessary
					continue
				}
			}

			(vbd.versionsDetected)[oldVersionStr]++

			if langSettings.Regex != nil {
				newContent := make([]string, 0)

				for _, line := range fileContent {
					var added bool
					for _, expression := range *langSettings.Regex {
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
					return []string{}, errors.Wrapf(err, ErrStrFormattedWritingToFile, filepath)
				}

				modifiedFiles = append(modifiedFiles, filepath)
			}

			if langSettings.JSONFields != nil {
				for _, field := range *langSettings.JSONFields {
					if gjson.Get(strings.Join(fileContent, ""), field).Exists() {
						newContent, err := sjson.Set(strings.Join(fileContent, "\n"), field, vbd.versionStr)
						if err != nil {
							return []string{}, errors.Wrapf(err, ErrStrFormattedSettingVersionInFile, file)
						}

						if err := writeFile(vbd.bump.FS, filepath, newContent); err != nil {
							return []string{}, errors.Wrapf(err, ErrStrFormattedWritingToFile, filepath)
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

func (vbd *versionBumpData) incrementVersionDryRun(dir string, files []string, langSettings *langs.Settings) ([]string, error) {
	var identified bool
	modifiedFiles := make([]string, 0)

	for _, file := range files {
		filepath := path.Join(dir, file)
		fileContent, err := readFile(vbd.bump.FS, filepath)
		if err != nil {
			return []string{}, errors.Wrapf(err, ErrStrFormattedReadingAFile, file)
		}
		var oldVersion *version.Version
		// get current version
		if langSettings.Regex != nil {
		outer:
			for _, line := range fileContent {
				for _, expression := range *langSettings.Regex {
					regex := regexp.MustCompile(expression)
					if regex.MatchString(line) {
						oldVersion, err = version.NewFromRegex(line, regex)
						if err != nil {
							return []string{}, errors.Wrapf(err, ErrStrFormattedParsingVersionFromFileAndVersion, filepath, oldVersion)
						}
						break outer
					}
				}
			}
		}

		//@TODO Are we gonna do one the other or both?
		if langSettings.JSONFields != nil {
			for _, field := range *langSettings.JSONFields {
				oldVersion, err = version.New(gjson.Get(strings.Join(fileContent, ""), field).String())
				if err != nil {
					return []string{}, errors.Wrapf(err, ErrStrFormattedParsingVersionFromFileAndVersion, filepath, oldVersion)
				}
				break
			}
		}

		if oldVersion != nil {

			oldVersionStr := oldVersion.String()
			(vbd.versionsDetected)[oldVersionStr]++
			err = oldVersion.Increment(vbd.runArgs.VersionType, vbd.runArgs.PreReleaseType, vbd.runArgs.PreReleaseMetadata)
			if err != nil {
				return []string{}, errors.Wrapf(err, ErrStrFormattedBumpingVersion, filepath)
			}
			vbd.versionStr = oldVersion.String()

			if strings.Compare(oldVersionStr, vbd.versionStr) == 0 {
				//no changes in version
				continue
			}

			identified = true

			if langSettings.Regex != nil {
				//newContent := make([]string, 0)

				for _, line := range fileContent {
					//var added bool
					for _, expression := range *langSettings.Regex {
						regex := regexp.MustCompile(expression)
						if regex.MatchString(line) {
							fmt.Printf("lang.Regex:\n\tfile: %s/%s\n\tline: %s\n\tcurrent version: %s\n\tnew version %s", dir, file, line, oldVersionStr, vbd.versionStr)
							//l := strings.ReplaceAll(line, oldVersionStr, vbd.versionStr)
							//newContent = append(newContent, l)
							//added = true
						}
					}

					//if !added {
					//	newContent = append(newContent, line)
					//}
				}

				//newContent = append(newContent, "")
				//if err = writeFile(vbd.bump.FS, filepath, strings.Join(newContent, "\n")); err != nil {
				//	return []string{}, errors.Wrapf(err, ErrStrFormattedWritingToFile, filepath)
				//}

				modifiedFiles = append(modifiedFiles, filepath)
			}

			if langSettings.JSONFields != nil {
				for _, field := range *langSettings.JSONFields {
					if gjson.Get(strings.Join(fileContent, ""), field).Exists() {
						fmt.Printf("lang.JSONFields:\n\tfile: %s/%s\n\tfield: %s\n\tcurrent version: %s\n\tnew version %s", dir, file, field, oldVersionStr, vbd.versionStr)
						//newContent, err := sjson.Set(strings.Join(fileContent, "\n"), field, vbd.versionStr)
						//
						//if err != nil {
						//	return []string{}, errors.Wrapf(err, ErrStrFormattedSettingVersionInFile, file)
						//}
						//
						//if err := writeFile(vbd.bump.FS, filepath, newContent); err != nil {
						//	return []string{}, errors.Wrapf(err, ErrStrFormattedWritingToFile, filepath)
						//}
						//
						modifiedFiles = append(modifiedFiles, filepath)
					}
				}
			}

			if len(modifiedFiles) > 0 {
				fmt.Printf("Will Modify: %s\n", console.VersionUpdate(oldVersionStr, vbd.versionStr, filepath))
			}
		}
	}

	if len(files) > 0 && !identified {
		console.Error("    Version was not identified")
	}

	return modifiedFiles, nil
}

func NewVersionDetector() VersionsDetected {
	return (VersionsDetected)(make(stringedMap))
}

func (vd *VersionsDetected) String() string {
	var vdStr []string
	for key := range *vd {
		vdStr = append(vdStr, key)
	}
	sort.Strings(vdStr)
	return strings.Join(vdStr, ",")
}
