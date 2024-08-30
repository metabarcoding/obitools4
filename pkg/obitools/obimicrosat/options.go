package obimicrosat

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _MinUnitLength = 1
var _MaxUnitLength = 6
var _MinUnitCount = 5
var _MinLength = 20
var _MinFlankLength = 0
var _NotReoriented = false

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

	options.IntVar(&_MinLength, "min-length", _MinLength,
		options.Alias("l"),
		options.Description("Minimum length of a microsatellite."))

	options.IntVar(&_MinFlankLength, "min-flank-length", _MinFlankLength,
		options.Alias("f"),
		options.Description("Minimum length of the flanking sequences."))

	options.BoolVar(&_NotReoriented, "not-reoriented", _NotReoriented,
		options.Alias("n"),
		options.Description("Do not reorient the microsatellites."))
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

func CLIMinLength() int {
	return _MinLength
}

func CLIMinFlankLength() int {
	return _MinFlankLength
}

func CLIReoriented() bool {
	return !_NotReoriented
}
