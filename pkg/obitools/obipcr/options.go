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

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	PCROptionSet(options)
}

func ForwardPrimer() string {
	pattern, err := obiapat.MakeApatPattern(_ForwardPrimer, _AllowedMismatch)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return _ForwardPrimer
}

func ReversePrimer() string {
	pattern, err := obiapat.MakeApatPattern(_ReversePrimer, _AllowedMismatch)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return _ReversePrimer
}

func AllowedMismatch() int {
	return _AllowedMismatch
}

func Circular() bool {
	return _Circular
}

func MinLength() int {
	return _MinimumLength
}

func MaxLength() int {
	return _MaximumLength
}
