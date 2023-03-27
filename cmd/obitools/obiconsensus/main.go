package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconsensus"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiconvert.OptionSet)

	_, args := optionParser(os.Args)
	obiconvert.SetFullFileBatch()

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	consensus := obiconsensus.Consensus(fs, 0.95)
	obiconvert.CLIWriteBioSequences(consensus, true)

	obiiter.WaitForLastPipe()

}
