package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obipairing"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obitagpcr"
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
	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	obioptions.SetWorkerPerCore(1)

	optionParser := obioptions.GenerateOptionParser(obitagpcr.OptionSet)

	optionParser(os.Args)
	pairs, err := obipairing.CLIPairedSequence()

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	paired := obitagpcr.IPCRTagPESequencesBatch(pairs,
		obipairing.CLIGapPenality(),
		obipairing.CLIDelta(),
		obipairing.CLIMinOverlap(),
		obipairing.CLIMinIdentity(),
		obipairing.CLIWithStats())

	obiconvert.CLIWriteBioSequences(paired, true)

	obiiter.WaitForLastPipe()
}