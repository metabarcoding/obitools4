// It adds to a CLI every options proposed to the user to tune the parametters of the PCR simulation
// algorithm
package obipcr

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _Circular = false
var _ForwardPrimer string
var _ReversePrimer string
var _AllowedMismatch = 0
var _MinimumLength = 0
var _MaximumLength = -1
var _Fragmented = false
var _Delta = -1
var _OnlyFull = false

// PCROptionSet defines every options related to a simulated PCR.
//
// The function adds to a CLI every options proposed to the user
// to tune the parametters of the PCR simulation algorithm.
//
// # Parameters
//
// - option : is a pointer to a getoptions.GetOpt instance normaly
// produced by the
func PCROptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&_Circular, "circular", false,
		options.Alias("c"),
		options.Description("Considers that sequences are [c]ircular."))

	options.BoolVar(&_Fragmented, "fragmented", false,
		options.Description("Fragments long sequences in overlaping fragments to speedup computations"))

	options.StringVar(&_ForwardPrimer, "forward", "",
		options.Required("You must provide a forward primer"),
		options.Description("The forward primer used for the electronic PCR."))

	options.StringVar(&_ReversePrimer, "reverse", "",
		options.Required("You must provide a reverse primer"),
		options.Description("The reverse primer used for the electronic PCR."))

	options.IntVar(&_AllowedMismatch, "allowed-mismatches", 0,
		options.Alias("e"),
		options.Description("Maximum number of mismatches allowed for each primer."))

	options.IntVar(&_MinimumLength, "min-length", 0,
		options.Alias("l"),
		options.Description("Minimum length of the barcode (primers excluded)."))
	options.IntVar(&_MaximumLength, "max-length", -1,
		options.Alias("L"),
		options.Required("You must indicate the maximum size of the amplicon (excluded primer length)"),
		options.Description("Maximum length of the barcode (primers excluded)."))
	options.IntVar(&_Delta, "delta", -1,
		options.Alias("D"),
		options.Description("Lenght of the sequence fragment to be added to the barcode extremities."))
	options.BoolVar(&_OnlyFull, "only-complete-flanking", false,
		options.Description("Only fragments with complete flanking sequences are printed."))

}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	PCROptionSet(options)
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

// CLICircular returns the considered sequence topology as indicated by the
// --circular|-c command line option
func CLICircular() bool {
	return _Circular
}

// CLIMinLength returns the amplicon minimum length as indicated by the
// --min-length|-l command line option
func CLIMinLength() int {
	return _MinimumLength
}

// CLIMaxLength returns the amplicon maximum length as indicated by the
// --max-length|-L command line option
func CLIMaxLength() int {
	return _MaximumLength
}

func CLIFragmented() bool {
	return _Fragmented
}

func CLIWithExtension() bool {
	return _Delta >= 0
}

func CLIExtension() int {
	return _Delta
}

func CLIOnlyFull() bool {
	return _OnlyFull
}
