package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicleandb"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	obioptions.SetBatchSize(10)

	optionParser := obioptions.GenerateOptionParser(obicleandb.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	cleaned := obicleandb.ICleanDB(fs)

	toconsume, _ := obiconvert.CLIWriteBioSequences(cleaned, false)
	toconsume.Consume()

	obiutils.WaitForLastPipe()
}
