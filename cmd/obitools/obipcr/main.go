package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obipcr"
)

func main() {

	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	option_parser := obioptions.GenerateOptionParser(obipcr.OptionSet)

	_, args, _ := option_parser(os.Args)

	sequences, _ := obiconvert.ReadBioSequencesBatch(args...)
	amplicons, _ := obipcr.PCR(sequences)
	obiconvert.WriteBioSequences(amplicons)
}
