package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obipcr"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {

	// go tool pprof -nodefraction=0 -http=:8081 http://localhost:8080/debug/pprof/allocs
	// look at http://localhost:8080/debug/pprof for havng the possibilities
	//go http.ListenAndServe("localhost:8080", nil)

	// go tool trace cpu.trace
	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	obidefault.SetWorkerPerCore(2)
	obidefault.SetReadWorkerPerCore(0.5)
	obidefault.SetParallelFilesRead(obidefault.ParallelWorkers() / 4)
	obidefault.SetBatchSize(10)

	optionParser := obioptions.GenerateOptionParser(
		"obipcr",
		"simulates a PCR on a sequence files",
		obipcr.OptionSet)

	_, args := optionParser(os.Args)

	sequences, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	amplicons, _ := obipcr.CLIPCR(sequences)
	obiconvert.CLIWriteBioSequences(amplicons, true)
	obiutils.WaitForLastPipe()

}
