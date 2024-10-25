package version_test

import (
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
	a.ErrorContains(err, "Invalid Semantic Version")
	a.Nil(v)

	v, err = version.New("v1.0.1-alpha")
	a.Empty(err)
	a.Equal("1.0.1-alpha", v.String())
	a.Equal("alpha", v.GetPreReleaseString())

	v, err = version.New("v1.0.1-alpha1")
	a.Empty(err)
	a.Equal("1.0.1-alpha1", v.String())
	a.Equal("alpha1", v.GetPreReleaseString())

	v, err = version.New("v1.0.1-alpha.1")
	a.Empty(err)
	a.Equal("1.0.1-alpha.1", v.String())
	a.Equal("alpha.1", v.GetPreReleaseString())

	v, err = version.New("v1.0.1-alpha.beta")
	a.Empty(err)
	a.Equal("1.0.1-alpha.beta", v.String())
	a.Equal("alpha.beta", v.GetPreReleaseString())

	v, err = version.New("v1.0.1-alpha.beta.1")
	a.Empty(err)
	a.Equal("1.0.1-alpha.beta.1", v.String())
	a.Equal("alpha.beta.1", v.GetPreReleaseString())
}

func TestVersion_IncrementPreRelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.1-alpha")
	a.Empty(err)

	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("alpha.0", v.GetPreReleaseString())
	a.Equal("1.0.1-alpha.0", v.String())

	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("alpha.1", v.GetPreReleaseString())
	a.Equal("1.0.1-alpha.1", v.String())

	v, err = version.New("v1.0.1-alpha.beta")
	a.Empty(err)
	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("alpha.beta.0", v.GetPreReleaseString())
	a.Equal("1.0.1-alpha.beta.0", v.String())

	v, err = version.New("v1.0.1-rc")
	a.Empty(err)
	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("rc.0", v.GetPreReleaseString())
	a.Equal("1.0.1-rc.0", v.String())

	v, err = version.New("v1.0.1-rc0")
	a.Empty(err)
	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("rc0.0", v.GetPreReleaseString())
	a.Equal("1.0.1-rc0.0", v.String())
}

func TestVersion_IncrementPreReleaseWithMetadata(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.1-beta+metadata")
	a.Empty(err)
	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("beta.0", v.GetPreReleaseString())
	a.Equal("1.0.1-beta.0+metadata", v.String())
	a.Equal("metadata", v.GetMetaData())

	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("beta.1", v.GetPreReleaseString())
	a.Equal("1.0.1-beta.1+metadata", v.String())

	v, err = version.New("v1.0.1-rc.22+release-it-already")
	a.Empty(err)
	err = v.IncrementPreRelease()
	a.Empty(err)
	a.Equal("rc.23", v.GetPreReleaseString())
	a.Equal("1.0.1-rc.23+release-it-already", v.String())
	a.Equal("release-it-already", v.GetMetaData())
}

func TestVersion_IncrementPreReleaseIfNotAPrerelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.1")
	a.Empty(err)

	err = v.IncrementPreRelease()
	a.ErrorContains(err, "1.0.1 is not a prerelease")
}

