package obioptions

// Version is automatically updated by the Makefile from version.txt
// The patch number (third digit) is incremented on each push to the repository

var _Version = "Release 4.4.8"

// Version returns the version of the obitools package.
//
// No parameters.
// Returns a string representing the version of the obitools package.
func VersionString() string {
	return _Version
}
