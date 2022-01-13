package obioptions

import (
	"fmt"
	"os"

	"github.com/DavidGamba/go-getoptions"
)

var __debug__ = false
var __profiling__ = ""

type ArgumentParser func([]string) (*getoptions.GetOpt, []string, error)

func GenerateOptionParser(optionset ...func(*getoptions.GetOpt)) ArgumentParser {
	options := getoptions.New()
	options.Bool("help", false, options.Alias("h", "?"))
	options.BoolVar(&__debug__, "debug", false)
	// options.StringVar(&__profiling__, "profile", "")

	for _, o := range optionset {
		o(options)
	}

	return func(args []string) (*getoptions.GetOpt, []string, error) {

		remaining, err := options.Parse(args[1:])

		if options.Called("help") {
			fmt.Fprintf(os.Stderr, options.Help())
			os.Exit(1)
		}
		return options, remaining, err
	}
}

// Predicate indicating if the debug mode is activated
func IsDebugMode() bool {
	return __debug__
}

func DebugOn() {
	__debug__ = true
}

func DebugOff() {
	__debug__ = false
}
