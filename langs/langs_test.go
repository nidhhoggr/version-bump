package langs_test

import (
	"github.com/joe-at-startupmedia/version-bump/v2/langs/docker"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/generic"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/golang"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/js"
	"testing"

	"github.com/joe-at-startupmedia/version-bump/v2/langs"
	"github.com/stretchr/testify/assert"
)

func TestLangs_New(t *testing.T) {
	a := assert.New(t)

	type test struct {
		ExpectedResult *langs.Settings
	}

	suite := map[string]test{
		"Docker": {
			ExpectedResult: &langs.Settings{
				Name:  docker.Name,
				Files: docker.Files,
				Regex: &docker.Regex,
			},
		},
		"Go": {
			ExpectedResult: &langs.Settings{
				Name:  golang.Name,
				Files: golang.Files,
				Regex: &golang.Regex,
			},
		},
		"JavaScript": {
			ExpectedResult: &langs.Settings{
				Name:       js.Name,
				Files:      js.Files,
				JSONFields: &js.JSONFields,
			},
		},
		"Generic": {
			ExpectedResult: &langs.Settings{
				Name:  generic.Name,
				Regex: &generic.Regex,
			},
		},
		"Not Supported Settings": {
			ExpectedResult: nil,
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		r := langs.Supported[name]

		if name == "Not Supported Settings" {
			a.Equal(test.ExpectedResult, r)
		} else {
			a.EqualValues(test.ExpectedResult, r)
		}
	}
}
