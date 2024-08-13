package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiconvert.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	comp := fs.MakeIWorker(obiseq.ReverseComplementWorker(true), true)
	obiconvert.CLIWriteBioSequences(comp, true)

	obiiter.WaitForLastPipe()

}
