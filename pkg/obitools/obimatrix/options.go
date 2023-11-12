// obicount function utility package.
//
// The obitols/obicount package contains every
// functions specificaly required by the obicount utility.
package obimatrix

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var __threeColumns__ = false
var __mapAttribute__ = "merged_sample"
var __valueName__ = "count"
var __sampleName__ = "sample"
var __NAValue__ = "0"

func MatrixOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__threeColumns__, "three-columns", false,
		options.Description("Printouts the matrix in tree column format."))

	options.StringVar(&__mapAttribute__, "map", __mapAttribute__,
		options.Description("Which attribute is usd to produce th matrix."))

	options.StringVar(&__valueName__, "value-name", __valueName__,
		options.Description("Name of the coulumn containing the values in the three column format."))

	options.StringVar(&__sampleName__, "sample-name", __sampleName__,
		options.Description("Name of the coulumn containing the sample names in the three column format."))

	options.StringVar(&__NAValue__, "na-value", __NAValue__,
		options.Description("Value used when the map attribute is not defined for a sequence."))
}

func OptionSet(options *getoptions.GetOpt) {
	MatrixOptionSet(options)
	obiconvert.InputOptionSet(options)
}

func CLIOutFormat() string {
	if __threeColumns__ {
		return "three-columns"
	}

	return "matrix"
}

func CLIValueName() string {
	return __valueName__
}

func CLISampleName() string {
	return __sampleName__
}

func CLINaValue() string {
	return __NAValue__
}

func CLIMapAttribute() string {
	return __mapAttribute__
}
