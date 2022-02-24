package obiconvert

import (
	"github.com/DavidGamba/go-getoptions"
)

var __skipped_entries__ = 0
var __read_only_entries__ = -1

var __input_fastjson_format__ = false
var __input_fastobi_format__ = false

var __input_ecopcr_format__ = false
var __input_embl_format__ = false

var __input_solexa_quality__ = false

var __output_in_fasta__ = false
var __output_in_fastq__ = false
var __output_fastjson_format__ = false
var __output_fastobi_format__ = false
var __output_solexa_quality__ = false

func InputOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&__skipped_entries__, "skip", 0,
		options.Description("The N first sequence records of the file are discarded from the analysis and not reported to the output file."))

	options.IntVar(&__read_only_entries__, "only", -1,
		options.Description("Only the N next sequence records of the file are analyzed. The following sequences in the file are neither analyzed, neither reported to the output file. This option can be used conjointly with the –skip option."))

	options.BoolVar(&__input_fastjson_format__, "input-json-header", false,
		options.Description("FASTA/FASTQ title line annotations follow json format."))
	options.BoolVar(&__input_fastobi_format__, "input-OBI-header", false,
		options.Description("FASTA/FASTQ title line annotations follow OBI format."))

	options.BoolVar(&__input_ecopcr_format__, "ecopcr", false,
		options.Description("Read data following the ecoPCR output format."))

	options.BoolVar(&__input_embl_format__, "embl", false,
		options.Description("Read data following the EMBL flatfile format."))

	options.BoolVar(&__input_solexa_quality__, "solexa", false,
		options.Description("Decodes quality string according to the Solexa specification."))

}

func OutputOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__output_in_fasta__, "fasta-output", false,
		options.Description("Read data following the ecoPCR output format."))

	options.BoolVar(&__output_in_fastq__, "fastq-output", false,
		options.Description("Read data following the EMBL flatfile format."))

	options.BoolVar(&__output_fastjson_format__, "output-json-header", false,
		options.Description("output FASTA/FASTQ title line annotations follow json format."))
	options.BoolVar(&__output_fastobi_format__, "output-OBI-header", false,
		options.Alias("O"),
		options.Description("output FASTA/FASTQ title line annotations follow OBI format."))

	options.BoolVar(&__output_solexa_quality__, "solexa-output", false,
		options.Description("Encodes quality string according to the Solexa specification."))
}

func OptionSet(options *getoptions.GetOpt) {
	InputOptionSet(options)
	OutputOptionSet(options)
}

// Returns true if the number of reads described in the
// file has to be printed.
func InputFormat() string {
	switch {
	case __input_ecopcr_format__:
		return "ecopcr"
	case __input_embl_format__:
		return "embl"
	default:
		return "guessed"
	}
}

func OutputFormat() string {
	switch {
	case __output_in_fastq__:
		return "fastq"
	case __output_in_fasta__:
		return "fasta"
	default:
		return "guessed"
	}
}

func InputFastHeaderFormat() string {
	switch {
	case __input_fastjson_format__:
		return "json"
	case __input_fastobi_format__:
		return "obi"
	default:
		return "guessed"
	}
}

func OutputFastHeaderFormat() string {
	switch {
	case __output_fastjson_format__:
		return "json"
	case __output_fastobi_format__:
		return "obi"
	default:
		return "json"
	}
}

// Returns the count of sequences to skip at the beginning of the
// processing.
func SequencesToSkip() int {
	return __skipped_entries__
}

func AnalyzeOnly() int {
	return __read_only_entries__
}

func InputQualityShift() int {
	if __input_solexa_quality__ {
		return 64
	} else {
		return 33
	}
}

func OutputQualityShift() int {
	if __output_solexa_quality__ {
		return 64
	} else {
		return 33
	}
}
