package git

import "github.com/go-git/go-git/v5/config"

type ConfigParserInterface interface {
	GetSectionOption(string, string) string
	SetConfig(*config.Config)
}

type ConfigParser struct {
	Config *config.Config
}

func (cp *ConfigParser) SetConfig(config *config.Config) {
	cp.Config = config
}

func (cp *ConfigParser) GetSectionOption(section string, option string) string {
	gcSection := cp.Config.Raw.Section(section)
	return gcSection.Options.Get(option)
}
