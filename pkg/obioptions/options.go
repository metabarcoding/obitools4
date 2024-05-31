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
var _StrictReadWorker = 0
var _ParallelFilesRead = 0
var _MaxAllowedCPU = runtime.NumCPU()
var _BatchSize = 5000
var _Pprof = false
var _Quality_Shift_Input = byte(33)
var _Quality_Shift_Output = byte(33)
var _Version = "4.2.1"

type ArgumentParser func([]string) (*getoptions.GetOpt, []string)

func GenerateOptionParser(optionset ...func(*getoptions.GetOpt)) ArgumentParser {

	options := getoptions.New()
	options.SetMode(getoptions.Bundling)
	options.SetUnknownMode(getoptions.Fail)
	options.Bool("help", false, options.Alias("h", "?"))

	options.Bool("version", false,
		options.Description("Prints the version and exits."))

	options.BoolVar(&_Debug, "debug", false,
		options.GetEnv("OBIDEBUG"),
		options.Description("Enable debug mode, by setting log level to debug."))

	options.BoolVar(&_Pprof, "pprof", false,
		options.Description("Enable pprof server. Look at the log for details."))

	// options.IntVar(&_ParallelWorkers, "workers", _ParallelWorkers,
	// 	options.Alias("w"),
	// 	options.Description("Number of parallele threads computing the result"))

	options.IntVar(&_MaxAllowedCPU, "max-cpu", _MaxAllowedCPU,
		options.GetEnv("OBIMAXCPU"),
		options.Description("Number of parallele threads computing the result"))

	options.BoolVar(&_Pprof, "force-one-cpu", false,
		options.Description("Force to use only one cpu core for parallel processing"))

	options.IntVar(&_BatchSize, "batch-size", _BatchSize,
		options.GetEnv("OBIBATCHSIZE"),
		options.Description("Number of sequence per batch for paralelle processing"))

	options.Bool("solexa", false,
		options.GetEnv("OBISOLEXA"),
		options.Description("Decodes quality string according to the Solexa specification."))

	for _, o := range optionset {
		o(options)
	}

	return func(args []string) (*getoptions.GetOpt, []string) {

		remaining, err := options.Parse(args[1:])

		if options.Called("help") {
			fmt.Fprint(os.Stderr, options.Help())
			os.Exit(1)
		}

		if options.Called("version") {
			fmt.Fprintf(os.Stderr, "obitools version %s\n", _Version)
			os.Exit(0)
		}

		log.SetLevel(log.InfoLevel)
		if options.Called("debug") {
			log.SetLevel(log.DebugLevel)
			log.Debugln("Switch to debug level logging")
		}

		if options.Called("pprof") {
			url := "localhost:6060"
			go http.ListenAndServe(url, nil)
			log.Infof("Start a pprof server at address %s/debug/pprof", url)
			log.Info("Profil can be followed running concurrently the command :")
			log.Info("  go tool pprof -http=127.0.0.1:8080 'http://localhost:6060/debug/pprof/profile?seconds=30'")
		}

		// Handle user errors
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
			fmt.Fprint(os.Stderr, options.Help(getoptions.HelpSynopsis))
			os.Exit(1)
		}

		// Setup the maximum number of CPU usable by the program
		if _MaxAllowedCPU == 1 {
			log.Warn("Limitating the Maximum number of CPU to 1 is not recommanded")
			log.Warn("The number of CPU requested has been set to 2")
			SetMaxCPU(2)
		}

		if options.Called("force-one-cpu") {
			log.Warn("Limitating the Maximum number of CPU to 1 is not recommanded")
			log.Warn("The number of CPU has been forced to 1")
			log.Warn("This can lead to unexpected behavior")
			SetMaxCPU(1)
		}

		runtime.GOMAXPROCS(_MaxAllowedCPU)

		if options.Called("max-cpu") || options.Called("force-one-cpu") {
			log.Printf("CPU number limited to %d", _MaxAllowedCPU)
		}

		if options.Called("no-singleton") {
			log.Printf("No singleton option set")
		}

		log.Printf("Number of workers set %d", CLIParallelWorkers())

		// if options.Called("workers") {

		// }

		if options.Called("solexa") {
			SetInputQualityShift(64)
		}

		return options, remaining
	}
}

// CLIIsDebugMode returns whether the CLI is in debug mode.
//
// The debug mode is activated by the command line option --debug or
// the environment variable OBIDEBUG.
// It can be activated programmatically by the SetDebugOn() function.
//
// No parameters.
// Returns a boolean indicating if the CLI is in debug mode.
func CLIIsDebugMode() bool {
	return _Debug
}

// CLIParallelWorkers returns the number of parallel workers used for
// computing the result.
//
// The number of parallel workers is determined by the command line option
// --max-cpu|-m and the environment variable OBIMAXCPU. This number is
// multiplied by the variable _WorkerPerCore.
//
// No parameters.
// Returns an integer representing the number of parallel workers.
func CLIParallelWorkers() int {
	return int(float64(CLIMaxCPU()) * float64(WorkerPerCore()))
}

