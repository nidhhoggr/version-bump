package version

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

type PreReleaseType int

const (
	NotAPreRelease PreReleaseType = iota
	AlphaPreRelease
	BetaPreRelease
	ReleaseCandidate
)

var PreReleaseTypeStrings = []string{"alpha", "beta", "rc"}

func PreReleaseString(ptr PreReleaseType) string {
	if ptr == NotAPreRelease {
		return ""
	}
	return PreReleaseTypeStrings[ptr-1]
}

func FromPreReleaseTypeString(s string) PreReleaseType {
	switch s {
	case PreReleaseTypeStrings[0]:
		return AlphaPreRelease
	case PreReleaseTypeStrings[1]:
		return BetaPreRelease
	case PreReleaseTypeStrings[2]:
		return ReleaseCandidate
	}
	return NotAPreRelease
}

type PreRelease struct {
	Segments   []interface{}
	segmentLen int
}

func (v *Version) GetPreRelease() (*PreRelease, error) {
	preReleaseStr := v.GetPreReleaseString()
	preRelease, err := parsePreRelease(preReleaseStr)
	if err != nil {
		return nil, errors.WithMessage(err, "Could not parse pre-release tag")
	}
	return preRelease, nil
}

func (p *PreRelease) Length() int {
	return p.segmentLen
}

func (p *PreRelease) String() string {
	prElements := make([]string, 0, p.Length())
	for _, prElement := range p.Segments {
		prElements = append(prElements, fmt.Sprintf("%v", prElement))
	}
	return fmt.Sprintf("%s", strings.Join(prElements, "."))
}

func (p *PreRelease) Append(segment interface{}) {
	newSegments := append(make([]interface{}, 0, p.segmentLen), p.Segments...)
	p.Segments = append(newSegments, segment)
	p.segmentLen++
}

func (p *PreRelease) Increment() {
	lastIndex := p.segmentLen - 1
	lastSegment := p.Segments[lastIndex]
	sType := segmentType(lastSegment)
	switch sType {
	case "int64":
		//fmt.Printf("type: Integer: %v\n", sType)
		p.Segments[lastIndex] = lastSegment.(int64) + 1
	case "string":
		//fmt.Printf("type: String: %v\n", sType)
		p.Append(int64(0))
	}
}

func segmentType(segment interface{}) string {
	return fmt.Sprintf("%s", reflect.ValueOf(segment).Kind())
}

func parsePreRelease(str string) (*PreRelease, error) {

	if len(str) == 0 {
		return nil, nil
	}

	parts := strings.Split(str, ".")

	preReleaseTags := make([]interface{}, 0, len(parts)+1)

	for _, part := range parts {

		parsedPart, err := strconv.ParseInt(part, 10, 64)

		if err != nil {
			numErr, ok := err.(*strconv.NumError)

			if !ok || !errors.Is(numErr.Err, strconv.ErrSyntax) {
				return nil, errors.WithMessage(err, "Could not parse part '"+part+"' as int64")
			}

			preReleaseTags = append(preReleaseTags, part)
		} else {
			preReleaseTags = append(preReleaseTags, parsedPart)
		}

	}

	return &PreRelease{
		preReleaseTags,
		len(preReleaseTags),
	}, nil
}
