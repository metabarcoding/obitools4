package obioptions

import (
	"fmt"
)

// TODO: The version number is extracted from git. This induces that the version
// corresponds to the last commit, and not the one when the file will be
// commited
var _Commit = "bfe3d0e"
var _Version = "Release 4.2.0"

// Version returns the version of the obitools package.
//
// No parameters.
// Returns a string representing the version of the obitools package.
func VersionString() string {
	return fmt.Sprintf("%s (%s)", _Version, _Commit)
}