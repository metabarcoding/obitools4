package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiprimers"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
	obiprimers.SetReadingDefaults()

	optionParser := obioptions.GenerateOptionParser(
		"obiprimers",
		"designs the primer pairs amplifying a set of example sequences and evaluates the DNA barcode they delimit. "+
			"The taxonomy comes from the -t option, or is read from the taxonomic paths embedded in the sequences (obiannotate --taxonomic-path)",
		obiprimers.OptionSet)

	_, args := optionParser(os.Args)

	sequences, taxonomy, err := obiprimers.CLIReadSequencesAndTaxonomy(args)
	if err != nil {
		log.Fatalf("Cannot read the sequences: %v", err)
	}

	pairs, err := obiprimers.CLIPrimerPairs(sequences, taxonomy)
	if err != nil {
		log.Fatalf("%v", err)
	}

	obicsv.CLICSVWriter(obiprimers.PairsToCSVIterator(pairs, obiprimers.CLIReference() != ""), true)

	obiutils.WaitForLastPipe()
}
