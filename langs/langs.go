package langs

import (
	"github.com/joe-at-startupmedia/version-bump/v2/langs/docker"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/generic"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/golang"
	"github.com/joe-at-startupmedia/version-bump/v2/langs/js"
)

type Settings struct {
	Regex      *[]string
	JSONFields *[]string
	Name       string
	Files      []string
}

type Config struct {
	Regex        []string
	JSONFields   []string
	Name         string
	Directories  []string
	Files        []string
	ExcludeFiles []string `toml:"exclude_files"`
	Enabled      bool
}

type ConfigDecoder struct {
	Docker     Config
	Go         Config
	JavaScript Config
	Generic    Config
}

var Languages = []Settings{
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
	{
		Name:  generic.Name,
		Regex: &generic.Regex,
	},
}

var Supported map[string]*Settings

func init() {
	Supported = make(map[string]*Settings, len(Languages))
	for li := range Languages {
		Supported[Languages[li].Name] = &Languages[li]
	}
}

func (c *Config) GetDirectories() []string {
	if len(c.Directories) == 0 {
		c.Directories = []string{"."}
	}
	return c.Directories
}
