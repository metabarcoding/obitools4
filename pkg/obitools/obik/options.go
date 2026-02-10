package obik

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// Output format flags
var _jsonOutput bool
var _csvOutput bool
var _yamlOutput bool

// Set selection flags
var _setPatterns []string

// Force flag
var _force bool

// Jaccard flag
var _jaccard bool

// Per-set tags for index subcommand
var _setMetaTags = make(map[string]string, 0)

// ==============================
// Kmer index building options (moved from obikindex)
// ==============================

var _kmerSize = 31
var _minimizerSize = -1 // -1 means auto: ceil(k / 2.5)
var _indexId = ""
var _metadataFormat = "toml"
var _setTag = make(map[string]string, 0)
var _minOccurrence = 1
var _maxOccurrence = 0
var _saveFullFilter = false
var _saveFreqKmer = 0

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
		options.Description("Adds a group-level metadata attribute KEY with value VALUE."))

	options.IntVar(&_minOccurrence, "min-occurrence", _minOccurrence,
		options.Description("Minimum number of occurrences for a k-mer to be kept (default 1 = keep all)."))

	options.IntVar(&_maxOccurrence, "max-occurrence", _maxOccurrence,
		options.Description("Maximum number of occurrences for a k-mer to be kept (default 0 = no upper bound)."))

	options.BoolVar(&_saveFullFilter, "save-full-filter", _saveFullFilter,
		options.Description("When using --min-occurrence > 1, save the full frequency filter instead of just the filtered index."))

	options.IntVar(&_saveFreqKmer, "save-freq-kmer", _saveFreqKmer,
		options.Description("Save the N most frequent k-mers per set to a CSV file (top_kmers.csv)."))
}

// CLIKmerSize returns the k-mer size.
func CLIKmerSize() int {
	return _kmerSize
}

// CLIMinimizerSize returns the effective minimizer size.
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

// CLISetTag returns the group-level metadata key=value pairs.
func CLISetTag() map[string]string {
	return _setTag
}

// CLIMinOccurrence returns the minimum occurrence threshold.
func CLIMinOccurrence() int {
	return _minOccurrence
}

// CLIMaxOccurrence returns the maximum occurrence threshold (0 = no upper bound).
func CLIMaxOccurrence() int {
	return _maxOccurrence
}

// CLISaveFullFilter returns whether to save the full frequency filter.
func CLISaveFullFilter() bool {
	return _saveFullFilter
}

// CLISaveFreqKmer returns the number of top frequent k-mers to save (0 = disabled).
func CLISaveFreqKmer() int {
	return _saveFreqKmer
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

// OutputFormatOptionSet registers --json-output, --csv-output, --yaml-output.
func OutputFormatOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&_jsonOutput, "json-output", false,
		options.Description("Print results as JSON."))
	options.BoolVar(&_csvOutput, "csv-output", false,
		options.Description("Print results as CSV."))
	options.BoolVar(&_yamlOutput, "yaml-output", false,
		options.Description("Print results as YAML."))
}

// CLIOutFormat returns the selected output format: "json", "csv", "yaml", or "text".
func CLIOutFormat() string {
	if _jsonOutput {
		return "json"
	}
	if _csvOutput {
		return "csv"
	}
	if _yamlOutput {
		return "yaml"
	}
	return "text"
}

// SetSelectionOptionSet registers --set <glob_pattern> (repeatable).
func SetSelectionOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&_setPatterns, "set", 1, 1,
		options.Alias("s"),
		options.ArgName("PATTERN"),
		options.Description("Set ID or glob pattern (repeatable, supports *, ?, [...])."))
}

// CLISetPatterns returns the --set patterns provided by the user.
func CLISetPatterns() []string {
	return _setPatterns
}

// ForceOptionSet registers --force / -f.
func ForceOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&_force, "force", false,
		options.Alias("f"),
		options.Description("Force operation even if set ID already exists in destination."))
}

// CLIForce returns whether --force was specified.
func CLIForce() bool {
	return _force
}
