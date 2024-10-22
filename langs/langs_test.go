package langs_test

import (
	"github.com/joe-at-startupmedia/version-bump/v2/langs/docker"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/golang"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/js"
	"testing"

	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/stretchr/testify/assert"
)

func TestLangs_New(t *testing.T) {
	a := assert.New(t)

	type test struct {
		ExpectedResult *langs.Language
	}

	suite := map[string]test{
		"Docker": {
			ExpectedResult: &langs.Language{
				Name:  docker.Name,
				Files: docker.Files,
				Regex: &docker.Regex,
			},
		},
		"Go": {
			ExpectedResult: &langs.Language{
				Name:  golang.Name,
				Files: golang.Files,
				Regex: &golang.Regex,
			},
		},
		"JavaScript": {
			ExpectedResult: &langs.Language{
				Name:       js.Name,
				Files:      js.Files,
				JSONFields: &js.JSONFields,
			},
		},
		"Not Supported Language": {
			ExpectedResult: nil,
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		r := langs.Supported[name]

		if name == "Not Supported Language" {
			a.Equal(test.ExpectedResult, r)
		} else {
			a.EqualValues(test.ExpectedResult, r)
		}
	}
}
