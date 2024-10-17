package langs

import (
	"fmt"

	"github.com/anton-yurchenko/go-changelog"
)

var golangRegex = []string{
	fmt.Sprintf("^const [vV]ersion\\s*string = \"[vV]?(?P<version>%v)\"", changelog.SemVerRegex),
	fmt.Sprintf("^const [vV]ersion := \"[vV]?(?P<version>%v)\"", changelog.SemVerRegex),
	fmt.Sprintf("^\\s*[vV]ersion\\s*string = \"[vV]?(?P<version>%v)\"", changelog.SemVerRegex),
}
