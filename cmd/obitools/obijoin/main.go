package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obijoin"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	obidefault.SetStrictReadWorker(2)
	obidefault.SetStrictWriteWorker(2)

	optionParser := obioptions.GenerateOptionParser(
		"obijoin",
		"merge annotations contained in a file to another file",
		obijoin.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	joined := obijoin.CLIJoinSequences(fs)

	obiconvert.CLIWriteBioSequences(joined, true)

	obiutils.WaitForLastPipe()

}
