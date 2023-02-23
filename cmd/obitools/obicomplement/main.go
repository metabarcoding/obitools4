package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiconvert.OptionSet)

	_, args, _ := optionParser(os.Args)

	fs, _ := obiconvert.CLIReadBioSequences(args...)

	comp := fs.MakeIWorker(obiseq.ReverseComplementWorker(true))
	obiconvert.CLIWriteBioSequences(comp, true)

	obiiter.WaitForLastPipe()

}
