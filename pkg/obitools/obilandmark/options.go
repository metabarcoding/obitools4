package obilandmark

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _nCenter = 200

// ObilandmarkOptionSet sets the options for Obilandmark.
//
// options: a pointer to the getoptions.GetOpt struct.
// Return type: none.
func ObilandmarkOptionSet(options *getoptions.GetOpt) {

	options.IntVar(&_nCenter, "center", _nCenter,
		options.Alias("n"),
		options.Description("Maximum numbers of differences between two variant sequences (default: %d)."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	ObilandmarkOptionSet(options)
}

func NCenter() int {
	return _nCenter
}
