package mocks

import (
	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/mock"
)

type GitConfigParserMock struct {
	Config *config.Config
	mock.Mock
}

func (gcp *GitConfigParserMock) SetConfig(config *config.Config) {
	gcp.Config = config
}

func (gcp *GitConfigParserMock) GetSectionOption(section string, option string) string {
	args := gcp.Called(section, option)
	return args.String(0)
}
