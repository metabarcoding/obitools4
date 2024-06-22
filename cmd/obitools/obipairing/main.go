package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obipairing"
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

	optionParser := obioptions.GenerateOptionParser(obipairing.OptionSet)

	optionParser(os.Args)

	obioptions.SetStrictReadWorker(2)
	obioptions.SetStrictWriteWorker(2)
	pairs, err := obipairing.CLIPairedSequence()

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	paired := obipairing.IAssemblePESequencesBatch(pairs,
		obipairing.CLIGapPenality(),
		obipairing.CLIPenalityScale(),
		obipairing.CLIDelta(),
		obipairing.CLIMinOverlap(),
		obipairing.CLIMinIdentity(),
		obipairing.CLIFastMode(),
		obipairing.CLIFastRelativeScore(),
		obipairing.CLIWithStats(),
		obioptions.CLIParallelWorkers(),
	)

	obiconvert.CLIWriteBioSequences(paired, true)

	obiiter.WaitForLastPipe()
}
