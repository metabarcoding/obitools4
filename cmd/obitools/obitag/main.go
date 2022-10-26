package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obitag"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obitag.OptionSet)

	_, args, _ := optionParser(os.Args)

	fs, _ := obiconvert.ReadBioSequencesBatch(args...)
	identified := obitag.AssignTaxonomy(fs)

	obiconvert.WriteBioSequencesBatch(identified, true)
}
