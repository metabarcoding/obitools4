package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obioligotag"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(
		"obioligotag",
		"designs a family of oligonucleotides (tags) at a minimum distance from each other",
		obioligotag.OptionSet)

	_, args := optionParser(os.Args)
	if len(args) > 0 {
		log.Fatalf("unexpected arguments %v: use --oligo-list to provide candidate oligonucleotides", args)
	}

	oligos, err := obioligotag.CLIOligoFamily()
	if err != nil {
		log.Fatalf("%v", err)
	}

	obiconvert.CLIWriteBioSequences(obioligotag.CLIOligosToIterator(oligos), true)

	obiutils.WaitForLastPipe()
}
