package main

import (
	"log"
	"os"
	"runtime/pprof"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obitools/obipairing"
)

func main() {

	// go tool pprof -http=":8000" ./obipairing ./cpu.pprof
	f, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// go tool trace cpu.trace
	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	option_parser := obioptions.GenerateOptionParser(obipairing.OptionSet)

	option_parser(os.Args)
	pairs, _ := obipairing.IBatchPairedSequence()
	paired := obipairing.IAssemblePESequencesBatch(pairs, 2, 50, 20, true)
	written, _ := obiformats.WriteFastqBatchToStdout(paired)
	written.Destroy()
}
