package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obidemerge"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	obidefault.SetStrictReadWorker(2)
	obidefault.SetStrictWriteWorker(2)

	optionParser := obioptions.GenerateOptionParser(obidemerge.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	demerged := obidemerge.CLIDemergeSequences(fs)

	obiconvert.CLIWriteBioSequences(demerged, true)

	obiutils.WaitForLastPipe()

}
