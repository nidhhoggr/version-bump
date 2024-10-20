package version

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

const Regex = `[vV]?([0-9]*)\.([0-9]*)\.([0-9]*)(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))??`

type Type int

const (
	NotAVersion Type = iota
	Major
	Minor
	Patch
)

func FromString(s string) Type {
	switch s {
	case "major":
		return Major
	case "minor":
		return Minor
	case "patch":
		return Patch
	}
	return NotAVersion
}

type Version struct {
	semverPtr  *semver.Version
	preRelease *PreRelease
}

func (v *Version) SetSemverPtr(semverPtr *semver.Version) {
	v.semverPtr = semverPtr
}

func New(versionString string) (*Version, error) {
	versionString = strings.TrimLeft(versionString, "vV")
	//fmt.Printf("Get version from string %s \n", versionString)
	semverPtr, err := semver.StrictNewVersion(versionString)
	//fmt.Printf("Got version from string %s \n", semverPtr)
	return &Version{
		semverPtr,
		nil,
	}, err
}

func NewFromRegex(versionString string, regex *regexp.Regexp) (*Version, error) {
	//fmt.Printf("Get versionStr from regex: %s %s\n", versionString, regex)
	replaced := regex.ReplaceAllString(versionString, "${1}")
	if replaced == "" {
		return nil, fmt.Errorf("empty result when parsing versionStr from regex: %s %s", versionString, regex)
	}
	//fmt.Printf("Got versionStr: %s\n", replaced)
	return New(replaced)
}

func (v *Version) Increment(versionType Type, preReleaseType PreReleaseType) error {
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
		return fmt.Errorf("cannot prerelease a non-prerelease without incrementing a version type")
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
		err := v.PreRelease(preReleaseType)
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

func (v *Version) PreRelease(preReleaseType PreReleaseType) error {
	if preReleaseType == NotAPreRelease {
		return fmt.Errorf("cannot prerelease and empty type")
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
				return fmt.Errorf("cannot prerelease an alpha from an existing beta pre-release")
			} else if preReleaseType != BetaPreRelease { //only other option is rc
				err = v.SetPreReleaseString(PreReleaseString(preReleaseType))
				if err != nil {
					return err
				}
			}
		} else if strings.Contains(fmt.Sprintf("%s", firstSegment), PreReleaseString(ReleaseCandidate)) {
			if preReleaseType == AlphaPreRelease || preReleaseType == BetaPreRelease {
				return fmt.Errorf("cannot prerelease an alpha||beta from an existing beta release candidate")
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
			return fmt.Errorf("unsupported release type: -%v", preReleaseType)
		}
	}
	err := v.IncrementPreRelease()
	return err
}

func (v *Version) String() string {
	if v.semverPtr == nil {
		return ""
	}
	return v.semverPtr.String()
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

func (v *Version) GetMetaData() string {
	return v.semverPtr.Metadata()
}

func (v *Version) IncrementPreRelease() error {
	if v.IsPreRelease() {
		preRelease, err := v.GetPreRelease()
		if err != nil {
			return errors.Wrap(err, "could not get pre-release")
		}
		preRelease.Increment()
		if preRelease.Length() > 0 {
			err = v.SetPreReleaseString(preRelease.String())
		} else {
			err = errors.Wrap(err, "error incrementing pre-release")
		}
		return err

	} else {
		return fmt.Errorf("%v is not a prerelease", v)
	}
}
