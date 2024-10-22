package langs

import (
	"github.com/joe-at-startupmedia/version-bump/v2/langs/docker"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/golang"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/js"
)

type Language struct {
	Regex      *[]string
	JSONFields *[]string
	Name       string
	Files      []string
}

var Languages = []Language{
	{
		Name:  docker.Name,
		Files: docker.Files,
		Regex: &docker.Regex,
	},
	{
		Name:  golang.Name,
		Files: golang.Files,
		Regex: &golang.Regex,
	},
	{
		Name:       js.Name,
		Files:      js.Files,
		JSONFields: &js.JSONFields,
	},
}

var Supported map[string]*Language

func init() {
	Supported = make(map[string]*Language, len(Languages))
	for li := range Languages {
		Supported[Languages[li].Name] = &Languages[li]
	}
}
