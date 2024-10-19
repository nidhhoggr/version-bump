package langs

import (
	"fmt"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
)

var dockerRegex = []string{
	fmt.Sprintf("^LABEL .*org.opencontainers.image.version['\"= ]*(?P<version>%v)['\"]?.*", version.Regex),
	fmt.Sprintf("^\\s*['\"]?org.opencontainers.image.version['\"= ]*(?P<version>%v)['\"]?.*", version.Regex),
}
