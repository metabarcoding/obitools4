package obimicrosat

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _MinUnitLength = 1
var _MaxUnitLength = 6
var _MinUnitCount = 5

// PCROptionSet defines every options related to a simulated PCR.
//
// The function adds to a CLI every options proposed to the user
// to tune the parametters of the PCR simulation algorithm.
//
// # Parameters
//
// - option : is a pointer to a getoptions.GetOpt instance normaly
// produced by the
func MicroSatelliteOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&_MinUnitLength, "min-unit-length", _MinUnitLength,
		options.Alias("m"),
		options.Description("Minimum length of a microsatellite unit."))

	options.IntVar(&_MaxUnitLength, "max-unit-length", _MaxUnitLength,
		options.Alias("M"),
		options.Description("Maximum length of a microsatellite unit."))

	options.IntVar(&_MinUnitCount, "min-unit-count", _MinUnitCount,
		options.Description("Minumum number of repeated units."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	MicroSatelliteOptionSet(options)
}

func CLIMinUnitLength() int {
	return _MinUnitLength
}

func CLIMaxUnitLength() int {
	return _MaxUnitLength
}

func CLIMinUnitCount() int {
	return _MinUnitCount
}

func CLIMicroSatRegex() string {
	return fmt.Sprintf("([acgt]{%d,%d})\\1{%d}", _MinUnitLength, _MaxUnitLength, _MinUnitCount-1)
}
