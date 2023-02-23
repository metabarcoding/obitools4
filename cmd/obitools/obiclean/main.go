package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiclean"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obiclean.OptionSet)

	_, args, _ := optionParser(os.Args)

	fs, _ := obiconvert.CLIReadBioSequences(args...)

	cleaned := obiclean.IOBIClean(fs)

	obiconvert.CLIWriteBioSequences(cleaned, true)

	obiiter.WaitForLastPipe()

}
