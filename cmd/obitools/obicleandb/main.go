package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obicleandb"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obicleandb.OptionSet)

	_, args := optionParser(os.Args)

	fs, _ := obiconvert.CLIReadBioSequences(args...)

	cleaned := obicleandb.ICleanDB(fs)

	obiconvert.CLIWriteBioSequences(cleaned, true)
}
