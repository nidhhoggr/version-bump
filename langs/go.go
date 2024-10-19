package langs

import (
	"fmt"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
)

var golangRegex = []string{
	fmt.Sprintf("^const [vV]ersion\\s*string = \"(?P<version>%v)\"", version.Regex),
	fmt.Sprintf("^const [vV]ersion := \"(?P<version>%v)\"", version.Regex),
	fmt.Sprintf("^\\s*[vV]ersion\\s*string = \"(?P<version>%v)\"", version.Regex),
}
