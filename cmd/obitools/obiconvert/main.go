package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	obioptions.SetStrictReadWorker(2)
	obioptions.SetStrictWriteWorker(2)

	optionParser := obioptions.GenerateOptionParser(obiconvert.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	obiconvert.CLIWriteBioSequences(fs, true)

	obiiter.WaitForLastPipe()

}