func TestVersion_PreReleaseANonPreRelease(t *testing.T) {
	a := assert.New(t)

	v, err := version.New("1.0.0")
	a.Empty(err)
	a.Equal("1.0.0", v.String())
	err = v.Increment(version.Major, version.NotAPreRelease, "")
	a.Empty(err)
	a.Equal("2.0.0", v.String())
	a.Empty(err)

	v, err = version.New("1.0.0")
	a.Empty(err)
	err = v.Increment(version.Major, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("2.0.0-alpha.0", v.String())
	a.Empty(err)

	v, err = version.New("1.0.0")
	a.Empty(err)
	err = v.Increment(version.Minor, version.BetaPreRelease, "")
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

func TestVersion_PreReleaseWithUnsupportedType(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("1.0.0")
	a.Empty(err)
	a.Equal(false, v.IsPreRelease())
	a.Equal("1.0.0", v.String())
	err = v.Increment(version.Major, 43, "")
	a.ErrorContains(err, "unsupported release type:")
}

func TestVersion_PreReleaseAPreReleaseWithMajorResetsPrCounter(t *testing.T) {
	a := assert.New(t)

	v, err := version.New("1.0.0-alpha+junk-md")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Major, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("2.0.0-alpha.0", v.String())
}

func TestVersion_PreReleaseAPreReleaseWithMinorResetsPrCounter(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("1.0.0-alpha.1+junk-md")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Minor, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("1.1.0-alpha.0", v.String())
}

func TestVersion_PreReleaseAPreReleaseWithPatchResetsPrCounter(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())
}

func TestVersion_PreReleaseAPreReleaseWithReleasePromotionsFromAlpha(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.BetaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-beta.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.BetaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0-beta.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.0", v.String())

	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.NotAPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0-alpha.1", v.String())

	v, err = version.New("v1.0.0-rc.0")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.1", v.String())

	v, err = version.New("v1.0.0+junk")
	a.Empty(err)
	a.Equal(false, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "discard")
	a.ErrorContains(err, "cannot prerelease a non-prerelease without incrementing a version type")
	a.Equal("1.0.0+junk", v.String())
}

func TestVersion_PreReleaseAPreReleaseWithReleasePromotionsFromBeta(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.BetaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-beta.0", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.AlphaPreRelease, "")
	a.ErrorContains(err, "cannot prerelease an alpha from an existing beta pre-release")
	a.Equal("1.0.0-beta.1", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.BetaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0-beta.2", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.0", v.String())

	v, err = version.New("v1.0.0-beta.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.NotAPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0-beta.1", v.String())
}

func TestVersion_PreReleaseAPreReleaseWithReleasePromotionsFromReleaseCandidate(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.AlphaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-alpha.0", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.BetaPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.1-beta.0", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.AlphaPreRelease, "")
	a.ErrorContains(err, "cannot prerelease an alpha||beta from an existing beta release candidate")
	a.Equal("1.0.0-rc.1", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.BetaPreRelease, "")
	a.ErrorContains(err, "cannot prerelease an alpha||beta from an existing beta release candidate")
	a.Equal("1.0.0-rc.1", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.ReleaseCandidate, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.2", v.String())

	v, err = version.New("v1.0.0-rc.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.NotAPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0-rc.1", v.String())
}

func TestVersion_PromotePreRelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.NotAPreRelease, "ignore-me")
	a.Empty(err)
	a.Equal("1.0.0", v.String())

	v, err = version.New("v1.0.0-beta.2")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.NotAPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0", v.String())

	v, err = version.New("v1.0.0-rc.3")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.NotAPreRelease, "")
	a.Empty(err)
	a.Equal("1.0.0", v.String())
}

func TestVersion_PreReleaseWithNotAVersion(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.NotAVersion, version.AlphaPreRelease, "stay-alpha")
	a.Equal("1.0.0-alpha.2+stay-alpha", v.String())

	//tests the difference
	v, err = version.New("v1.0.0-alpha.1")
	a.Empty(err)
	a.Equal(true, v.IsPreRelease())
	err = v.Increment(version.Patch, version.AlphaPreRelease, "")
	a.Equal("1.0.1-alpha.0", v.String())
}

func TestVersion_NewWithBadRegex(t *testing.T) {
	a := assert.New(t)
	compile, err := regexp.Compile(version.Regex)
	a.Empty(err)
	_, err = version.NewFromRegex("", compile)
	a.ErrorContains(err, "empty result when parsing versionStr from regex")
}

func TestVersion_SetPrereleaseWithEmptyVersion(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	pr, err := v.GetPreRelease()
	a.Empty(err)
	a.Empty(pr)
}

func TestVersion_SetPrereleaseWithBadVersion(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	err := v.SetPreReleaseString("!/+%")
	a.ErrorContains(err, "Invalid Prerelease string")
}

func TestVersion_SetPrereleaseWithBadMetadata(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	err := v.SetPreReleaseMetadata("!/+%")
	a.ErrorContains(err, "Invalid Metadata string")
}

func TestVersion_PreReleaseEmptyType(t *testing.T) {
	a := assert.New(t)
	v := &version.Version{}
	v.SetSemverPtr(&semver.Version{})
	err := v.PreRelease(version.NotAPreRelease, "")
	a.ErrorContains(err, "cannot prerelease and empty type")
}

