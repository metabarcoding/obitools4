package main

import (
	"fmt"
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obicount"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
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
		obiconvert.InputOptionSet,
		obicount.OptionSet,
	)

	_, args, _ := optionParser(os.Args)

	fs, _ := obiconvert.ReadBioSequencesBatch(args...)
	nread, nvariant, nsymbol := fs.Count(true)

	if obicount.CLIIsPrintingVariantCount() {
		fmt.Printf(" %d", nvariant)
	}

	if obicount.CLIIsPrintingReadCount() {
		fmt.Printf(" %d", nread)
	}

	if obicount.CLIIsPrintingSymbolCount() {
		fmt.Printf(" %d", nsymbol)
	}

	fmt.Printf("\n")
}
