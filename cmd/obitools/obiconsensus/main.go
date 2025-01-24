package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconsensus"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiconsensus.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	cleaned := obiconsensus.CLIOBIMinion(fs)

	obiconvert.CLIWriteBioSequences(cleaned, true)

	obiutils.WaitForLastPipe()

}
