package version

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/nidhhoggr/version-bump/console"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

var (
	ErrStrPreReleasingNonPrerelease    = "cannot Prerelease a non-Prerelease without incrementing a version type"
	ErrStrPrereleaseEmptyType          = "cannot Prerelease an empty type"
	ErrStrPrereleaseAlphaFromBeta      = "cannot Prerelease an alpha from an existing beta prerelease"
	ErrStrPrereleaseNonRcFromRc        = "cannot Prerelease a non-rc from a release candidate"
	ErrStrParsePrereleaseTag           = "could not parse prerelease tag"
	ErrStrIncrementerGettingPrerelease = "incrementing: could not get prerelease"
	ErrStrIncrementingPrerelease       = "incrementing prerelease"

	ErrStrFormattedUnsupportedReleaseType  = "unsupported release type (%d)"
	ErrStrFormattedRegexParsingResultEmpty = "empty result when parsing versionStr(%s)from regex(%s)"
	ErrStrFormattedNotAPrerelease          = "%v is not a Prerelease"
)

const Regex = `[vV]?([0-9]*)\.([0-9]*)\.([0-9]*)(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

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

func (v *Version) Increment(versionType Type, PrereleaseType PrereleaseType, PrereleaseMetadata string) error {
	var newVersion semver.Version

	isVersionBumping := versionType > NotAVersion
	isPreReleasing := PrereleaseType > NotAPrerelease

	// see https://github.com/Masterminds/semver/issues/251
	if v.IsPrerelease() &&
		isVersionBumping &&
		isPreReleasing &&
		versionType == Patch {
		_ = v.SetPrereleaseString("")
	} else if !v.IsPrerelease() &&
		isPreReleasing &&
		!isVersionBumping {
		return errors.New(ErrStrPreReleasingNonPrerelease)
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
		err := v.Prerelease(PrereleaseType, PrereleaseMetadata)
		if err != nil {
			console.Error(err)
			return err
		}
	}

	return nil
}

func (v *Version) IsPrerelease() bool {
	return len(v.semverPtr.Prerelease()) > 0
}

func (v *Version) Prerelease(PrereleaseType PrereleaseType, PrereleaseMetadata string) error {
	if PrereleaseType == NotAPrerelease {
		return errors.New(ErrStrPrereleaseEmptyType)
	}
	if v.IsPrerelease() {
		Prerelease, err := v.GetPrerelease()
		if err != nil {
			return err
		}
		firstSegment := Prerelease.Segments[0]
		if strings.Contains(fmt.Sprintf("%s", firstSegment), PrereleaseString(AlphaPrerelease)) {
			if PrereleaseType != AlphaPrerelease {
				err = v.SetPrereleaseString(PrereleaseString(PrereleaseType))
				if err != nil {
					return err
				}
			}
		} else if strings.Contains(fmt.Sprintf("%s", firstSegment), PrereleaseString(BetaPrerelease)) {
			if PrereleaseType == AlphaPrerelease {
				return errors.New(ErrStrPrereleaseAlphaFromBeta)
			} else if PrereleaseType != BetaPrerelease { //only other option is rc
				err = v.SetPrereleaseString(PrereleaseString(PrereleaseType))
				if err != nil {
					return err
				}
			}
		} else if strings.Contains(fmt.Sprintf("%s", firstSegment), PrereleaseString(ReleaseCandidate)) {
			if PrereleaseType == AlphaPrerelease || PrereleaseType == BetaPrerelease {
				return errors.New(ErrStrPrereleaseNonRcFromRc)
			}
		}
	} else {
		switch PrereleaseType {
		case AlphaPrerelease:
			fallthrough
		case BetaPrerelease:
			fallthrough
		case ReleaseCandidate:
			err := v.SetPrereleaseString(PrereleaseString(PrereleaseType))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf(ErrStrFormattedUnsupportedReleaseType, PrereleaseType)
		}
	}
	err := v.IncrementPrerelease()
	if err != nil {
		return err
	}
	if PrereleaseMetadata != "" || v.GetMetaData() != "" {
		err = v.SetPrereleaseMetadata(PrereleaseMetadata)
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

func (v *Version) GetPrerelease() (*Prerelease, error) {
	PrereleaseStr := v.GetPrereleaseString()
	Prerelease, err := parsePrerelease(PrereleaseStr)
	if err != nil {
		return nil, errors.WithMessage(err, ErrStrParsePrereleaseTag)
	}
	return Prerelease, nil
}

func (v *Version) GetPrereleaseString() string {
	return v.semverPtr.Prerelease()
}

func (v *Version) SetPrereleaseString(PrereleaseStr string) error {
	//fmt.Printf("setting Prerelease string: %s\n", PrereleaseStr)
	setPrerelease, err := v.semverPtr.SetPrerelease(PrereleaseStr)
	if err != nil {
		return err
	}
	v.semverPtr = &setPrerelease
	return nil
}

func (v *Version) SetPrereleaseMetadata(metadataStr string) error {
	//fmt.Printf("setting Prerelease string: %s\n", PrereleaseStr)
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

func (v *Version) IncrementPrerelease() error {
	if v.IsPrerelease() {
		Prerelease, err := v.GetPrerelease()
		if err != nil {
			return errors.Wrap(err, ErrStrIncrementerGettingPrerelease)
		}
		Prerelease.Increment()
		if Prerelease.Length() > 0 {
			err = v.SetPrereleaseString(Prerelease.String())
		} else {
			err = errors.Wrap(err, ErrStrIncrementingPrerelease)
		}
		return err

	} else {
		return fmt.Errorf(ErrStrFormattedNotAPrerelease, v)
	}
}
