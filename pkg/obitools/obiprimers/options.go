// obiprimers function utility package.
//
// The obitools/obiprimers package contains every function specific to the
// obiprimers utility, the OBITools4 version of ecoPrimers: it designs the
// primer pairs that best amplify a set of example sequences, optionally
// avoiding a set of counterexample sequences, and measures the quality of the
// DNA marker each pair delimits.
package obiprimers

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiprimer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _restrict = []string{}
var _ignore = []string{}
var _except = []string{}
var _rank = "species"

var _primerLength = 18
var _allowedMismatch = 3
var _minLength = 0
var _maxLength = 1000
var _threePrime = 0
var _circular = false
var _noMultiMatch = false
var _reference = ""

var _strictQuorum = 0.7
var _sensitivity = 0.9
var _falsePositive = 0.1
var _specificity = 0.6

var _saltMethod = "santalucia"
var _salt = obithermo.DefaultSalt
var _primerConc = obithermo.DefaultPrimerConc

var _partitions = 0

// OptionSet registers the options of obiprimers.
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputModeOptionSet(options, true)
	obioptions.LoadTaxonomyOptionSet(options, false, false)

	options.StringSliceVar(&_restrict, "restrict-to-taxon", 1, 1,
		options.Alias("r"),
		options.ArgName("TAXID"),
		options.Description("Restricts the examples (the targets, the sequences the primers must amplify) to "+
			"the sequences belonging to this taxon. The other sequences are non-targets: they are not "+
			"required to be amplified, but the barcodes they give are taken into account to decide "+
			"whether a target is identified (Bs, Bd). Can be repeated. Without this option every "+
			"sequence is an example."))

	options.StringSliceVar(&_ignore, "ignore-taxon", 1, 1,
		options.Alias("i"),
		options.ArgName("TAXID"),
		options.Description("Declares the sequences belonging to this taxon as counterexamples "+
			"(sequences the primers must not amplify). Can be repeated."))

	options.StringSliceVar(&_except, "exception-taxon", 1, 1,
		options.Alias("E"),
		options.ArgName("TAXID"),
		options.Description("Declares a sub-clade of the examples as counterexamples. Can be repeated."))

	options.StringVar(&_rank, "rank", _rank,
		options.ArgName("RANK"),
		options.Description("Taxonomic rank at which coverage (Bc) and specificity (Bs) are computed."))

	options.BoolVar(&_circular, "circular", _circular,
		options.Alias("c"),
		options.Description("Considers the sequences as circular (whole organelle genomes)."))

	options.BoolVar(&_noMultiMatch, "no-multi-match", _noMultiMatch,
		options.Alias("U"),
		options.Description("Discards the primers that match, with the allowed mismatches, more than once "+
			"in any sequence (examples and counterexamples, both strands)."))

	options.StringVar(&_reference, "reference", _reference,
		options.Alias("R"),
		options.ArgName("SEQUENCE_ID"),
		options.Description("Identifier of an example sequence taken as reference: for each pair amplifying it, "+
			"adds the location of the primer sites and the amplified sequence."))

	options.IntVar(&_primerLength, "primer-length", _primerLength,
		options.ArgName("LENGTH"),
		options.Description("Size of the primers, from 2 to 31."))

	options.IntVar(&_allowedMismatch, "allowed-mismatches", _allowedMismatch,
		options.Alias("e"),
		options.ArgName("ERRORS"),
		options.Description("Maximum number of substitutions between a primer and a site it matches. "+
			"The search is fast up to about 5; beyond, the number of matches explodes."))

	options.IntVar(&_minLength, "min-length", _minLength,
		options.Alias("l"),
		options.ArgName("LENGTH"),
		options.Description("The amplified region, primers excluded, must be longer than this."))

	options.IntVar(&_maxLength, "max-length", _maxLength,
		options.Alias("L"),
		options.ArgName("LENGTH"),
		options.Description("The amplified region, primers excluded, must be shorter than this."))

	options.IntVar(&_threePrime, "three-prime", _threePrime,
		options.Alias("3"),
		options.ArgName("BASES"),
		options.Description("Number of bases at the 3' end of each primer that must match the "+
			"amplified sequence exactly (an ambiguous base is not a match)."))

	options.Float64Var(&_strictQuorum, "strict-quorum", _strictQuorum,
		options.Alias("q"),
		options.ArgName("FRACTION"),
		options.Description("Fraction of the examples in which a primer word must be found without any error."))

	options.Float64Var(&_sensitivity, "sensitivity", _sensitivity,
		options.Alias("s"),
		options.ArgName("FRACTION"),
		options.Description("Fraction of the examples a primer must match (with allowed mismatches), "+
			"and a pair must amplify."))

	options.Float64Var(&_falsePositive, "false-positive", _falsePositive,
		options.Alias("x"),
		options.ArgName("FRACTION"),
		options.Description("Fraction of the counterexamples a good primer may match, and a pair may amplify."))

	options.Float64Var(&_specificity, "specificity", _specificity,
		options.Alias("T"),
		options.ArgName("FRACTION"),
		options.Description("Pairs whose specificity Bs is lower than this are dropped."))

	options.StringVar(&_saltMethod, "salt-correction", _saltMethod,
		options.ArgName("METHOD"),
		options.Description("Salt correction of the melting temperatures: 'santalucia' or 'owczarzy'."))

	options.Float64Var(&_salt, "salt", _salt,
		options.ArgName("MOLAR"),
		options.Description("Monovalent salt concentration (M) used for the melting temperatures."))

	options.Float64Var(&_primerConc, "primer-concentration", _primerConc,
		options.ArgName("MOLAR"),
		options.Description("Primer concentration (M) used for the melting temperatures."))

	options.IntVar(&_partitions, "partitions", _partitions,
		options.ArgName("COUNT"),
		options.Description("Number of temporary partition files used to count the conserved words "+
			"(0 = chosen from the size of the examples: 256 to 1024). The memory of the count is bounded "+
			"whatever the number; more files only make the largest partitions smaller."))
}

