package obilandmark

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"github.com/DavidGamba/go-getoptions"
)

var _nCenter = 200

// LandmarkOptionSet sets the options for Obilandmark.
//
// options: a pointer to the getoptions.GetOpt struct.
// Return type: none.
func LandmarkOptionSet(options *getoptions.GetOpt) {

	options.IntVar(&_nCenter, "center", _nCenter,
		options.Alias("n"),
		options.Description("Number of landmark sequences to be selected."))
}

// OptionSet is a function that sets the options for the GetOpt struct.
//
// It takes a pointer to a GetOpt struct as its parameter and does not return anything.
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	obifind.LoadTaxonomyOptionSet(options, false, false)
	LandmarkOptionSet(options)
}

// CLINCenter returns desired number of centers as specified by user.
//
// No parameters.
// Returns an integer value.
func CLINCenter() int {
	return _nCenter
}
