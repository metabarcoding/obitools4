package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

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
		obiconvert.InputOptionSet,
		obicount.OptionSet,
	)

	_, args := optionParser(os.Args)

	fs, err := obiconvert.CLIReadBioSequences(args...)

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	nvariant, nread, nsymbol := fs.Count(true)

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
