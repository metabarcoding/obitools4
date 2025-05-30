package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obipairing"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obitagpcr"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
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

	obidefault.SetWorkerPerCore(1)

	optionParser := obioptions.GenerateOptionParser(
		"obitagpcr",
		"split a paired raw read data set per sample",
		obitagpcr.OptionSet)

	optionParser(os.Args)
	pairs, err := obipairing.CLIPairedSequence()

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	paired := obitagpcr.IPCRTagPESequencesBatch(pairs,
		obipairing.CLIGapPenality(),
		obipairing.CLIPenalityScale(),
		obipairing.CLIDelta(),
		obipairing.CLIMinOverlap(),
		obipairing.CLIMinIdentity(),
		obipairing.CLIFastMode(),
		obipairing.CLIFastRelativeScore(),
		obipairing.CLIWithStats())

	obiconvert.CLIWriteBioSequences(paired, true)

	obiutils.WaitForLastPipe()
}
