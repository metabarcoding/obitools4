package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"runtime/pprof"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obimultiplex"
)

func main() {

	f, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// ftrace, err := os.Create("cpu.trace")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// trace.Start(ftrace)
	// defer trace.Stop()

	optionParser := obioptions.GenerateOptionParser(obimultiplex.OptionSet)

	_, args, _ := optionParser(os.Args)

	sequences, _ := obiconvert.ReadBioSequencesBatch(args...)
	amplicons, _ := obimultiplex.IExtractBarcodeBatches(sequences)
	obiconvert.WriteBioSequencesBatch(amplicons, true)
	amplicons.Wait()
}
