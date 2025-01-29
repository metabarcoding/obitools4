package obioptions

import (
	"fmt"
	"os"
	"runtime"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	log "github.com/sirupsen/logrus"

	"github.com/DavidGamba/go-getoptions"

	"net/http"
	_ "net/http/pprof"
)

var _Debug = false
var _BatchSize = 2000
var _Pprof = false
var _PprofMudex = 10
var _PprofGoroutine = 6060

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

	options.IntVar(obidefault.MaxCPUPtr(), "max-cpu", obidefault.MaxCPU(),
		options.GetEnv("OBIMAXCPU"),
		options.Description("Number of parallele threads computing the result"))

	options.BoolVar(&_Pprof, "force-one-cpu", false,
		options.Description("Force to use only one cpu core for parallel processing"))

	options.IntVar(&_PprofMudex, "pprof-mutex", _PprofMudex,
		options.GetEnv("OBIPPROFMUTEX"),
		options.Description("Enable profiling of mutex lock."))

	options.IntVar(&_PprofGoroutine, "pprof-goroutine", _PprofGoroutine,
		options.GetEnv("OBIPPROFGOROUTINE"),
		options.Description("Enable profiling of goroutine blocking profile."))

	options.IntVar(obidefault.BatchSizePtr(), "batch-size", obidefault.BatchSize(),
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
			fmt.Fprintf(os.Stderr, "OBITools %s\n", VersionString())
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

		if options.Called("pprof-mutex") {
			url := "localhost:6060"
			go http.ListenAndServe(url, nil)
			runtime.SetMutexProfileFraction(_PprofMudex)
			log.Infof("Start a pprof server at address %s/debug/pprof", url)
			log.Info("Profil can be followed running concurrently the command :")
			log.Info("  go tool pprof -http=127.0.0.1:8080 'http://localhost:6060/debug/pprof/mutex'")
		}

		if options.Called("pprof-goroutine") {
			url := "localhost:6060"
			go http.ListenAndServe(url, nil)
			runtime.SetBlockProfileRate(_PprofGoroutine)
			log.Infof("Start a pprof server at address %s/debug/pprof", url)
			log.Info("Profil can be followed running concurrently the command :")
			log.Info("  go tool pprof -http=127.0.0.1:8080 'http://localhost:6060/debug/pprof/block'")
		}

		// Handle user errors
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
			fmt.Fprint(os.Stderr, options.Help(getoptions.HelpSynopsis))
			os.Exit(1)
		}

		// Setup the maximum number of CPU usable by the program
		if obidefault.MaxCPU() == 1 {
			log.Warn("Limitating the Maximum number of CPU to 1 is not recommanded")
			log.Warn("The number of CPU requested has been set to 2")
			obidefault.SetMaxCPU(2)
		}

		if options.Called("force-one-cpu") {
			log.Warn("Limitating the Maximum number of CPU to 1 is not recommanded")
			log.Warn("The number of CPU has been forced to 1")
			log.Warn("This can lead to unexpected behavior")
			obidefault.SetMaxCPU(1)
		}

		runtime.GOMAXPROCS(obidefault.MaxCPU())

		if options.Called("max-cpu") || options.Called("force-one-cpu") {
			log.Printf("CPU number limited to %d", obidefault.MaxCPU())
		}

		if options.Called("no-singleton") {
			log.Printf("No singleton option set")
		}

		log.Printf("Number of workers set %d", obidefault.ParallelWorkers())

		// if options.Called("workers") {

		// }

		if options.Called("solexa") {
			obidefault.SetReadQualitiesShift(64)
		}

		return options, remaining
	}
}

func LoadTaxonomyOptionSet(options *getoptions.GetOpt, required, alternatiive bool) {
	if required {
		options.StringVar(obidefault.SelectedTaxonomyPtr(), "taxonomy", obidefault.SelectedTaxonomy(),
			options.Alias("t"),
			options.Required(),
			options.Description("Path to the taxonomy database."))
	} else {
		options.StringVar(obidefault.SelectedTaxonomyPtr(), "taxonomy", obidefault.SelectedTaxonomy(),
			options.Alias("t"),
			options.Description("Path to the taxonomy database."))
	}
	if alternatiive {
		options.BoolVar(obidefault.AlternativeNamesSelectedPtr(), "alternative-names", obidefault.AreAlternativeNamesSelected(),
			options.Alias("a"),
			options.Description("Enable the search on all alternative names and not only scientific names."))
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

// SetDebugOn sets the debug mode on.
func SetDebugOn() {
	_Debug = true
}

// SetDebugOff sets the debug mode off.
func SetDebugOff() {
	_Debug = false
}
