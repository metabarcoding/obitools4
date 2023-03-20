package obimultiplex

import (
	"fmt"
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obingslibrary"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _NGSFilterFile = ""
var _UnidentifiedFile = ""
var _AllowedMismatch = int(2)
var _AllowsIndel = false
var _ConservedError = false

// PCROptionSet defines every options related to a simulated PCR.
//
// The function adds to a CLI every options proposed to the user
// to tune the parametters of the PCR simulation algorithm.
//
// Parameters
//
// - option : is a pointer to a getoptions.GetOpt instance normaly
// produced by the
func MultiplexOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_NGSFilterFile, "tag-list", _NGSFilterFile,
		options.Alias("t"),
		options.Required("You must provide a tag list file following the NGSFilter format"),
		options.Description("File name of the NGSFilter file describing PCRs."))

	options.BoolVar(&_ConservedError, "keep-errors", _ConservedError,
		options.Description("Prints symbol counts."))

		options.BoolVar(&_AllowsIndel, "with-indels", _AllowsIndel,
		options.Description("Allows for indels during the primers matching."))

	options.StringVar(&_UnidentifiedFile, "unidentified", _UnidentifiedFile,
		options.Alias("u"),
		options.Description("Filename used to store the sequences unassigned to any sample."))

	options.IntVar(&_AllowedMismatch, "allowed-mismatches", _AllowedMismatch,
		options.Alias("e"),
		options.Description("Used to specify the number of errors allowed for matching primers."))

}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	MultiplexOptionSet(options)
}

func CLIAllowedMismatch() int {
	return _AllowedMismatch
}

func CLIAllowsIndel() bool {
	return _AllowsIndel
}
func CLIUnidentifiedFileName() string {
	return _UnidentifiedFile
}

func CLIConservedErrors() bool {
	return _UnidentifiedFile != "" || _ConservedError
}

func CLINGSFIlter() (obingslibrary.NGSLibrary, error) {
	file, err := os.Open(_NGSFilterFile)

	if err != nil {
		return nil, fmt.Errorf("open file error: %v", err)
	}

	ngsfiler, err := obiformats.ReadNGSFilter(file)

	if err != nil {
		return nil, fmt.Errorf("NGSfilter reading file error: %v", err)
	}

	return ngsfiler, nil
}
