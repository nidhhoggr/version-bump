package version

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"regexp"
	"strings"
)

const Regex = `[vV]?([0-9]*)\.([0-9]*)\.([0-9]*)(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))??`

type Version semver.Version

func New(versionString string) (*Version, error) {
	versionString = strings.TrimLeft(versionString, "vV")
	//fmt.Printf("Get version from string %s \n", versionString)
	version, err := semver.StrictNewVersion(versionString)
	//fmt.Printf("Got version from string %s \n", version)
	return (*Version)(version), err
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

func (v *Version) Increment(action int) *Version {
	var newVersion semver.Version

	semverPtr := (*semver.Version)(v)

	switch action {
	case 1:
		newVersion = semverPtr.IncMajor()
	case 2:
		newVersion = semverPtr.IncMinor()
	case 3:
		newVersion = semverPtr.IncPatch()
	}

	return (*Version)(&newVersion)
}

func (v *Version) String() string {
	semverPtr := (*semver.Version)(v)
	return semverPtr.String()
}
