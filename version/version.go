package version

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

var (
	ErrStrPreReleasingNonPreRelease    = "cannot prerelease a non-prerelease without incrementing a version type"
	ErrStrPreReleaseEmptyType          = "cannot prerelease an empty type"
	ErrStrPreReleaseAlphaFromBeta      = "cannot prerelease an alpha from an existing beta pre-release"
	ErrStrPreReleaseNonRcFromRc        = "cannot prerelease a non-rc from a release candidate"
	ErrStrParsePreReleaseTag           = "could not parse pre-release tag"
	ErrStrIncrementerGettingPreRelease = "incrementing: could not get pre-release"
	ErrStrIncrementingPreRelease       = "incrementing pre-release"

	ErrStrFormattedUnsupportedReleaseType  = "unsupported release type: %d"
	ErrStrFormattedRegexParsingResultEmpty = "empty result when parsing versionStr from regex: %s %s"
	ErrStrFormattedNotAPrerelease          = "%v is not a prerelease"
)

const Regex = `[vV]?([0-9]*)\.([0-9]*)\.([0-9]*)(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))??`

type Type int

var TypeStrings = []string{"major", "minor", "patch"}

type SemverInterface interface {
	IncMajor() semver.Version
	IncMinor() semver.Version
	IncPatch() semver.Version
	Prerelease() string
	Metadata() string
	SetPrerelease(string) (semver.Version, error)
	SetMetadata(string) (semver.Version, error)
	String() string
}

type Version struct {
	semverPtr SemverInterface
}

const (
	NotAVersion Type = iota
	Major
	Minor
	Patch
)

func FromString(s string) Type {
	switch s {
	case TypeStrings[0]:
		return Major
	case TypeStrings[1]:
		return Minor
	case TypeStrings[2]:
		return Patch
	}
	return NotAVersion
}

func (v *Version) SetSemverPtr(semverPtr SemverInterface) {
	v.semverPtr = semverPtr
}

func New(versionString string) (*Version, error) {
	versionString = strings.TrimLeft(versionString, "vV")
	console.Debug("Version.New()", fmt.Sprintf("get version from string %s \n", versionString))
	semverPtr, err := semver.StrictNewVersion(versionString)
	if err != nil {
		return nil, err
	}
	//console.Debug("Version.New()", fmt.Sprintf("got version from string %s \n", semverPtr))

	return &Version{
		semverPtr: semverPtr,
	}, nil
}

func NewFromRegex(versionString string, regex *regexp.Regexp) (*Version, error) {
	//trim surrounding whitespace
	versionString = strings.Trim(versionString, " ")
	console.Debug("Version.NewFromRegex()", fmt.Sprintf("get versionStr from regex: %s %s\n", versionString, regex))
	replaced := regex.ReplaceAllString(versionString, "${1}")
	if replaced == "" {
		return nil, fmt.Errorf(ErrStrFormattedRegexParsingResultEmpty, versionString, regex)
	}
	console.Debug("Version.NewFromRegex()", fmt.Sprintf("got versionStr: %s\n", replaced))
	return New(replaced)
}

func (v *Version) Increment(versionType Type, preReleaseType PreReleaseType, preReleaseMetadata string) error {
	var newVersion semver.Version

	isVersionBumping := versionType > NotAVersion
	isPreReleasing := preReleaseType > NotAPreRelease

	// see https://github.com/Masterminds/semver/issues/251
	if v.IsPreRelease() &&
		isVersionBumping &&
		isPreReleasing &&
		versionType == Patch {
		_ = v.SetPreReleaseString("")
	} else if !v.IsPreRelease() &&
		isPreReleasing &&
		!isVersionBumping {
		return errors.New(ErrStrPreReleasingNonPreRelease)
	}

	if isVersionBumping {
		switch versionType {
		case Major:
			newVersion = v.semverPtr.IncMajor()
		case Minor:
			newVersion = v.semverPtr.IncMinor()
		case Patch:
			newVersion = v.semverPtr.IncPatch()
		}
		v.semverPtr = &newVersion
	}

	if isPreReleasing {
		err := v.PreRelease(preReleaseType, preReleaseMetadata)
		if err != nil {
			console.Error(err)
			return err
		}
	}

	return nil
}

