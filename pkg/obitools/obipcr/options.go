package obipcr

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _Circular = false
var _ForwardPrimer string
var _ReversePrimer string
var _AllowedMismatch = 0
var _MinimumLength = 0
var _MaximumLength = -1

// PCROptionSet defines every options related to a simulated PCR.
//
// The function adds to a CLI every options proposed to the user
// to tune the parametters of the PCR simulation algorithm.
//
// Parameters
//
// - option : is a pointer to a getoptions.GetOpt instance normaly
// produced by the
func PCROptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&_Circular, "circular", false,
		options.Alias("c"),
		options.Description("Considers that sequences are [c]ircular."))

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
		options.Description("Maximum length of the barcode (primers excluded)."))
}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	PCROptionSet(options)
}

// ForwardPrimer returns the sequence of the forward primer as indicated by the
// --forward command line option
func ForwardPrimer() string {
	pattern, err := obiapat.MakeApatPattern(_ForwardPrimer, _AllowedMismatch)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return _ForwardPrimer
}

// ReversePrimer returns the sequence of the reverse primer as indicated by the
// --reverse command line option
func ReversePrimer() string {
	pattern, err := obiapat.MakeApatPattern(_ReversePrimer, _AllowedMismatch)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return _ReversePrimer
}

// AllowedMismatch returns the allowed mistmatch count between each
// primer and the sequences as indicated by the
// --allowed-mismatches|-e command line option
func AllowedMismatch() int {
	return _AllowedMismatch
}

// Circular returns the considered sequence topology as indicated by the
// --circular|-c command line option
func Circular() bool {
	return _Circular
}

// MinLength returns the amplicon minimum length as indicated by the
// --min-length|-l command line option
func MinLength() int {
	return _MinimumLength
}

// MaxLength returns the amplicon maximum length as indicated by the
// --max-length|-L command line option
func MaxLength() int {
	return _MaximumLength
}
