package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	option_parser := obioptions.GenerateOptionParser(obiconvert.OptionSet)

	_, args, _ := option_parser(os.Args)

	fs, _ := obiconvert.ReadBioSequences(args...)
	obiconvert.WriteBioSequences(fs)
}
