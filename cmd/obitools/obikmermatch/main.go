package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obikmersim"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
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

	optionParser := obioptions.GenerateOptionParser(
		"obikmermatch",
		"",
		obikmersim.MatchOptionSet)

	_, args := optionParser(os.Args)

	var err error
	sequences := obiiter.NilIBioSequence

	if !obikmersim.CLISelf() {
		sequences, err = obiconvert.CLIReadBioSequences(args...)
	}

	obiconvert.OpenSequenceDataErrorMessage(args, err)

	selected := obikmersim.CLIAlignSequences(sequences)
	obiconvert.CLIWriteBioSequences(selected, true)
	obiutils.WaitForLastPipe()

}
