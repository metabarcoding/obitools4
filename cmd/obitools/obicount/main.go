package main

import (
	"fmt"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiblackboard"
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

	black := obiblackboard.NewBlackBoard(obioptions.CLIParallelWorkers())

	black.ReadSequences(args)

	counter := obiblackboard.CountSequenceAggregator("to_delete")

	black.RegisterRunner("sequences", counter.Runner)
	black.RegisterRunner("to_delete", obiblackboard.RecycleSequences(true, "final"))

	black.Run()

	fmt.Print("entity,n\n")

	if obicount.CLIIsPrintingVariantCount() {
		fmt.Printf("variants,%d\n", counter.Variants)
	}

	if obicount.CLIIsPrintingReadCount() {
		fmt.Printf("reads,%d\n", counter.Reads)
	}

	if obicount.CLIIsPrintingSymbolCount() {
		fmt.Printf("nucleotides,%d\n", counter.Nucleotides)
	}
}
