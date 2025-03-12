package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	obidefault.SetStrictReadWorker(2)
	obidefault.SetStrictWriteWorker(2)

	optionParser := obioptions.GenerateOptionParser(
		"obiconvert",
		"convertion of sequence files to various formats",
		obiconvert.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	obiconvert.CLIWriteBioSequences(fs, true)

	obiutils.WaitForLastPipe()

}
