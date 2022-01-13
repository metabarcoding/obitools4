package obipcr

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var __circular__ = false
var __forward_primer__ string
var __reverse_primer__ string
var __allowed_mismatch__ = 0
var __minimum_length__ = 0
var __maximum_length__ = -1

func PCROptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__circular__, "circular", false,
		options.Alias("c"),
		options.Description("Considers that sequences are [c]ircular."))

	options.StringVar(&__forward_primer__, "forward", "",
		options.Required("You must provide a forward primer"),
		options.Description("The forward primer used for the electronic PCR."))

	options.StringVar(&__reverse_primer__, "reverse", "",
		options.Required("You must provide a reverse primer"),
		options.Description("The reverse primer used for the electronic PCR."))

	options.IntVar(&__allowed_mismatch__, "allowed-mismatches", 0,
		options.Alias("e"),
		options.Description("Maximum number of mismatches allowed for each primer."))

	options.IntVar(&__minimum_length__, "min-length", 0,
		options.Alias("l"),
		options.Description("Minimum length of the barcode (primers excluded)."))
	options.IntVar(&__maximum_length__, "max-length", -1,
		options.Alias("L"),
		options.Description("Maximum length of the barcode (primers excluded)."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	PCROptionSet(options)
}

func ForwardPrimer() string {
	pattern, err := obiapat.MakeApatPattern(__forward_primer__, __allowed_mismatch__)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return __forward_primer__
}

func ReversePrimer() string {
	pattern, err := obiapat.MakeApatPattern(__reverse_primer__, __allowed_mismatch__)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	pattern.Free()

	return __reverse_primer__
}

func AllowedMismatch() int {
	return __allowed_mismatch__
}

func Circular() bool {
	return __circular__
}

func MinLength() int {
	return __minimum_length__
}

func MaxLength() int {
	return __maximum_length__
}
