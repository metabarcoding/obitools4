package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obipcr"
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

	obioptions.SetWorkerPerCore(2)
	obioptions.SetReadWorkerPerCore(0.5)

	optionParser := obioptions.GenerateOptionParser(obipcr.OptionSet)

	_, args := optionParser(os.Args)

	sequences, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	amplicons, _ := obipcr.CLIPCR(sequences)
	obiconvert.CLIWriteBioSequences(amplicons, true)
	obiiter.WaitForLastPipe()

}
