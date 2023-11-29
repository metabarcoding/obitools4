package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiclean"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiclean.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	cleaned := obiclean.CLIOBIClean(fs)

	obiconvert.CLIWriteBioSequences(cleaned, true)

	obiiter.WaitForLastPipe()

}
