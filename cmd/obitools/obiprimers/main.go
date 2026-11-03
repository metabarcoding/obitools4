package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiprimers"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(
		"obiprimers",
		"designs the primer pairs amplifying a set of example sequences and evaluates the DNA barcode they delimit",
		obiprimers.OptionSet)

	_, args := optionParser(os.Args)

	if !obidefault.HasSelectedTaxonomy() {
		log.Fatal("you must indicate a taxonomy using the -t or --taxonomy option")
	}

	sequences, err := obiconvert.CLIReadBioSequences(args...)
	if err != nil {
		log.Fatalf("Cannot open the sequence files: %v", err)
	}

	pairs, err := obiprimers.CLIPrimerPairs(sequences)
	if err != nil {
		log.Fatalf("%v", err)
	}

	obicsv.CLICSVWriter(obiprimers.PairsToCSVIterator(pairs), true)

	obiutils.WaitForLastPipe()
}
