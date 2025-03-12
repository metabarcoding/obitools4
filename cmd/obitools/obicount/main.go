package main

import (
	"fmt"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicount"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
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

	optionParser := obioptions.GenerateOptionParser(
		"obicount",
		"counts the sequences present in a file of sequences",
		obiconvert.InputOptionSet,
		obicount.OptionSet,
	)

	_, args := optionParser(os.Args)

	obidefault.SetStrictReadWorker(min(4, obidefault.ParallelWorkers()))
	fs, err := obiconvert.CLIReadBioSequences(args...)
	obiconvert.OpenSequenceDataErrorMessage(args, err)

	nvariant, nread, nsymbol := fs.Count(true)

	fmt.Print("entities,n\n")

	if obicount.CLIIsPrintingVariantCount() {
		fmt.Printf("variants,%d\n", nvariant)
	}

	if obicount.CLIIsPrintingReadCount() {
		fmt.Printf("reads,%d\n", nread)
	}

	if obicount.CLIIsPrintingSymbolCount() {
		fmt.Printf("symbols,%d\n", nsymbol)
	}

}
