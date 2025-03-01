package golang

import (
	"fmt"
	"github.com/nidhhoggr/version-bump/version"
)

const Name = "Go"

var Files = []string{"*.go"}

var Regex = []string{
	fmt.Sprintf("^const [vV]ersion\\s*string = \"(?P<version>%v)\"", version.Regex),
	fmt.Sprintf("^const [vV]ersion := \"(?P<version>%v)\"", version.Regex),
	fmt.Sprintf("^\\s*[vV]ersion\\s*string = \"(?P<version>%v)\"", version.Regex),
}
