package main

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
)

func main() {
	// optionParser := obioptions.GenerateOptionParser(obiconvert.OptionSet)

	// _, args, _ := optionParser(os.Args)

	// fs, _ := obiconvert.ReadBioSequences(args...)
	// //obicsv.CLIWriteCSV(fs, true)

	obiiter.WaitForLastPipe()

}
