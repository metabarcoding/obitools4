package obiconvert

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	log "github.com/sirupsen/logrus"

	"github.com/DavidGamba/go-getoptions"
)

var __skipped_entries__ = 0
var __read_only_entries__ = -1

var __no_ordered_input__ = false

var __input_fastjson_format__ = false
var __input_fastobi_format__ = false

var __input_ecopcr_format__ = false
var __input_embl_format__ = false
var __input_genbank_format__ = false
var __input_fastq_format__ = false
var __input_fasta_format__ = false

var __output_in_fasta__ = false
var __output_in_fastq__ = false
var __output_in_json__ = false
var __output_fastjson_format__ = false
var __output_fastobi_format__ = false

var __no_progress_bar__ = false
var __skip_empty__ = false

var __output_file_name__ = "-"
var __paired_file_name__ = ""

var __full_file_batch__ = false

func InputOptionSet(options *getoptions.GetOpt) {
	// options.IntVar(&__skipped_entries__, "skip", __skipped_entries__,
	// 	options.Description("The N first sequence records of the file are discarded from the analysis and not reported to the output file."))

	// options.IntVar(&__read_only_entries__, "only", __read_only_entries__,
	// 	options.Description("Only the N next sequence records of the file are analyzed. The following sequences in the file are neither analyzed, neither reported to the output file. This option can be used conjointly with the –skip option."))

	options.BoolVar(&__input_fastjson_format__, "input-json-header", __input_fastjson_format__,
		options.Description("FASTA/FASTQ title line annotations follow json format."))
	options.BoolVar(&__input_fastobi_format__, "input-OBI-header", __input_fastobi_format__,
		options.Description("FASTA/FASTQ title line annotations follow OBI format."))

	options.BoolVar(&__input_ecopcr_format__, "ecopcr", __input_ecopcr_format__,
		options.Description("Read data following the ecoPCR output format."))

	options.BoolVar(&__input_embl_format__, "embl", __input_embl_format__,
		options.Description("Read data following the EMBL flatfile format."))

	options.BoolVar(&__input_genbank_format__, "genbank", __input_genbank_format__,
		options.Description("Read data following the Genbank flatfile format."))

	options.BoolVar(&__input_fastq_format__, "fastq", __input_fastq_format__,
		options.Description("Read data following the fastq format."))

	options.BoolVar(&__input_fasta_format__, "fasta", __input_fasta_format__,
		options.Description("Read data following the fasta format."))

	options.BoolVar(&__no_ordered_input__, "no-order", __no_ordered_input__,
		options.Description("When several input files are provided, "+
			"indicates that there is no order among them."))

}

func OutputModeOptionSet(options *getoptions.GetOpt, compressed bool) {
	options.BoolVar(&__no_progress_bar__, "no-progressbar", false,
		options.Description("Disable the progress bar printing"))

	if compressed {
		options.BoolVar(obidefault.CompressedPtr(), "compressed", obidefault.CompressOutput(),
			options.Alias("Z"),
			options.Description("Compress all the result using gzip"))

	}

	options.StringVar(&__output_file_name__, "out", __output_file_name__,
		options.Alias("o"),
		options.ArgName("FILENAME"),
		options.Description("Filename used for saving the output"),
	)
}

func OutputOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__skip_empty__, "skip-empty", __skip_empty__,
		options.Description("Sequences of length equal to zero are suppressed from the output"))

	options.BoolVar(&__output_in_fasta__, "fasta-output", false,
		options.Description("Write sequence in fasta format (default if no quality data available)."))

	options.BoolVar(&__output_in_fastq__, "fastq-output", false,
		options.Description("Write sequence in fastq format (default if quality data available)."))

	options.BoolVar(&__output_in_json__, "json-output", false,
		options.Description("Write sequence in json format."))

	options.BoolVar(&__output_fastjson_format__, "output-json-header", false,
		options.Description("output FASTA/FASTQ title line annotations follow json format."))
	options.BoolVar(&__output_fastobi_format__, "output-OBI-header", false,
		options.Alias("O"),
		options.Description("output FASTA/FASTQ title line annotations follow OBI format."))

	OutputModeOptionSet(options, true)
}

func PairedFilesOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&__paired_file_name__, "paired-with", __paired_file_name__,
		options.ArgName("FILENAME"),
		options.Description("Filename containing the paired reads"),
	)
}

func OptionSet(options *getoptions.GetOpt) {
	obioptions.LoadTaxonomyOptionSet(options, false, false)
	InputOptionSet(options)
	OutputOptionSet(options)
	PairedFilesOptionSet(options)
}

// Returns true if the number of reads described in the
// file has to be printed.
func CLIInputFormat() string {
	switch {
	case __input_fasta_format__:
		return "fasta"
	case __input_fastq_format__:
		return "fastq"
	case __input_ecopcr_format__:
		return "ecopcr"
	case __input_embl_format__:
		return "embl"
	case __input_genbank_format__:
		return "genbank"
	default:
		return "guessed"
	}
}

// Returns true if the order among several imput files has not to be considered
func CLINoInputOrder() bool {
	return __no_ordered_input__
}

func CLIOutputFormat() string {
	switch {
	case __output_in_fastq__:
		return "fastq"
	case __output_in_fasta__:
		return "fasta"
	case __output_in_json__:
		return "json"
	default:
		return "guessed"
	}
}

func CLISkipEmpty() bool {
	return __skip_empty__
}

func CLIInputFastHeaderFormat() string {
	switch {
	case __input_fastjson_format__:
		return "json"
	case __input_fastobi_format__:
		return "obi"
	default:
		return "guessed"
	}
}

func CLIOutputFastHeaderFormat() string {
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
func CLISequencesToSkip() int {
	return __skipped_entries__
}

func CLIAnalyzeOnly() int {
	return __read_only_entries__
}

func CLIProgressBar() bool {
	// If the output is not a terminal, then we do not display the progress bar
	o, _ := os.Stderr.Stat()
	onTerminal := (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	if !onTerminal {
		log.Info("Stderr is redirected, progress bar disabled")
	}

	return onTerminal && !__no_progress_bar__
}

func CLIOutPutFileName() string {
	return __output_file_name__
}

func CLIHasPairedFile() bool {
	return __paired_file_name__ != ""
}
func CLIPairedFileName() string {
	return __paired_file_name__
}

func SetFullFileBatch() {
	__full_file_batch__ = true
}
func FullFileBatch() bool {
	return __full_file_batch__
}
