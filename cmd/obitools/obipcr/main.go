package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"runtime/trace"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obipcr"
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

	optionParser := obioptions.GenerateOptionParser(obipcr.OptionSet)

	_, args, _ := optionParser(os.Args)

	sequences, _ := obiconvert.ReadBioSequencesBatch(args...)
	amplicons, _ := obipcr.PCR(sequences)
	obiconvert.WriteBioSequencesBatch(amplicons, true)
}
