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
	"sync"

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
		FS:        fs,
		Git:       gitInstance,
		WaitGroup: new(sync.WaitGroup),
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

	console.Debug("Bump.From()", fmt.Sprintf("configuration: %-v", o))

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

	console.DebuggingEnabled = ra.ShouldDebug
	console.IncrementProjectVersion(ra.IsDryRun)

	vbd := &versionBumpData{
		bump:             b,
		versionsDetected: NewVersionDetector(),
		runArgs:          ra,
	}

	files := make([]string, 0)

	b.errChanVersionGathering = make(chan error, 1)
	b.errChanPostProcessing = make(chan error, 1)

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

	var err error
	if len(versionsDetected) > 1 {
		err = fmt.Errorf(ErrStrFormattedInconsistentVersioning, versionsDetected.String())
	} else if len(versionsDetected) == 0 {
		err = errors.New(ErrStrZeroFilesUpdated)
	}

	b.errChanVersionGathering <- err
	//notify the goroutines we're ready to start processing if err is nil
	close(b.errChanVersionGathering)
	//wait for all the goroutines to finish
	b.WaitGroup.Wait()

	if err != nil {
		return err
	}

	select {
	case err = <-b.errChanPostProcessing:
		return err
		//a goroutine errored
	default:
		//all goroutines finished successfully
	}

	if !ra.IsDryRun {

		if len(files) != 0 {

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

			console.CommittingChanges()

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

	langDirs := langConfig.Directories

	//
	if len(langDirs) == 0 {
		langDirs = []string{"."}
	}

	for _, dir := range langDirs {

		console.Debug("Bump.bumpComponent()", fmt.Sprintf("lang: %s, dir: %s\n", langSettings.Name, dir))
		f, err := getFiles(vbd.bump.FS, dir, langConfig.ExcludeFiles)
		if err != nil {
			return []string{}, errors.Wrap(err, ErrStrListingDirectoryFiles)
		}

		langFiles := langSettings.Files

		if len(langConfig.Files) > 0 {
			langFiles = langConfig.Files
		}

		if len(langConfig.Regex) > 0 {
			langSettings.Regex = &langConfig.Regex
		}

		if len(langConfig.JSONFields) > 0 {
			langSettings.JSONFields = &langConfig.JSONFields
		}

		filteredFiles := filterFiles(langFiles, f)

		console.Debug("Bump.bumpComponent()", fmt.Sprintf("langfiles: %-v, f: %-v, filteredFiled: %-v\n", langSettings.Files, f, filteredFiles))

		if len(filteredFiles) > 0 {

			modifiedFiles, err := vbd.incrementVersion(
				dir,
				filteredFiles,
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
		console.Debug("Bump.incrementVersion()", fmt.Sprintf("lang: %s, file: %s, dir: %s\n", langSettings.Name, file, dir))
		fileContent, err := readFile(vbd.bump.FS, filepath)
		if err != nil {
			return []string{}, errors.Wrapf(err, ErrStrFormattedReadingAFile, file)
		}
		var oldVersion *version.Version
		// get current version
		if langSettings.Regex != nil {
		outerR:
			for lineNumber, line := range fileContent {
				for _, expression := range *langSettings.Regex {
					regex := regexp.MustCompile(expression)
					if regex.MatchString(line) {
						oldVersion, err = version.NewFromRegex(line, regex)
						if err != nil {
							return []string{}, errors.Wrapf(err, ErrStrFormattedParsingVersionFromFileAndVersion, fmt.Sprintf("%s %s %s", filepath, line, regex), oldVersion)
						} else if oldVersion != nil {

							oldVersionStr := oldVersion.String()

							versionsAreSame, err := vbd.incrementAndCompareVersions(oldVersion)

							if err != nil {
								return []string{}, errors.Wrapf(err, ErrStrFormattedBumpingVersion, filepath)
							} else if versionsAreSame {
								continue outerR
							}

							identified = true

							if !vbd.runArgs.IsDryRun {
								confirmed, err := vbd.versionConfirmationPrompt(oldVersionStr, file)
								if err != nil {
									return []string{}, errors.Wrap(err, ErrStrDuringConfirmationPrompt)
								} else if !confirmed {
									//continue allows scenarios where denying changes in specific file(s) is necessary
									continue
								}
							}

							go vbd.runRegexReplacement(langSettings, fileContent, lineNumber, filepath, oldVersionStr)

							modifiedFiles = append(modifiedFiles, filepath)
						}
						break outerR
					}
				}
			}
		}

		if langSettings.JSONFields != nil {
		outerJ:
			for _, field := range *langSettings.JSONFields {
				matched := gjson.Get(strings.Join(fileContent, ""), field).String()
				if matched == "" {
					continue
				}
				oldVersion, err = version.New(matched)
				if err != nil {
					return []string{}, errors.Wrapf(err, ErrStrFormattedParsingVersionFromFileAndVersion, filepath, oldVersion)
				} else if oldVersion != nil {

					oldVersionStr := oldVersion.String()

					versionsAreSame, err := vbd.incrementAndCompareVersions(oldVersion)

					if err != nil {
						return []string{}, errors.Wrapf(err, ErrStrFormattedBumpingVersion, filepath)
					} else if versionsAreSame {
						continue outerJ
					}

					identified = true

					if !vbd.runArgs.IsDryRun {
						confirmed, err := vbd.versionConfirmationPrompt(oldVersionStr, file)
						if err != nil {
							return []string{}, errors.Wrap(err, ErrStrDuringConfirmationPrompt)
						} else if !confirmed {
							//continue allows scenarios where denying changes in specific file(s) is necessary
							continue
						}
					}

					go vbd.runJsonFieldReplacement(langSettings, fileContent, field, filepath, oldVersionStr)

					modifiedFiles = append(modifiedFiles, filepath)
				}
				break outerJ
			}
		}
	}

	if len(files) > 0 && !identified {
		console.Error("    Version was not identified")
	}

	return modifiedFiles, nil
}

func (vbd *versionBumpData) incrementAndCompareVersions(oldVersion *version.Version) (bool, error) {
	oldVersionStr := oldVersion.String()
	vbd.versionsDetected[oldVersionStr]++
	err := oldVersion.Increment(vbd.runArgs.VersionType, vbd.runArgs.PreReleaseType, vbd.runArgs.PreReleaseMetadata)
	if err != nil {
		return false, err
	}
	vbd.versionStr = oldVersion.String()
	if strings.Compare(oldVersionStr, vbd.versionStr) == 0 {
		//no changes in version
		return true, nil
	}
	return false, nil
}

func (vbd *versionBumpData) runRegexReplacement(langSettings *langs.Settings, fileContent []string, lineNumber int, filepath string, oldVersionStr string) {
	err := <-vbd.bump.errChanVersionGathering
	if err != nil {
		return
	}

	vbd.bump.WaitGroup.Add(1)

	vbd.bump.mutex.Lock()

	console.Language(langSettings.Name, vbd.runArgs.IsDryRun)

	line := fileContent[lineNumber]

	if !vbd.runArgs.IsDryRun {
		console.Debug("Bump.runRegexReplacement()", fmt.Sprintf("line: %s\n", line))
		replacedLine := strings.ReplaceAll(line, oldVersionStr, vbd.versionStr)
		fileContent[lineNumber] = strings.ReplaceAll(fileContent[lineNumber], line, replacedLine)
		fileContent = append(fileContent, "")
		if err := writeFile(vbd.bump.FS, filepath, strings.Join(fileContent, "\n")); err != nil {
			vbd.bump.errChanPostProcessing <- errors.Wrapf(err, ErrStrFormattedWritingToFile, filepath)
		}
	}

	console.VersionUpdateLine(oldVersionStr, vbd.versionStr, filepath, line)

	vbd.bump.mutex.Unlock()

	vbd.bump.WaitGroup.Done()

	return
}

func (vbd *versionBumpData) runJsonFieldReplacement(langSettings *langs.Settings, fileContent []string, field string, filepath string, oldVersionStr string) {

	err := <-vbd.bump.errChanVersionGathering
	if err != nil {
		return
	}

	vbd.bump.WaitGroup.Add(1)

	vbd.bump.mutex.Lock()

	console.Language(langSettings.Name, vbd.runArgs.IsDryRun)

	if !vbd.runArgs.IsDryRun {
		if gjson.Get(strings.Join(fileContent, ""), field).Exists() {
			newContent, err := sjson.Set(strings.Join(fileContent, "\n"), field, vbd.versionStr)
			if err != nil {
				vbd.bump.errChanPostProcessing <- errors.Wrapf(err, ErrStrFormattedSettingVersionInFile, filepath)
				return
			}

			if err := writeFile(vbd.bump.FS, filepath, newContent); err != nil {
				vbd.bump.errChanPostProcessing <- errors.Wrapf(err, ErrStrFormattedWritingToFile, filepath)
				return
			}
		}
	}

	console.VersionUpdateField(oldVersionStr, vbd.versionStr, filepath, field)

	vbd.bump.mutex.Unlock()

	vbd.bump.WaitGroup.Done()
}

func (vbd *versionBumpData) versionConfirmationPrompt(oldVersionStr string, file string) (bool, error) {
	if vbd.runArgs.ConfirmationPrompt != nil {
		confirmed, err := vbd.runArgs.ConfirmationPrompt(oldVersionStr, vbd.versionStr, file)
		if err != nil {
			return false, err
		}
		return confirmed, nil
	}
	return true, nil //auto confirm when ConfirmationPrompt function is not declared
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
