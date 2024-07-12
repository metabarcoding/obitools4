package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obirefidx"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obirefidx.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	indexed := obirefidx.IndexFamilyDB(fs)

	obiconvert.CLIWriteBioSequences(indexed, true)
	obiiter.WaitForLastPipe()

}
