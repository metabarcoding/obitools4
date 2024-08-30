package main

import (
	"log"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obikmersim"
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

	optionParser := obioptions.GenerateOptionParser(obikmersim.CountOptionSet)

	_, args := optionParser(os.Args)

	var err error
	sequences := obiiter.NilIBioSequence

	if !obikmersim.CLISelf() {
		sequences, err = obiconvert.CLIReadBioSequences(args...)
	}

	obiconvert.OpenSequenceDataErrorMessage(args, err)

	counted := obikmersim.CLILookForSharedKmers(sequences)
	topull, err := obiconvert.CLIWriteBioSequences(counted, false)

	if err != nil {
		log.Panic(err)
	}

	topull.Consume()

	obiiter.WaitForLastPipe()

}
