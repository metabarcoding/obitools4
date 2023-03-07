package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obicsv"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obicsv.OptionSet)

	_, args := optionParser(os.Args)

	fs, _ := obiconvert.CLIReadBioSequences(args...)
	obicsv.CLIWriteCSV(fs, true)

	obiiter.WaitForLastPipe()

}
