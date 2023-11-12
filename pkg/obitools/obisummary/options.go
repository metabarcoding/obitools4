// obicount function utility package.
//
// The obitols/obicount package contains every
// functions specificaly required by the obicount utility.
package obisummary

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var __json_output__ = false
var __yaml_output__ = false

func SummaryOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__json_output__, "json-output", false,
		options.Description("Print results as JSON record."))

	options.BoolVar(&__yaml_output__, "yaml-output", false,
		options.Description("Print results as YAML record."))
}

func OptionSet(options *getoptions.GetOpt) {
	SummaryOptionSet(options)
	obiconvert.InputOptionSet(options)
}

func CLIOutFormat() string {
	if __yaml_output__ && !__json_output__ {
		return "yaml"
	}

	return "json"
}
