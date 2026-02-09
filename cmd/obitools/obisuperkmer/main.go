package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obisuperkmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
	// Generate option parser
	optionParser := obioptions.GenerateOptionParser(
		"obisuperkmer",
		"extract super k-mers from sequence files",
		obisuperkmer.OptionSet)

	// Parse command-line arguments
	_, args := optionParser(os.Args)

	// Read input sequences
	sequences, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	// Extract super k-mers
	superkmers := obisuperkmer.CLIExtractSuperKmers(sequences)

	// Write output sequences
	obiconvert.CLIWriteBioSequences(superkmers, true)

	// Wait for pipeline completion
	obiutils.WaitForLastPipe()
}
