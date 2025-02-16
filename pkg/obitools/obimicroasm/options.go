package obimicroasm

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
	log "github.com/sirupsen/logrus"
)

var _ForwardFile = ""
var _ReverseFile = ""
var _ForwardPrimer string
var _ReversePrimer string
var _AllowedMismatch = 0
var _kmerSize = -1

var _saveGraph = "__@@NOSAVE@@__"

func MicroAsmOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_ForwardFile, "forward-reads", "",
		options.Alias("F"),
		options.ArgName("FILENAME_F"),
		options.Required("You must provide at a forward file"),
		options.Description("The file names containing the forward reads"))
	options.StringVar(&_ReverseFile, "reverse-reads", "",
		options.Alias("R"),
		options.ArgName("FILENAME_R"),
		options.Required("You must provide a reverse file"),
		options.Description("The file names containing the reverse reads"))
	options.StringVar(&_ForwardPrimer, "forward", "",
		options.Required("You must provide a forward primer"),
		options.Description("The forward primer used for the electronic PCR."))

	options.StringVar(&_ReversePrimer, "reverse", "",
		options.Required("You must provide a reverse primer"),
		options.Description("The reverse primer used for the electronic PCR."))

	options.IntVar(&_AllowedMismatch, "allowed-mismatches", 0,
		options.Alias("e"),
		options.Description("Maximum number of mismatches allowed for each primer."))
	options.IntVar(&_kmerSize, "kmer-size", _kmerSize,
		options.ArgName("SIZE"),
		options.Description("The size of the kmer used to build the consensus. "+
			"Default value = -1, which means that the kmer size is estimated from the data"),
	)

	options.StringVar(&_saveGraph, "save-graph", _saveGraph,
		options.Description("Creates a directory containing the set of DAG used by the obiclean clustering algorithm. "+
			"The graph files follow the graphml format."),
	)

}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	MicroAsmOptionSet(options)
}

// CLIForwardPrimer returns the sequence of the forward primer as indicated by the
// --forward command line option
func CLIForwardPrimer() string {
	pattern, err := obiapat.MakeApatPattern(_ForwardPrimer, _AllowedMismatch, false)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return _ForwardPrimer
}

// CLIReversePrimer returns the sequence of the reverse primer as indicated by the
// --reverse command line option
func CLIReversePrimer() string {
	pattern, err := obiapat.MakeApatPattern(_ReversePrimer, _AllowedMismatch, false)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return _ReversePrimer
}

// CLIAllowedMismatch returns the allowed mistmatch count between each
// primer and the sequences as indicated by the
// --allowed-mismatches|-e command line option
func CLIAllowedMismatch() int {
	return _AllowedMismatch
}

func CLIPairedSequence() (obiiter.IBioSequence, error) {
	forward, err := obiconvert.CLIReadBioSequences(_ForwardFile)
	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	reverse, err := obiconvert.CLIReadBioSequences(_ReverseFile)
	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	paired := forward.PairTo(reverse)

	return paired, nil
}

func CLIForwardFile() string {
	return _ForwardFile
}

// Returns true it the obliclean graphs must be saved
func CLISaveGraphToFiles() bool {
	return _saveGraph != "__@@NOSAVE@@__"
}

// It returns the directory where the graph files are saved
func CLIGraphFilesDirectory() string {
	return _saveGraph
}

// CLIKmerSize returns the value of the kmer size to use for building the consensus.
//
// The value of the kmer size is set by the user with the `-k` flag.
// The value -1 means that the kmer size is estimated as the minimum value that
// insure that no kmer are present more than one time in a sequence.
//
// No parameters.
// Returns an integer value.
func CLIKmerSize() int {
	return _kmerSize
}

func SetKmerSize(kmerSize int) {
	_kmerSize = kmerSize
}
