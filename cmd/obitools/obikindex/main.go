package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obikindex"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(
		"obikindex",
		"builds a disk-based kmer index from sequence files",
		obikindex.OptionSet)

	_, args := optionParser(os.Args)

	sequences, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	obikindex.CLIBuildKmerIndex(sequences)
}
