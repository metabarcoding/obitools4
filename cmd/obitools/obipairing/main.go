package main

import (
	"log"
	"os"
	"runtime/trace"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obipairing"
)

func main() {

	// go tool pprof -http=":8000" ./obipairing ./cpu.pprof
	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// go tool trace cpu.trace
	ftrace, err := os.Create("cpu.trace")
	if err != nil {
		log.Fatal(err)
	}
	trace.Start(ftrace)
	defer trace.Stop()

	optionParser := obioptions.GenerateOptionParser(obipairing.OptionSet)

	optionParser(os.Args)
	pairs, _ := obipairing.IBatchPairedSequence()
	paired := obipairing.IAssemblePESequencesBatch(pairs, 2, 50, 20, true)
	obiconvert.WriteBioSequencesBatch(paired, true)
}