// CLIReadParallelWorkers returns the number of parallel workers used for
// reading files.
//
// The number of parallel workers is determined by the command line option
// --max-cpu|-m and the environment variable OBIMAXCPU. This number is
// multiplied by the variable _ReadWorkerPerCore.
//
// No parameters.
// Returns an integer representing the number of parallel workers.
func CLIReadParallelWorkers() int {
	if StrictReadWorker() == 0 {
		return int(float64(CLIMaxCPU()) * ReadWorkerPerCore())
	} else {
		return StrictReadWorker()
	}
}

// CLIMaxCPU returns the maximum number of CPU cores allowed.
//
// The maximum number of CPU cores is determined by the command line option
// --max-cpu|-m and the environment variable OBIMAXCPU.
//
// No parameters.
// Returns an integer representing the maximum number of CPU cores allowed.
func CLIMaxCPU() int {
	return _MaxAllowedCPU
}

// CLIBatchSize returns the expected size of the sequence batches.
//
// In Obitools, the sequences are processed in parallel by batches.
// The number of sequence in each batch is determined by the command line option
// --batch-size and the environment variable OBIBATCHSIZE.
//
// No parameters.
// Returns an integer value.
func CLIBatchSize() int {
	return _BatchSize
}

// SetDebugOn sets the debug mode on.
func SetDebugOn() {
	_Debug = true
}

// SetDebugOff sets the debug mode off.
func SetDebugOff() {
	_Debug = false
}

// SetWorkerPerCore sets the number of workers per CPU core.
//
// It takes a float64 parameter representing the number of workers
// per CPU core and does not return any value.
func SetWorkerPerCore(n float64) {
	_WorkerPerCore = n
}

// SetReadWorkerPerCore sets the number of worker per CPU
// core for reading files.
//
// n float64
func SetReadWorkerPerCore(n float64) {
	_ReadWorkerPerCore = n
}

// WorkerPerCore returns the number of workers per CPU core.
//
// No parameters.
// Returns a float64 representing the number of workers per CPU core.
func WorkerPerCore() float64 {
	return _WorkerPerCore
}

// ReadWorkerPerCore returns the number of worker per CPU core for
// computing the result.
//
// No parameters.
// Returns a float64 representing the number of worker per CPU core.
func ReadWorkerPerCore() float64 {
	return _ReadWorkerPerCore
}

// SetBatchSize sets the size of the sequence batches.
//
// n - an integer representing the size of the sequence batches.
func SetBatchSize(n int) {
	_BatchSize = n
}

// InputQualityShift returns the quality shift value for input.
//
// It can be set programmatically by the SetInputQualityShift() function.
// This value is used to decode the quality scores in FASTQ files.
// The quality shift value defaults to 33, which is the correct value for
// Sanger formated FASTQ files.
// The quality shift value can be modified to 64 by the command line option
// --solexa, for decoding old Solexa formated FASTQ files.
//
// No parameters.
// Returns an integer representing the quality shift value for input.
func InputQualityShift() byte {
	return _Quality_Shift_Input
}

// OutputQualityShift returns the quality shift value used for FASTQ output.
//
// No parameters.
// Returns an integer representing the quality shift value for output.
func OutputQualityShift() byte {
	return _Quality_Shift_Output
}

// SetInputQualityShift sets the quality shift value for decoding FASTQ.
//
// n - an integer representing the quality shift value to be set.
func SetInputQualityShift[T int | byte](n T) {
	_Quality_Shift_Input = byte(n)
}

// SetOutputQualityShift sets the quality shift value used for FASTQ output.
//
// n - an integer representing the quality shift value to be set.
func SetOutputQualityShift[T int | byte](n T) {
	_Quality_Shift_Output = byte(n)
}

// SetMaxCPU sets the maximum number of CPU cores allowed.
//
// n - an integer representing the new maximum number of CPU cores.
func SetMaxCPU(n int) {
	_MaxAllowedCPU = n
}

// SetReadWorker sets the number of workers for reading files.
//
// The number of worker dedicated to reading files is determined
// as the number of allowed CPU cores multiplied by number of read workers per core.
// Setting the number of read workers using this function allows to decouple the number
// of read workers from the number of CPU cores.
//
// n - an integer representing the number of workers to be set.
func SetStrictReadWorker(n int) {
	_StrictReadWorker = n
}

// ReadWorker returns the number of workers for reading files.
//
// No parameters.
// Returns an integer representing the number of workers.
func StrictReadWorker() int {
	return _StrictReadWorker
}

// ParallelFilesRead returns the number of files to be read in parallel.
//
// No parameters.
// Returns an integer representing the number of files to be read.
func ParallelFilesRead() int {
	if _ParallelFilesRead == 0 {
		return CLIParallelWorkers()
	} else {
		return _ParallelFilesRead
	}
}

// SetParallelFilesRead sets the number of files to be read in parallel.
//
// n - an integer representing the number of files to be set.
func SetParallelFilesRead(n int) {
	_ParallelFilesRead = n
}
