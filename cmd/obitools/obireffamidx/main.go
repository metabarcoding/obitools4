package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obirefidx"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(
		"obireffamidx",
		"",
		obirefidx.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	indexed := obirefidx.IndexFamilyDB(fs)

	obiconvert.CLIWriteBioSequences(indexed, true)
	obiutils.WaitForLastPipe()

}
