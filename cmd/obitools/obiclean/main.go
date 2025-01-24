package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiclean"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiclean.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	cleaned := obiclean.CLIOBIClean(fs)

	obiconvert.CLIWriteBioSequences(cleaned, true)

	obiutils.WaitForLastPipe()

}
