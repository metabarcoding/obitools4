package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obifind"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obitag"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obitag2"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func main() {

	// go tool pprof -http=":8000" ./build/obitag ./cpu.pprof
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

	obioptions.SetWorkerPerCore(2)
	obioptions.SetStrictReadWorker(1)
	obioptions.SetStrictWriteWorker(1)
	obioptions.SetBatchSize(10)

	optionParser := obioptions.GenerateOptionParser(obitag.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	taxo, error := obifind.CLILoadSelectedTaxonomy()
	if error != nil {
		log.Panicln(error)
	}

	references := obitag.CLIRefDB()

	identified := obitag2.CLIAssignTaxonomy(fs, references, taxo)

	obiconvert.CLIWriteBioSequences(identified, true)
	obiiter.WaitForLastPipe()

	fmt.Println("")
}