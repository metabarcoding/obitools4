package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiclust"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(
		"obiclust",
		"star clustering of sequences (the algorithm of sumaclust, UCLUST and CD-HIT)",
		obiclust.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	clustered := obiclust.CLIOBIClust(fs)

	obiconvert.CLIWriteBioSequences(clustered, true)

	obiutils.WaitForLastPipe()
}