func TestVersion_PreReleaseErrorGettingPreRelease(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0-alpha.1")
	err = v.PreRelease(version.AlphaPreRelease, "-%43")
	a.ErrorContains(err, "Invalid Metadata string")
}

func TestVersion_PreReleaseErrorGettingPreReleaseFromMajor(t *testing.T) {
	a := assert.New(t)
	v, err := version.New("v1.0.0")
	err = v.PreRelease(version.AlphaPreRelease, "-%43")
	a.ErrorContains(err, "Invalid Metadata string")
}

func TestBump_StringToVersionType(t *testing.T) {
	a := assert.New(t)
	a.Equal(version.FromString("major"), version.Major)
	a.Equal(version.FromString("minor"), version.Minor)
	a.Equal(version.FromString("patch"), version.Patch)
	a.Equal(version.FromString("nonexistent"), version.NotAVersion)
}

func TestBump_PreReleaseString(t *testing.T) {
	a := assert.New(t)
	a.Equal(version.PreReleaseString(version.AlphaPreRelease), "alpha")
	a.Equal(version.PreReleaseString(version.BetaPreRelease), "beta")
	a.Equal(version.PreReleaseString(version.ReleaseCandidate), "rc")
	a.Equal(version.PreReleaseString(version.NotAPreRelease), "")
}

func TestBump_FromPreReleaseString(t *testing.T) {
	a := assert.New(t)
	a.Equal(version.FromPreReleaseTypeString("alpha"), version.AlphaPreRelease)
	a.Equal(version.FromPreReleaseTypeString("beta"), version.BetaPreRelease)
	a.Equal(version.FromPreReleaseTypeString("rc"), version.ReleaseCandidate)
	a.Equal(version.FromPreReleaseTypeString(""), version.NotAPreRelease)
}

func TestVersion_PreReleaseErrorGettingPreReleaseTag(t *testing.T) {
	a := assert.New(t)
	versionStr := "v1.0.0-alpha.1"
	v, _ := version.New(versionStr)
	sm := new(mocks.Semver)
	versionString := strings.TrimLeft(versionStr, "vV")
	semverPtr, err := semver.StrictNewVersion(versionString)
	sm.On("SetPrerelease", mock.Anything).Return(*semverPtr, err)
	sm.On("Prerelease").Return("-%43")
	v.SetSemverPtr(sm)
	err = v.PreRelease(version.AlphaPreRelease, "")
	a.ErrorContains(err, "Could not parse pre-release tag: prerelease contains invalid value: -%43")
}

func TestVersion_PreReleaseErrorGettingPreReleaseTagTwo(t *testing.T) {
	a := assert.New(t)
	versionStr := "v1.0.0-alpha.1"
	v, _ := version.New(versionStr)
	sm := new(mocks.Semver)
	versionString := strings.TrimLeft(versionStr, "vV")
	semverPtr, err := semver.StrictNewVersion(versionString)
	sm.On("SetPrerelease", mock.Anything).Return(*semverPtr, err)
	sm.On("Prerelease").Return("-%43")
	v.SetSemverPtr(sm)
	err = v.IncrementPreRelease()
	a.ErrorContains(err, "error incrementing: could not get pre-release: Could not parse pre-release tag: prerelease contains invalid value: -%43")
}

func TestVersion_EmptyPtrReturnsEmptyString(t *testing.T) {
	a := assert.New(t)
	versionStr := "v1.0.0-alpha.1"
	v, _ := version.New(versionStr)
	v.SetSemverPtr(nil)
	a.Equal(v.String(), "")
}

//func TestVersion_PreReleaseErrorGettingPreReleaseTagThree(t *testing.T) {
//	a := assert.New(t)
//	versionStr := "v1.0.0"
//	v, _ := version.New(versionStr)
//	sm := new(mocks.Semver)
//	versionString := strings.TrimLeft(versionStr, "vV")
//	semverPtr, err := semver.StrictNewVersion(versionString + "-alpha.1")
//	sm.On("SetPrerelease", mock.Anything).Return(*semverPtr, err)
//	sm.On("Prerelease").Return("").Once()
//	sm.On("Prerelease").Return("-%43").Once()
//	v.SetSemverPtr(sm)
//	err = v.PreRelease(version.AlphaPreRelease, "")
//	a.Nil(err)
//}