// CLIRestrictTaxa returns the taxids restricting the examples.
func CLIRestrictTaxa() []string { return _restrict }

// CLIIgnoreTaxa returns the taxids of the counterexample clades.
func CLIIgnoreTaxa() []string { return _ignore }

// CLIExceptionTaxa returns the taxids of the sub-clades turned into counterexamples.
func CLIExceptionTaxa() []string { return _except }

// CLIReference returns the identifier of the reference sequence, or "".
func CLIReference() string { return _reference }

// CLIRank returns the taxonomic rank of the coverage and specificity indices.
func CLIRank() string { return _rank }

// CLISearchParams builds the search parameters from the command line.
func CLISearchParams() (obiprimer.SearchParams, error) {
	var method obithermo.SaltMethod
	switch _saltMethod {
	case "santalucia":
		method = obithermo.SantaLucia
	case "owczarzy":
		method = obithermo.Owczarzy
	default:
		return obiprimer.SearchParams{}, fmt.Errorf("unknown salt correction %q: use 'santalucia' or 'owczarzy'", _saltMethod)
	}
	model, err := obithermo.NewModel(_primerConc, obithermo.DefaultTargetConc, _salt, method)
	if err != nil {
		return obiprimer.SearchParams{}, err
	}

	return obiprimer.SearchParams{
		PrimerLength:   _primerLength,
		MaxErrors:      _allowedMismatch,
		MinLength:      _minLength,
		MaxLength:      _maxLength,
		ThreePrime:     _threePrime,
		StrictQuorum:   _strictQuorum,
		Sensitivity:    _sensitivity,
		FalsePositive:  _falsePositive,
		MinSpecificity: _specificity,
		Circular:       _circular,
		NoMultiMatch:   _noMultiMatch,
		Reference:      _reference,
		Partitions:     _partitions,
		Model:          model,
		Release:        releaseMemory,
	}, nil
}
