package obioptions

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/DavidGamba/go-getoptions"

	"net/http"
	_ "net/http/pprof"
)

var _Debug = false
var _WorkerPerCore = 2.0
var _ReadWorkerPerCore = 1.0
var _MaxAllowedCPU = runtime.NumCPU()
var _BatchSize = 5000
var _Pprof = false

type ArgumentParser func([]string) (*getoptions.GetOpt, []string)

func GenerateOptionParser(optionset ...func(*getoptions.GetOpt)) ArgumentParser {

	options := getoptions.New()
	options.SetMode(getoptions.Bundling)
	options.SetUnknownMode(getoptions.Fail)
	options.Bool("help", false, options.Alias("h", "?"))
	options.BoolVar(&_Debug, "debug", false)
	options.BoolVar(&_Pprof, "pprof", false)

	// options.IntVar(&_ParallelWorkers, "workers", _ParallelWorkers,
	// 	options.Alias("w"),
	// 	options.Description("Number of parallele threads computing the result"))

	options.IntVar(&_MaxAllowedCPU, "max-cpu", _MaxAllowedCPU,
		options.GetEnv("OBIMAXCPU"),
		options.Description("Number of parallele threads computing the result"))

	for _, o := range optionset {
		o(options)
	}

	return func(args []string) (*getoptions.GetOpt, []string) {

		remaining, err := options.Parse(args[1:])

		if options.Called("help") {
			fmt.Fprint(os.Stderr, options.Help())
			os.Exit(1)
		}

		log.SetLevel(log.InfoLevel)
		if options.Called("debug") {
			log.SetLevel(log.DebugLevel)
			log.Debugln("Switch to debug level logging")
		}

		if options.Called("pprof") {
			go http.ListenAndServe("localhost:8080", nil)
			log.Infoln("Start a pprof server at address http://localhost:8080/debug/pprof")
		}

		// Handle user errors
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
			fmt.Fprint(os.Stderr, options.Help(getoptions.HelpSynopsis))
			os.Exit(1)
		}

		// Setup the maximum number of CPU usable by the program
		runtime.GOMAXPROCS(_MaxAllowedCPU)
		if options.Called("max-cpu") {
			log.Printf("CPU number limited to %d", _MaxAllowedCPU)
		}

		if options.Called("no-singleton") {
			log.Printf("No singleton option set")
		}

		log.Printf("Number of workers set %d", CLIParallelWorkers())
		return options, remaining
	}
}

// Predicate indicating if the debug mode is activated.
func CLIIsDebugMode() bool {
	return _Debug
}

// CLIParallelWorkers returns the number of parallel workers requested by
// the command line option --workers|-w.
func CLIParallelWorkers() int {
	return int(float64(_MaxAllowedCPU) * float64(_WorkerPerCore))
}

func CLIReadParallelWorkers() int {
	return int(float64(_MaxAllowedCPU) * float64(_ReadWorkerPerCore))
}


// CLIParallelWorkers returns the number of parallel workers requested by
// the command line option --workers|-w.
func CLIMaxCPU() int {
	return _MaxAllowedCPU
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

func SetWorkerPerCore(n float64) {
	_WorkerPerCore = n
}

func SetReadWorkerPerCore(n float64) {
	_ReadWorkerPerCore = n
}

func WorkerPerCore() float64 {
	return _WorkerPerCore
}

func ReadWorkerPerCore() float64 {
	return _ReadWorkerPerCore
}

func SetBatchSize(n int) {
	_BatchSize = n
}
