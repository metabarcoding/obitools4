package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obisummary"
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
		"obisummary",
		"resume main information from a sequence file",
		obisummary.OptionSet)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	summary := obisummary.ISummary(fs, obisummary.CLIMapSummary())

	if obisummary.CLIOutFormat() == "json" {
		output, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Print(string(output))
	} else {
		output, _ := yaml.Marshal(summary)
		fmt.Print(string(output))
	}
	fmt.Printf("\n")
}
