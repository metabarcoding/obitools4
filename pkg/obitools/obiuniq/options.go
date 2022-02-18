package obiuniq

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _StatsOn = make([]string, 0, 10)
var _Keys = make([]string, 0, 10)
var _OnDisk = false
var _chunks = 100
var _NAValue = "NA"

func UniqueOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&_StatsOn, "merge",
		1, 1,
		options.Alias("m"),
		options.Description("Adds a merged attribute containing the list of sequence record ids merged within this group."))

	options.StringSliceVar(&_Keys, "category-attribute",
		1, 1,
		options.Alias("c"),
		options.Description("Adds one attribute to the list of attributes used to define sequence groups (this option can be used several times)."))

	options.StringVar(&_NAValue, "na-value", _NAValue,
		options.Description("Value used when the classifier tag is not defined for a sequence."))

	options.BoolVar(&_OnDisk, "on-disk", true,
		options.Description("Allows for using a disk cache during the dereplication process. "))

	options.IntVar(&_chunks, "chunk-count", _chunks,
		options.Description("In how many chunk the dataset is pre-devided for speeding up the process."))

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

func CLIUniqueInMemory() bool {
	return _OnDisk
}

func CLINumberOfChunks() int {
	if _chunks <= 1 {
		return 1
	}

	return _chunks
}

func CLINAValue() string {
	return _NAValue
}
