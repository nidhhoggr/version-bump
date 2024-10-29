package generic

import (
	"fmt"
	"github.com/joe-at-startupmedia/version-bump/v2/version"
)

const Name = "Generic"

var Regex = []string{
	fmt.Sprintf("^.*?(?P<version>%v)", version.Regex),
}
