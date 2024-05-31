package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obimultiplex"
)

func main() {

	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	optionParser := obioptions.GenerateOptionParser(obimultiplex.OptionSet)

	_, args := optionParser(os.Args)

	if obimultiplex.CLIAskConfigTemplate() {
		fmt.Print(obimultiplex.CLIConfigTemplate())
		os.Exit(0)
	}

	if !obimultiplex.CLIHasNGSFilterFile() {
		log.Error("You must provide a tag list file following the NGSFilter format")
		os.Exit(1)
	}

	sequences, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}
	amplicons, _ := obimultiplex.IExtractBarcode(sequences)
	obiconvert.CLIWriteBioSequences(amplicons, true)
	amplicons.Wait()
	obiiter.WaitForLastPipe()

}
