package main

import (
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiannotate"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
)

func main() {

	defer obiseq.LogBioSeqStatus()

	// go tool pprof -http=":8000" ./obipairing ./cpu.pprof
	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// go tool trace cpu.trace
	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	optionParser := obioptions.GenerateOptionParser(obiannotate.OptionSet)

	_, args, _ := optionParser(os.Args)

	sequences, _ := obiconvert.ReadBioSequences(args...)
	annotator := obiannotate.CLIAnnotationPipeline()
	obiconvert.WriteBioSequences(sequences.Pipe(annotator), true)
}