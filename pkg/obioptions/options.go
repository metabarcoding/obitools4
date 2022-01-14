package obioptions

import (
	"fmt"
	"os"
	"runtime"

	"github.com/DavidGamba/go-getoptions"
)

var _Debug = false
var _ParallelWorkers = runtime.NumCPU() - 1
var _BufferSize = 1
var _BatchSize = 5000

type ArgumentParser func([]string) (*getoptions.GetOpt, []string, error)

func GenerateOptionParser(optionset ...func(*getoptions.GetOpt)) ArgumentParser {
	options := getoptions.New()
	options.Bool("help", false, options.Alias("h", "?"))
	options.BoolVar(&_Debug, "debug", false)

	options.IntVar(&_ParallelWorkers, "workers", runtime.NumCPU()-1,
		options.Alias("w"),
		options.Description("Number of parallele threads computing the result"))

	for _, o := range optionset {
		o(options)
	}

	return func(args []string) (*getoptions.GetOpt, []string, error) {

		remaining, err := options.Parse(args[1:])

		if options.Called("help") {
			fmt.Fprint(os.Stderr, options.Help())
			os.Exit(1)
		}
		return options, remaining, err
	}
}

// Predicate indicating if the debug mode is activated.
func IsDebugMode() bool {
	return _Debug
}

// ParallelWorkers returns the number of parallel workers requested by
// the command line option --workers|-w.
func ParallelWorkers() int {
	return _ParallelWorkers
}

// BufferSize returns the expeted channel buffer size for obitools
func BufferSize() int {
	return _BufferSize
}

// BatchSize returns the expeted size of the sequence batches
func BatchSize() int {
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
