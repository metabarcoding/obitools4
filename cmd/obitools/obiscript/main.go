package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiscript"
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

	optionParser := obioptions.GenerateOptionParser(obiscript.OptionSet)

	_, args := optionParser(os.Args)

	if obiscript.CLIAskScriptTemplate() {
		fmt.Print(obiscript.CLIScriptTemplate())
		os.Exit(0)
	}

	sequences, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	annotator := obiscript.CLIScriptPipeline()
	obiconvert.CLIWriteBioSequences(sequences.Pipe(annotator), true)

	obiiter.WaitForLastPipe()

}