package obicsv

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	"github.com/DavidGamba/go-getoptions"
)

var _outputIds = false
var _outputCount = false
var _outputTaxon = false
var _outputSequence = false
var _outputQuality = false
var _outputDefinition = false
var _obipairing = false
var _autoColumns = false
var _keepOnly = make([]string, 0)
var _naValue = "NA"

var _softAttributes = map[string][]string{
	"obipairing": {"mode", "seq_a_single", "seq_b_single",
		"ali_dir", "score", "score_norm",
		"seq_ab_match", "pairing_mismatches",
	},
}

func CSVOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&_outputIds, "ids", _outputIds,
		options.Alias("i"),
		options.Description("Prints sequence ids in the ouput."))

	options.BoolVar(&_outputSequence, "sequence", _outputSequence,
		options.Alias("s"),
		options.Description("Prints sequence itself in the output."))

	options.BoolVar(&_outputQuality, "quality", _outputQuality,
		options.Alias("q"),
		options.Description("Prints sequence quality in the output."))

	options.BoolVar(&_outputDefinition, "definition", _outputDefinition,
		options.Alias("d"),
		options.Description("Prints sequence definition in the output."))

	options.BoolVar(&_autoColumns, "auto", _autoColumns,
		options.Description("Based on the first sequences, propose a list of attibutes to print"))

	options.BoolVar(&_outputCount, "count", _outputCount,
		options.Description("Prints the count attribute in the output"))

	options.BoolVar(&_outputTaxon, "taxon", _outputTaxon,
		options.Description("Prints the NCBI taxid and its related scientific name"))

	options.BoolVar(&_obipairing, "obipairing", _obipairing,
		options.Description("Prints the attributes added by obipairing"))

	options.StringSliceVar(&_keepOnly, "keep", 1, 1,
		options.Alias("k"),
		options.ArgName("KEY"),
		options.Description("Keeps only attribute with key <KEY>. Several -k options can be combined."))

	options.StringVar(&_naValue, "na-value", _naValue,
		options.ArgName("NAVALUE"),
		options.Description("A string representing non available values in the CSV file."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OutputModeOptionSet(options)
	CSVOptionSet(options)
}

func CLIPrintId() bool {
	return _outputIds
}

func CLIPrintSequence() bool {
	return _outputSequence
}

func CLIPrintCount() bool {
	if i := obiutils.LookFor(_keepOnly, "count"); i >= 0 {
		_keepOnly = obiutils.RemoveIndex(_keepOnly, i)
		_outputCount = true
	}

	return _outputCount
}
func CLIPrintTaxon() bool {
	if i := obiutils.LookFor(_keepOnly, "taxid"); i >= 0 {
		_keepOnly = obiutils.RemoveIndex(_keepOnly, i)
		_outputTaxon = true
	}

	if i := obiutils.LookFor(_keepOnly, "scientific_name"); i >= 0 {
		_keepOnly = obiutils.RemoveIndex(_keepOnly, i)
		_outputTaxon = true
	}

	return _outputTaxon
}
func CLIPrintQuality() bool {
	return _outputQuality
}

func CLIPrintDefinition() bool {
	return _outputDefinition
}

func CLIAutoColumns() bool {
	return _autoColumns
}

func CLIHasToBeKeptAttributes() bool {
	return len(_keepOnly) > 0
}

func CLIToBeKeptAttributes() []string {
	if _obipairing {
		_keepOnly = append(_keepOnly, _softAttributes["obipairing"]...)
	}

	if i := obiutils.LookFor(_keepOnly, "count"); i >= 0 {
		_keepOnly = obiutils.RemoveIndex(_keepOnly, i)
		_outputCount = true
	}

	if i := obiutils.LookFor(_keepOnly, "taxid"); i >= 0 {
		_keepOnly = obiutils.RemoveIndex(_keepOnly, i)
		_outputTaxon = true
	}

	if i := obiutils.LookFor(_keepOnly, "scientific_name"); i >= 0 {
		_keepOnly = obiutils.RemoveIndex(_keepOnly, i)
		_outputTaxon = true
	}

	return _keepOnly
}

func CLINAValue() string {
	return _naValue
}
