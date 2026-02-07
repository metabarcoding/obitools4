package obisuperkmer

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// Private variables for storing option values
var _KmerSize = 31
var _MinimizerSize = 13

// SuperKmerOptionSet defines every option related to super k-mer extraction.
//
// The function adds to a CLI every option proposed to the user
// to tune the parameters of the super k-mer extraction algorithm.
//
// Parameters:
// - options: is a pointer to a getoptions.GetOpt instance normally
//   produced by the obioptions.GenerateOptionParser function.
func SuperKmerOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&_KmerSize, "kmer-size", _KmerSize,
		options.Alias("k"),
		options.Description("Size of k-mers (must be between m+1 and 31)."))

	options.IntVar(&_MinimizerSize, "minimizer-size", _MinimizerSize,
		options.Alias("m"),
		options.Description("Size of minimizers (must be between 1 and k-1)."))
}

// OptionSet adds to the basic option set every option declared for
// the obisuperkmer command.
//
// It takes a pointer to a GetOpt struct as its parameter and does not return anything.
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(false)(options)
	SuperKmerOptionSet(options)
}

// CLIKmerSize returns the k-mer size to use for super k-mer extraction.
//
// It does not take any parameters.
// It returns an integer representing the k-mer size.
func CLIKmerSize() int {
	return _KmerSize
}

// SetKmerSize sets the k-mer size for super k-mer extraction.
//
// Parameters:
// - k: the k-mer size (must be between m+1 and 31).
func SetKmerSize(k int) {
	_KmerSize = k
}

// CLIMinimizerSize returns the minimizer size to use for super k-mer extraction.
//
// It does not take any parameters.
// It returns an integer representing the minimizer size.
func CLIMinimizerSize() int {
	return _MinimizerSize
}

// SetMinimizerSize sets the minimizer size for super k-mer extraction.
//
// Parameters:
// - m: the minimizer size (must be between 1 and k-1).
func SetMinimizerSize(m int) {
	_MinimizerSize = m
}
