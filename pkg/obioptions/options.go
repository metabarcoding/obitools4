package obioptions

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/DavidGamba/go-getoptions"
)

var _Debug = false
var _ParallelWorkers = runtime.NumCPU() - 1
var _MaxAllowedCPU = runtime.NumCPU()
var _BufferSize = 1
var _BatchSize = 5000

type ArgumentParser func([]string) (*getoptions.GetOpt, []string, error)

func GenerateOptionParser(optionset ...func(*getoptions.GetOpt)) ArgumentParser {
	options := getoptions.New()
	options.Bool("help", false, options.Alias("h", "?"))
	options.BoolVar(&_Debug, "debug", false)

	options.IntVar(&_ParallelWorkers, "workers", _ParallelWorkers,
		options.Alias("w"),
		options.Description("Number of parallele threads computing the result"))

	options.IntVar(&_MaxAllowedCPU, "max-cpu", _MaxAllowedCPU,
		options.Description("Number of parallele threads computing the result"))

	for _, o := range optionset {
		o(options)
	}

	return func(args []string) (*getoptions.GetOpt, []string, error) {

		remaining, err := options.Parse(args[1:])

		// Setup the maximum number of CPU usable by the program
		runtime.GOMAXPROCS(_MaxAllowedCPU)
		if options.Called("max-cpu") {
			log.Printf("CPU number limited to %d", _MaxAllowedCPU)
		}

		if options.Called("no-singleton") {
			log.Printf("No singleton option set")
		}

		if options.Called("help") {
			fmt.Fprint(os.Stderr, options.Help())
			os.Exit(1)
		}

		log.SetLevel(log.InfoLevel)
		if options.Called("debug") {
			log.SetLevel(log.DebugLevel)
			log.Debugln("Switch to debug level logging")
		}

		return options, remaining, err
	}
}

// Predicate indicating if the debug mode is activated.
func CLIIsDebugMode() bool {
	return _Debug
}

// CLIParallelWorkers returns the number of parallel workers requested by
// the command line option --workers|-w.
func CLIParallelWorkers() int {
	return _ParallelWorkers
}

// CLIParallelWorkers returns the number of parallel workers requested by
// the command line option --workers|-w.
func CLIMaxCPU() int {
	return _MaxAllowedCPU
}

// CLIBufferSize returns the expeted channel buffer size for obitools
func CLIBufferSize() int {
	return _BufferSize
}

// CLIBatchSize returns the expeted size of the sequence batches
func CLIBatchSize() int {
	return _BatchSize
}

// DebugOn sets the debug mode on.
func DebugOn() {
	_Debug = true
}

// DebugOff sets the debug mode off.
func DebugOff() {
	_Debug = false
}
