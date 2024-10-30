package version_test

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/joe-at-startupmedia/version-bump/v2/mocks"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"regexp"
	"strings"
	"testing"
)

func TestVersion_Construction(t *testing.T) {
	a := assert.New(t)

	v, err := version.New("1.0.0")
	a.Empty(err)
	a.Equal("1.0.0", v.String())

	v, err = version.New("v1.0.1")
	a.Empty(err)
	a.Equal("1.0.1", v.String())

	v, err = version.New("v1.0")
	a.Equal(err, semver.ErrInvalidSemVer)
	a.Nil(v)

	v, err = version.New("v1.0.1-alpha")
	a.Empty(err)
	a.Equal("1.0.1-alpha", v.String())
	a.Equal("alpha", v.GetPrereleaseString())

	v, err = version.New("v1.0.1-alpha1")
	a.Empty(err)
	a.Equal("1.0.1-alpha1", v.String())
	a.Equal("alpha1", v.GetPrereleaseString())

	v, err = version.New("v1.0.1-alpha.1")
	a.Empty(err)
	a.Equal("1.0.1-alpha.1", v.String())
	a.Equal("alpha.1", v.GetPrereleaseString())

	v, err = version.New("v1.0.1-alpha.beta")
	a.Empty(err)
	a.Equal("1.0.1-alpha.beta", v.String())
	a.Equal("alpha.beta", v.GetPrereleaseString())

	v, err = version.New("v1.0.1-alpha.beta.1")
	a.Empty(err)
	a.Equal("1.0.1-alpha.beta.1", v.String())
	a.Equal("alpha.beta.1", v.GetPrereleaseString())
}

func TestVersion_IncrementPrerelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.1-alpha")
	a.Empty(err)

	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("alpha.0", v.GetPrereleaseString())
	a.Equal("1.0.1-alpha.0", v.String())

	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("alpha.1", v.GetPrereleaseString())
	a.Equal("1.0.1-alpha.1", v.String())

	v, err = version.New("v1.0.1-alpha.beta")
	a.Empty(err)
	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("alpha.beta.0", v.GetPrereleaseString())
	a.Equal("1.0.1-alpha.beta.0", v.String())

	v, err = version.New("v1.0.1-rc")
	a.Empty(err)
	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("rc.0", v.GetPrereleaseString())
	a.Equal("1.0.1-rc.0", v.String())

	v, err = version.New("v1.0.1-rc0")
	a.Empty(err)
	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("rc0.0", v.GetPrereleaseString())
	a.Equal("1.0.1-rc0.0", v.String())
}

func TestVersion_IncrementPrereleaseWithMetadata(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.1-beta+metadata")
	a.Empty(err)
	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("beta.0", v.GetPrereleaseString())
	a.Equal("1.0.1-beta.0+metadata", v.String())
	a.Equal("metadata", v.GetMetaData())

	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("beta.1", v.GetPrereleaseString())
	a.Equal("1.0.1-beta.1+metadata", v.String())

	v, err = version.New("v1.0.1-rc.22+release-it-already")
	a.Empty(err)
	err = v.IncrementPrerelease()
	a.Empty(err)
	a.Equal("rc.23", v.GetPrereleaseString())
	a.Equal("1.0.1-rc.23+release-it-already", v.String())
	a.Equal("release-it-already", v.GetMetaData())
}

func TestVersion_IncrementPrereleaseIfNotAPrerelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.1")
	a.Empty(err)

	err = v.IncrementPrerelease()
	a.ErrorContains(err, fmt.Sprintf(version.ErrStrFormattedNotAPrerelease, "1.0.1"))
}

