// obicount function utility package.
//
// The obitols/obicount package contains every
// functions specificaly required by the obicount utility.
package obimatrix

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"github.com/DavidGamba/go-getoptions"
)

var __threeColumns__ = false
var __transpose__ = true
var __mapAttribute__ = "merged_sample"
var __valueName__ = "count"
var __sampleName__ = "sample"
var __MapNAValue__ = "0"
var __AllowEmpty__ = false

func MatrixOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__threeColumns__, "three-columns", false,
		options.Description("Printouts the matrix in tree column format."))

	options.BoolVar(&__transpose__, "transpose", __transpose__,
		options.Description("Printouts the transposed matrix."))

	options.StringVar(&__mapAttribute__, "map", __mapAttribute__,
		options.Description("Which attribute is usd to produce th matrix."))

	options.StringVar(&__valueName__, "value-name", __valueName__,
		options.Description("Name of the coulumn containing the values in the three column format."))

	options.StringVar(&__sampleName__, "sample-name", __sampleName__,
		options.Description("Name of the coulumn containing the sample names in the three column format."))

	options.StringVar(&__MapNAValue__, "map-na-value", __MapNAValue__,
		options.Description("Value used when the map attribute is not defined for a sequence."))

	options.BoolVar(&__AllowEmpty__, "allow-empty", __AllowEmpty__,
		options.Description("Allow sequences with empty map"))
}

func OptionSet(options *getoptions.GetOpt) {
	MatrixOptionSet(options)
	obicsv.CSVOptionSet(options)
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

func CLIMapNaValue() string {
	return __MapNAValue__
}

func CLIMapAttribute() string {
	return __mapAttribute__
}

func CLITranspose() bool {
	return __transpose__
}

func CLIStrict() bool {
	return !__AllowEmpty__
}
