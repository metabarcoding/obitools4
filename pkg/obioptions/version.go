package obioptions

import (
	"fmt"
)

var _Commit = ""
var _Version = "Release 4.2.0"

// Version returns the version of the obitools package.
//
// No parameters.
// Returns a string representing the version of the obitools package.
func VersionString() string {
	return fmt.Sprintf("%s (%s)", _Version, _Commit)
}