func TestVersion_PrereleaseANonPrerelease(t *testing.T) {
	a := assert.New(t)

	v, err := version.New("1.0.0")
	a.Empty(err)
	a.Equal("1.0.0", v.String())
	err = v.Increment(version.Major, version.NotAPrerelease, "")
	a.Empty(err)
	a.Equal("2.0.0", v.String())
	a.Empty(err)

	v, err = version.New("1.0.0")
	a.Empty(err)
	err = v.Increment(version.Major, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("2.0.0-alpha.0", v.String())
	a.Empty(err)

	v, err = version.New("1.0.0")
	a.Empty(err)
	err = v.Increment(version.Minor, version.BetaPrerelease, "")
	a.Empty(err)
	a.Equal("1.1.0-beta.0", v.String())
	a.Empty(err)

	v, err = version.New("1.0.0+metadata")
	a.Empty(err)
	err = v.Increment(version.Patch, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.1-rc.0", v.String())
	a.Empty(err)
}

func TestVersion_PrereleaseWithUnsupportedType(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("1.0.0")
	a.Empty(err)
	a.Equal(false, v.IsPrerelease())
	a.Equal("1.0.0", v.String())
	err = v.Increment(version.Major, 43, "")
	a.ErrorContains(err, fmt.Sprintf(version.ErrStrFormattedUnsupportedReleaseType, 43))
}

func TestVersion_PrereleaseAPrereleaseWithMajorResetsPrCounter(t *testing.T) {
	a := assert.New(t)

	v, err := version.New("1.0.0-alpha+junk-md")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Major, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("2.0.0-alpha.0", v.String())
}

func TestVersion_PrereleaseAPrereleaseWithMinorResetsPrCounter(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("1.0.0-alpha.1+junk-md")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Minor, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("1.1.0-alpha.0", v.String())
}

func TestVersion_PrereleaseAPrereleaseWithPatchResetsPrCounter(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())
}

func TestVersion_PrereleaseAPrereleaseWithReleasePromotionsFromAlpha(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.BetaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-beta.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.BetaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0-beta.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.NotAPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0-alpha.1", v.String())

	v, err = version.New("v1.0.0-rc.0")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.1", v.String())

	v, err = version.New("v1.0.0+junk")
	a.Empty(err)
	a.Equal(false, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "discard")
	a.ErrorContains(err, version.ErrStrPreReleasingNonPrerelease)
	a.Equal("1.0.0+junk", v.String())
}

func TestVersion_PrereleaseAPrereleaseWithReleasePromotionsFromBeta(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.BetaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-beta.0", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.AlphaPrerelease, "")
	a.ErrorContains(err, version.ErrStrPrereleaseAlphaFromBeta)
	a.Equal("1.0.0-beta.1", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.BetaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0-beta.2", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.0", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.NotAPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0-beta.1", v.String())
}

func TestVersion_PrereleaseAPrereleaseWithReleasePromotionsFromReleaseCandidate(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.AlphaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.BetaPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.1-beta.0", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.AlphaPrerelease, "")
	a.ErrorContains(err, version.ErrStrPrereleaseNonRcFromRc)
	a.Equal("1.0.0-rc.1", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.BetaPrerelease, "")
	a.ErrorContains(err, version.ErrStrPrereleaseNonRcFromRc)
	a.Equal("1.0.0-rc.1", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.2", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.NotAPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.1", v.String())
}

func TestVersion_PromotePrerelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.NotAPrerelease, "ignore-me")
	a.Empty(err)
	a.Equal("1.0.0", v.String())

	v, err = version.New("v1.0.0-beta.2")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.NotAPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0", v.String())

	v, err = version.New("v1.0.0-rc.3")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.NotAPrerelease, "")
	a.Empty(err)
	a.Equal("1.0.0", v.String())
}

func TestVersion_PrereleaseWithNotAVersion(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.NotAVersion, version.AlphaPrerelease, "stay-alpha")
	a.Equal("1.0.0-alpha.2+stay-alpha", v.String())

	//tests the difference
	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPrerelease())
	err = v.Increment(version.Patch, version.AlphaPrerelease, "")
	a.Equal("1.0.1-alpha.0", v.String())
}

func TestVersion_NewWithBadRegex(t *testing.T) {
	a := assert.New(t)
	compile, err := regexp.Compile(version.Regex)
	a.Empty(err)
	_, err = version.NewFromRegex("", compile)
	a.ErrorContains(err, fmt.Sprintf(version.ErrStrFormattedRegexParsingResultEmpty, "", version.Regex))
}

func TestVersion_SetPrereleaseWithEmptyVersion(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	pr, err := v.GetPrerelease()
	a.Empty(err)
	a.Empty(pr)
}

func TestVersion_SetPrereleaseWithBadVersion(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	err := v.SetPrereleaseString("!/+%")
	a.Equal(err, semver.ErrInvalidPrerelease)
}

