package obikindex

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// Private variables for storing option values
var _kmerSize = 31
var _minimizerSize = -1 // -1 means auto: ceil(k / 2.5)
var _indexId = ""
var _metadataFormat = "toml"
var _setTag = make(map[string]string, 0)
var _minOccurrence = 1
var _saveFullFilter = false

// KmerIndexOptionSet defines every option related to kmer index building.
func KmerIndexOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&_kmerSize, "kmer-size", _kmerSize,
		options.Alias("k"),
		options.Description("Size of k-mers (must be between 2 and 31)."))

	options.IntVar(&_minimizerSize, "minimizer-size", _minimizerSize,
		options.Alias("m"),
		options.Description("Size of minimizers for parallelization (-1 for auto = ceil(k/2.5))."))

	options.StringVar(&_indexId, "index-id", _indexId,
		options.Description("Identifier for the kmer index."))

	options.StringVar(&_metadataFormat, "metadata-format", _metadataFormat,
		options.Description("Format for metadata file (toml, yaml, json)."))

	options.StringMapVar(&_setTag, "set-tag", 1, 1,
		options.Alias("S"),
		options.ArgName("KEY=VALUE"),
		options.Description("Adds a metadata attribute KEY with value VALUE to the index."))

	options.IntVar(&_minOccurrence, "min-occurrence", _minOccurrence,
		options.Description("Minimum number of occurrences for a k-mer to be kept (default 1 = keep all)."))

	options.BoolVar(&_saveFullFilter, "save-full-filter", _saveFullFilter,
		options.Description("When using --min-occurrence > 1, save the full frequency filter instead of just the filtered index."))
}

// OptionSet adds to the basic option set every option declared for
// the obikindex command.
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputModeOptionSet(options, false)
	KmerIndexOptionSet(options)
}

// CLIKmerSize returns the k-mer size.
func CLIKmerSize() int {
	return _kmerSize
}

// CLIMinimizerSize returns the effective minimizer size.
// If -1 (auto), computes ceil(k / 2.5) then applies constraints:
//   - minimum: ceil(log(nworkers) / log(4))
//   - maximum: k - 1
func CLIMinimizerSize() int {
	m := _minimizerSize
	if m < 0 {
		m = obikmer.DefaultMinimizerSize(_kmerSize)
	}
	nworkers := obidefault.ParallelWorkers()
	m = obikmer.ValidateMinimizerSize(m, _kmerSize, nworkers)
	return m
}

// CLIIndexId returns the index identifier.
func CLIIndexId() string {
	return _indexId
}

// CLIMetadataFormat returns the metadata format.
func CLIMetadataFormat() obikmer.MetadataFormat {
	switch strings.ToLower(_metadataFormat) {
	case "toml":
		return obikmer.FormatTOML
	case "yaml":
		return obikmer.FormatYAML
	case "json":
		return obikmer.FormatJSON
	default:
		log.Warnf("Unknown metadata format %q, defaulting to TOML", _metadataFormat)
		return obikmer.FormatTOML
	}
}

// CLISetTag returns the metadata key=value pairs.
func CLISetTag() map[string]string {
	return _setTag
}

// CLIMinOccurrence returns the minimum occurrence threshold.
func CLIMinOccurrence() int {
	return _minOccurrence
}

// CLISaveFullFilter returns whether to save the full frequency filter.
func CLISaveFullFilter() bool {
	return _saveFullFilter
}

// CLIOutputDirectory returns the output directory path.
func CLIOutputDirectory() string {
	return obiconvert.CLIOutPutFileName()
}

// SetKmerSize sets the k-mer size (for testing).
func SetKmerSize(k int) {
	_kmerSize = k
}

// SetMinimizerSize sets the minimizer size (for testing).
func SetMinimizerSize(m int) {
	_minimizerSize = m
}

// SetMinOccurrence sets the minimum occurrence (for testing).
func SetMinOccurrence(n int) {
	_minOccurrence = n
}
