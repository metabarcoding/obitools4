package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obidemerge"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {
	obioptions.SetStrictReadWorker(2)
	obioptions.SetStrictWriteWorker(2)

	optionParser := obioptions.GenerateOptionParser(obidemerge.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	demerged := obidemerge.CLIDemergeSequences(fs)

	obiconvert.CLIWriteBioSequences(demerged, true)

	obiiter.WaitForLastPipe()

}
