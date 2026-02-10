package main

import (
	"context"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obik"
	"github.com/DavidGamba/go-getoptions"
)

func main() {
	defer obiseq.LogBioSeqStatus()

	opt, parser := obioptions.GenerateSubcommandParser(
		"obik",
		"Manage disk-based kmer indices",
		obik.OptionSet,
	)

	_, remaining := parser(os.Args)

	err := opt.Dispatch(context.Background(), remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			os.Exit(0)
		}
		log.Fatalf("Error: %v", err)
	}
}