func TestVersion_SetPrereleaseWithBadMetadata(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	err := v.SetPrereleaseMetadata("!/+%")
	a.Equal(err, semver.ErrInvalidMetadata)
}

func TestVersion_PrereleaseEmptyType(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	err := v.Prerelease(version.NotAPrerelease, "")
	a.ErrorContains(err, version.ErrStrPrereleaseEmptyType)
}

func TestVersion_PrereleaseErrorGettingPrerelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	err = v.Prerelease(version.AlphaPrerelease, "-%43")
	a.Equal(err, semver.ErrInvalidMetadata)
}

func TestVersion_PrereleaseErrorGettingPrereleaseFromMajor(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0")
	err = v.Prerelease(version.AlphaPrerelease, "-%43")
	a.Equal(err, semver.ErrInvalidMetadata)
}

func TestBump_StringToVersionType(t *testing.T) {
	a := assert.New(t)
	a.Equal(version.FromString("major"), version.Major)
	a.Equal(version.FromString("minor"), version.Minor)
	a.Equal(version.FromString("patch"), version.Patch)
	a.Equal(version.FromString("nonexistent"), version.NotAVersion)
}

func TestBump_PrereleaseString(t *testing.T) {
	a := assert.New(t)
	a.Equal(version.PrereleaseString(version.AlphaPrerelease), "alpha")
	a.Equal(version.PrereleaseString(version.BetaPrerelease), "beta")
	a.Equal(version.PrereleaseString(version.ReleaseCandidate), "rc")
	a.Equal(version.PrereleaseString(version.NotAPrerelease), "")
}

func TestBump_FromPrereleaseString(t *testing.T) {
	a := assert.New(t)
	a.Equal(version.FromPrereleaseTypeString("alpha"), version.AlphaPrerelease)
	a.Equal(version.FromPrereleaseTypeString("beta"), version.BetaPrerelease)
	a.Equal(version.FromPrereleaseTypeString("rc"), version.ReleaseCandidate)
	a.Equal(version.FromPrereleaseTypeString(""), version.NotAPrerelease)
}

func TestVersion_PrereleaseErrorGettingPrereleaseTag(t *testing.T) {
	a := assert.New(t)
	versionStr := "v1.0.0-alpha.1"
	v, _ := version.New(versionStr)
	sm := new(mocks.Semver)
	versionString := strings.TrimLeft(versionStr, "vV")
	semverPtr, err := semver.StrictNewVersion(versionString)
	sm.On("SetPrerelease", mock.Anything).Return(*semverPtr, err)
	sm.On("Prerelease").Return("-%43")
	v.SetSemverPtr(sm)
	err = v.Prerelease(version.AlphaPrerelease, "")
	a.ErrorContains(err, fmt.Sprintf("%s: %s", version.ErrStrParsePrereleaseTag, fmt.Sprintf(version.ErrStrFormattedPrereleaseContainsInvalidValue, "-%43")))
}

func TestVersion_PrereleaseErrorGettingPrereleaseTagTwo(t *testing.T) {
	a := assert.New(t)
	versionStr := "v1.0.0-alpha.1"
	v, _ := version.New(versionStr)
	sm := new(mocks.Semver)
	versionString := strings.TrimLeft(versionStr, "vV")
	semverPtr, err := semver.StrictNewVersion(versionString)
	sm.On("SetPrerelease", mock.Anything).Return(*semverPtr, err)
	sm.On("Prerelease").Return("-%43")
	v.SetSemverPtr(sm)
	err = v.IncrementPrerelease()
	a.ErrorContains(err, fmt.Sprintf("%s: %s", version.ErrStrIncrementerGettingPrerelease, fmt.Sprintf("%s: %s", version.ErrStrParsePrereleaseTag, fmt.Sprintf(version.ErrStrFormattedPrereleaseContainsInvalidValue, "-%43"))))
}

func TestVersion_EmptyPtrReturnsEmptyString(t *testing.T) {
	a := assert.New(t)
	versionStr := "v1.0.0-alpha.1"
	v, _ := version.New(versionStr)
	v.SetSemverPtr(nil)
	a.Equal(v.String(), "")
}
