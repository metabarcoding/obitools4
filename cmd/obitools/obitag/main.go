package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obitag"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

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

	obidefault.SetWorkerPerCore(2)
	obidefault.SetStrictReadWorker(1)
	obidefault.SetStrictWriteWorker(1)
	obidefault.SetBatchSize(10)

	optionParser := obioptions.GenerateOptionParser(
		"obitag",
		"realizes taxonomic assignment",
		obitag.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	taxo := obitax.DefaultTaxonomy()

	references := obitag.CLIRefDB()

	if references == nil {
		log.Panicln("No loaded reference database")
	}

	if taxo == nil {
		taxo, err = references.ExtractTaxonomy(nil)

		if err != nil {
			log.Fatalf("No taxonomy specified or extractable from reference database: %v", err)
		}

		taxo.SetAsDefault()
	}

	if taxo == nil {
		log.Panicln("No loaded taxonomy")
	}

	var identified obiiter.IBioSequence

	if obitag.CLIGeometricMode() {
		identified = obitag.CLIGeomAssignTaxonomy(fs, references, taxo)
	} else {
		identified = obitag.CLIAssignTaxonomy(fs, references, taxo)
	}

	obiconvert.CLIWriteBioSequences(identified, true)
	obiutils.WaitForLastPipe()

	obitag.CLISaveRefetenceDB(references)

	fmt.Println("")
}
