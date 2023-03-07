package main

import (
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obitag"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
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

	optionParser := obioptions.GenerateOptionParser(obitag.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)",err)
		os.Exit(1)
	}

	identified := obitag.AssignTaxonomy(fs)

	obiconvert.CLIWriteBioSequences(identified, true)
	obiiter.WaitForLastPipe()

	fmt.Println("")
}
