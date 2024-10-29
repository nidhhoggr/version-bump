package version

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrStrFormattedPrereleaseContainsInvalidValue = "Prerelease contains invalid value: %s"
	ErrStrFormattedParsingPrereleasePart          = "Could not parse part '%-v' as int64"
)

type Prerelease struct {
	Segments   []interface{}
	segmentLen int
}

type PrereleaseType int

const (
	NotAPrerelease PrereleaseType = iota
	AlphaPrerelease
	BetaPrerelease
	ReleaseCandidate
)

var PrereleaseTypeStrings = []string{"alpha", "beta", "rc"}

func PrereleaseString(ptr PrereleaseType) string {
	if ptr == NotAPrerelease {
		return ""
	}
	return PrereleaseTypeStrings[ptr-1]
}

func FromPrereleaseTypeString(s string) PrereleaseType {
	switch s {
	case PrereleaseTypeStrings[0]:
		return AlphaPrerelease
	case PrereleaseTypeStrings[1]:
		return BetaPrerelease
	case PrereleaseTypeStrings[2]:
		return ReleaseCandidate
	}
	return NotAPrerelease
}

func (p *Prerelease) Length() int {
	return p.segmentLen
}

func (p *Prerelease) String() string {
	prElements := make([]string, 0, p.Length())
	for _, prElement := range p.Segments {
		prElements = append(prElements, fmt.Sprintf("%v", prElement))
	}
	return fmt.Sprintf("%s", strings.Join(prElements, "."))
}

func (p *Prerelease) Append(segment interface{}) {
	newSegments := append(make([]interface{}, 0, p.segmentLen), p.Segments...)
	p.Segments = append(newSegments, segment)
	p.segmentLen++
}

func (p *Prerelease) Increment() {
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

const (
	num     string = "0123456789"
	allowed string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-" + num
)

func containsOnly(s string, comp string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(comp, r)
	}) == -1
}

func parsePrerelease(str string) (*Prerelease, error) {

	if len(str) == 0 {
		return nil, nil
	}

	parts := strings.Split(str, ".")

	PrereleaseTags := make([]interface{}, 0, len(parts)+1)

	for _, part := range parts {

		if !containsOnly(part, allowed) {
			return nil, fmt.Errorf(ErrStrFormattedPrereleaseContainsInvalidValue, part)
		}

		parsedPart, err := strconv.ParseInt(part, 10, 64)

		if err != nil {
			numErr, ok := err.(*strconv.NumError)

			if !ok || !errors.Is(numErr.Err, strconv.ErrSyntax) {
				return nil, errors.WithMessage(err, fmt.Sprintf(ErrStrFormattedParsingPrereleasePart, part))
			}

			PrereleaseTags = append(PrereleaseTags, part)
		} else {
			PrereleaseTags = append(PrereleaseTags, parsedPart)
		}

	}

	return &Prerelease{
		PrereleaseTags,
		len(PrereleaseTags),
	}, nil
}
