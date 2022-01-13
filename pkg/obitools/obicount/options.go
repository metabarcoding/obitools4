// obicount function utility package.
//
// The obitols/obicount package contains every
// functions specificaly required by the obicount utility.
package obicount

import (
	"github.com/DavidGamba/go-getoptions"
)

var __read_count__ bool
var __variant_count__ bool
var __symbol_count__ bool

func OptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__variant_count__, "variants", false,
		options.Alias("v"),
		options.Description("Prints variant counts."))

	options.BoolVar(&__read_count__, "reads", false,
		options.Alias("r"),
		options.Description("Prints read counts."))

	options.BoolVar(&__symbol_count__, "symbols", false,
		options.Alias("s"),
		options.Description("Prints symbol counts."))
}

// Returns true if the number of reads described in the
// file has to be printed.
func IsPrintingReadCount() bool {
	return __read_count__ ||
		!(__read_count__ || __variant_count__ || __symbol_count__)
}

// Returns true if the number of sequence variants described in the
// file has to be printed.
func IsPrintingVariantCount() bool {
	return __variant_count__ ||
		!(__read_count__ || __variant_count__ || __symbol_count__)
}

// Returns true if the number of symbols (sum of the sequence lengths)
// described in the file has to be printed.
func IsPrintingSymbolCount() bool {
	return __symbol_count__ ||
		!(__read_count__ || __variant_count__ || __symbol_count__)
}
