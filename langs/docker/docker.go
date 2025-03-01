package docker

import (
	"fmt"
	"github.com/nidhhoggr/version-bump/version"
)

const Name = "Docker"

var Files = []string{"Dockerfile"}

var Regex = []string{
	fmt.Sprintf("^LABEL .*org.opencontainers.image.version['\"= ]*(?P<version>%v)['\"]?.*", version.Regex),
	fmt.Sprintf("^\\s*['\"]?org.opencontainers.image.version['\"= ]*(?P<version>%v)['\"]?.*", version.Regex),
}
