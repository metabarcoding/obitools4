package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obirefidx"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obirefidx.OptionSet)

	_, args, _ := optionParser(os.Args)

	fs, _ := obiconvert.ReadBioSequences(args...)
	indexed := obirefidx.IndexReferenceDB(fs)

	written, _ := obiconvert.WriteBioSequences(indexed, false)
	written.Consume()
}