func (v *Version) IsPreRelease() bool {
	return len(v.semverPtr.Prerelease()) > 0
}

func (v *Version) PreRelease(preReleaseType PreReleaseType, preReleaseMetadata string) error {
	if preReleaseType == NotAPreRelease {
		return errors.New(ErrStrPreReleaseEmptyType)
	}
	if v.IsPreRelease() {
		preRelease, err := v.GetPreRelease()
		if err != nil {
			return err
		}
		firstSegment := preRelease.Segments[0]
		if strings.Contains(fmt.Sprintf("%s", firstSegment), PreReleaseString(AlphaPreRelease)) {
			if preReleaseType != AlphaPreRelease {
				err = v.SetPreReleaseString(PreReleaseString(preReleaseType))
				if err != nil {
					return err
				}
			}
		} else if strings.Contains(fmt.Sprintf("%s", firstSegment), PreReleaseString(BetaPreRelease)) {
			if preReleaseType == AlphaPreRelease {
				return errors.New(ErrStrPreReleaseAlphaFromBeta)
			} else if preReleaseType != BetaPreRelease { //only other option is rc
				err = v.SetPreReleaseString(PreReleaseString(preReleaseType))
				if err != nil {
					return err
				}
			}
		} else if strings.Contains(fmt.Sprintf("%s", firstSegment), PreReleaseString(ReleaseCandidate)) {
			if preReleaseType == AlphaPreRelease || preReleaseType == BetaPreRelease {
				return errors.New(ErrStrPreReleaseNonRcFromRc)
			}
		}
	} else {
		switch preReleaseType {
		case AlphaPreRelease:
			fallthrough
		case BetaPreRelease:
			fallthrough
		case ReleaseCandidate:
			err := v.SetPreReleaseString(PreReleaseString(preReleaseType))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf(ErrStrFormattedUnsupportedReleaseType, preReleaseType)
		}
	}
	err := v.IncrementPreRelease()
	if err != nil {
		return err
	}
	if preReleaseMetadata != "" || v.GetMetaData() != "" {
		err = v.SetPreReleaseMetadata(preReleaseMetadata)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *Version) String() string {
	if v.semverPtr == nil {
		return ""
	}
	return v.semverPtr.String()
}

func (v *Version) GetPreRelease() (*PreRelease, error) {
	preReleaseStr := v.GetPreReleaseString()
	preRelease, err := parsePreRelease(preReleaseStr)
	if err != nil {
		return nil, errors.WithMessage(err, ErrStrParsePreReleaseTag)
	}
	return preRelease, nil
}

func (v *Version) GetPreReleaseString() string {
	return v.semverPtr.Prerelease()
}

func (v *Version) SetPreReleaseString(preReleaseStr string) error {
	//fmt.Printf("setting prerelease string: %s\n", preReleaseStr)
	setPrerelease, err := v.semverPtr.SetPrerelease(preReleaseStr)
	if err != nil {
		return err
	}
	v.semverPtr = &setPrerelease
	return nil
}

func (v *Version) SetPreReleaseMetadata(metadataStr string) error {
	//fmt.Printf("setting prerelease string: %s\n", preReleaseStr)
	setPrerelease, err := v.semverPtr.SetMetadata(metadataStr)
	if err != nil {
		return err
	}
	v.semverPtr = &setPrerelease
	return nil
}

func (v *Version) GetMetaData() string {
	return v.semverPtr.Metadata()
}

func (v *Version) IncrementPreRelease() error {
	if v.IsPreRelease() {
		preRelease, err := v.GetPreRelease()
		if err != nil {
			return errors.Wrap(err, ErrStrIncrementerGettingPreRelease)
		}
		preRelease.Increment()
		if preRelease.Length() > 0 {
			err = v.SetPreReleaseString(preRelease.String())
		} else {
			err = errors.Wrap(err, ErrStrIncrementingPreRelease)
		}
		return err

	} else {
		return fmt.Errorf(ErrStrFormattedNotAPrerelease, v)
	}
}
