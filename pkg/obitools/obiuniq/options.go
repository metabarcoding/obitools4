package obiuniq

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _StatsOn = make([]string, 0, 10)
var _Keys = make([]string, 0, 10)

func UniqueOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&_StatsOn, "merge",
		1, 1000,
		options.Alias("m"),
		options.Description("Adds a merged attribute containing the list of sequence record ids merged within this group."))
	options.StringSliceVar(&_Keys, "category-attribute",
		1, 1000,
		options.Alias("c"),
		options.Description("Adds one attribute to the list of attributes used to define sequence groups (this option can be used several times)."))

}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	UniqueOptionSet(options)
}

func CLIStatsOn() []string {
	return _StatsOn
}

func CLIKeys() []string {
	return _Keys
}
