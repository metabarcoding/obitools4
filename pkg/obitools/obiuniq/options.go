package obiuniq

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _StatsOn = make([]string, 0, 10)
var _Keys = make([]string, 0, 10)
var _InMemory = false
var _chunks = 100
var _NAValue = "NA"
var _NoSingleton = false

// UniqueOptionSet sets up unique options for the obiuniq command.
//
// It configures various options such as merging attributes, category attributes,
// defining the NA value, handling singleton sequences, choosing between in-memory
// or disk storage, and specifying the chunk count for dataset division.
func UniqueOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&_StatsOn, "merge",
		1, 1,
		options.Alias("m"),
		options.ArgName("KEY"),
		options.Description("Adds a merged attribute containing the list of sequence record ids merged within this group."))

	options.StringSliceVar(&_Keys, "category-attribute",
		1, 1,
		options.Alias("c"),
		options.ArgName("CATEGORY"),
		options.Description("Adds one attribute to the list of attributes used to define sequence groups (this option can be used several times)."))

	options.StringVar(&_NAValue, "na-value", _NAValue,
		options.ArgName("NA_NAME"),
		options.Description("Value used when the classifier tag is not defined for a sequence."))

	options.BoolVar(&_NoSingleton, "no-singleton", _NoSingleton,
		options.Description("If set, sequences occurring a single time in the data set are discarded."))

	options.BoolVar(&_InMemory, "in-memory", _InMemory,
		options.Description("Use memory instead of disk to store data chunks."))

	options.IntVar(&_chunks, "chunk-count", _chunks,
		options.Description("In how many chunk the dataset is pre-devided for speeding up the process."))

}

// OptionSet adds to the basic option set every options declared for
// the obiuniq command
//
// It takes a pointer to a GetOpt struct as its parameter and does not return anything.
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(false)(options)
	UniqueOptionSet(options)
}

// CLIStatsOn returns the list of variables on witch statistics are computed.
//
// It does not take any parameters.
// It returns a slice of strings representing the statistics on values.
func CLIStatsOn() []string {
	return _StatsOn
}

// SetStatsOn sets the list of variables on witch statistics are computed.
//
// It takes a slice of strings as its parameter and does not return anything.
func SetStatsOn(statsOn []string) {
	_StatsOn = statsOn
}

// AddStatsOn adds a variable to the list of variables on witch statistics are computed.
//
// Parameters:
// - statsOn: variadic strings representing the statistics to be added.
func AddStatsOn(statsOn ...string) {
	_StatsOn = append(_StatsOn, statsOn...)
}

// CLIKeys returns the keys used to distinguished among identical sequences.
//
// It does not take any parameters.
// It returns a slice of strings representing the keys used by the CLI.
func CLIKeys() []string {
	return _Keys
}

// CLIUniqueInMemory returns if the unique function is running in memory only.
//
// It does not take any parameters.
// It returns a boolean value indicating whether the function is running in memory or not.
func CLIUniqueInMemory() bool {
	return _InMemory
}

// SetUniqueInMemory sets whether the unique function is running in memory or not.
//
// inMemory bool - A boolean value indicating whether the function is running in memory.
// No return value.
func SetUniqueInMemory(inMemory bool) {
	_InMemory = inMemory
}

// CLINumberOfChunks returns the number of chunks used for the first bucket sort step used by the unique function.
//
// It does not take any parameters.
// It returns an integer representing the number of chunks.
func CLINumberOfChunks() int {
	if _chunks <= 1 {
		return 1
	}

	return _chunks
}

// SetNumberOfChunks sets the number of chunks used for the first bucket sort step used by the unique function.
//
// chunks int - The number of chunks to be set.
// No return value.
func SetNumberOfChunks(chunks int) {
	_chunks = chunks
}

// CLINAValue returns the value used as a placeholder for missing values.
//
// No parameters.
// Return type: string.
func CLINAValue() string {
	return _NAValue
}

// SetNAValue sets the NA value to the specified string.
//
// value string - The value to set as the NA value.
func SetNAValue(value string) {
	_NAValue = value
}

// CLINoSingleton returns a boolean value indicating whether or not singleton sequences should be discarded.
//
// No parameters.
// Returns a boolean value indicating whether or not singleton sequences should be discarded.
func CLINoSingleton() bool {
	return _NoSingleton
}

// SetNoSingleton sets the boolean value indicating whether or not singleton sequences should be discarded.
//
// noSingleton bool - The boolean value to set for _NoSingleton.
func SetNoSingleton(noSingleton bool) {
	_NoSingleton = noSingleton
}
