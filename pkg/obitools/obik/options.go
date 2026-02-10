package obik

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// MaskingMode defines how to handle low-complexity regions
type MaskingMode int

const (
	MaskMode    MaskingMode = iota // Replace low-complexity regions with masked characters
	SplitMode                      // Split sequence into high-complexity fragments
	ExtractMode                    // Extract low-complexity fragments
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
// Shared kmer options (used by index, super, lowmask)
// ==============================

var _kmerSize = 31
var _minimizerSize = -1 // -1 means auto: ceil(k / 2.5)

// KmerSizeOptionSet registers --kmer-size / -k.
// Shared by index, super, and lowmask subcommands.
func KmerSizeOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&_kmerSize, "kmer-size", _kmerSize,
		options.Alias("k"),
		options.Description("Size of k-mers (must be between 2 and 31)."))
}

// MinimizerOptionSet registers --minimizer-size / -m.
// Shared by index and super subcommands.
func MinimizerOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&_minimizerSize, "minimizer-size", _minimizerSize,
		options.Alias("m"),
		options.Description("Size of minimizers for parallelization (-1 for auto = ceil(k/2.5))."))
}

// ==============================
// Lowmask-specific options
// ==============================

var _entropySize = 6
var _entropyThreshold = 0.5
var _splitMode = false
var _extractMode = false
var _maskingChar = "."
var _keepShorter = false

// LowMaskOptionSet registers options specific to low-complexity masking.
func LowMaskOptionSet(options *getoptions.GetOpt) {
	KmerSizeOptionSet(options)

	options.IntVar(&_entropySize, "entropy-size", _entropySize,
		options.Description("Maximum word size considered for entropy estimate."))

	options.Float64Var(&_entropyThreshold, "threshold", _entropyThreshold,
		options.Description("Entropy threshold below which a kmer is masked (0 to 1)."))

	options.BoolVar(&_splitMode, "extract-high", _splitMode,
		options.Description("Extract only high-complexity regions."))

	options.BoolVar(&_extractMode, "extract-low", _extractMode,
		options.Description("Extract only low-complexity regions."))

	options.StringVar(&_maskingChar, "masking-char", _maskingChar,
		options.Description("Character used to mask low complexity regions."))

	options.BoolVar(&_keepShorter, "keep-shorter", _keepShorter,
		options.Description("Keep fragments shorter than kmer-size in split/extract mode."))
}

// ==============================
// Index-specific options
// ==============================

var _indexId = ""
var _metadataFormat = "toml"
var _setTag = make(map[string]string, 0)
var _minOccurrence = 1
var _maxOccurrence = 0
var _saveFullFilter = false
var _saveFreqKmer = 0
var _indexEntropyThreshold = 0.0
var _indexEntropySize = 6

// KmerIndexOptionSet defines every option related to kmer index building.
func KmerIndexOptionSet(options *getoptions.GetOpt) {
	KmerSizeOptionSet(options)
	MinimizerOptionSet(options)

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

	options.Float64Var(&_indexEntropyThreshold, "entropy-filter", _indexEntropyThreshold,
		options.Description("Filter low-complexity k-mers with entropy <= threshold (0 = disabled)."))

	options.IntVar(&_indexEntropySize, "entropy-filter-size", _indexEntropySize,
		options.Description("Maximum word size for entropy filter computation (default 6)."))
}

// EntropyFilterOptionSet registers entropy filter options for commands
// that process existing indices (e.g. filter).
func EntropyFilterOptionSet(options *getoptions.GetOpt) {
	options.Float64Var(&_indexEntropyThreshold, "entropy-filter", _indexEntropyThreshold,
		options.Description("Filter low-complexity k-mers with entropy <= threshold (0 = disabled)."))

	options.IntVar(&_indexEntropySize, "entropy-filter-size", _indexEntropySize,
		options.Description("Maximum word size for entropy filter computation (default 6)."))
}

// ==============================
// Super kmer options
// ==============================

// SuperKmerOptionSet registers options specific to super k-mer extraction.
func SuperKmerOptionSet(options *getoptions.GetOpt) {
	KmerSizeOptionSet(options)
	MinimizerOptionSet(options)
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

// CLIMaskingMode returns the masking mode from CLI flags.
func CLIMaskingMode() MaskingMode {
	switch {
	case _extractMode:
		return ExtractMode
	case _splitMode:
		return SplitMode
	default:
		return MaskMode
	}
}

// CLIMaskingChar returns the masking character, validated.
func CLIMaskingChar() byte {
	mask := strings.TrimSpace(_maskingChar)
	if len(mask) != 1 {
		log.Fatalf("--masking-char option accepts a single character, not %s", mask)
	}
	return []byte(mask)[0]
}

// CLIEntropySize returns the entropy word size.
func CLIEntropySize() int {
	return _entropySize
}

// CLIEntropyThreshold returns the entropy threshold.
func CLIEntropyThreshold() float64 {
	return _entropyThreshold
}

// CLIKeepShorter returns whether to keep short fragments.
func CLIKeepShorter() bool {
	return _keepShorter
}

// ==============================
// Match-specific options
// ==============================

var _indexDirectory = ""

// IndexDirectoryOptionSet registers --index / -i (mandatory directory for match).
func IndexDirectoryOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_indexDirectory, "index", _indexDirectory,
		options.Alias("i"),
		options.Required(),
		options.ArgName("DIRECTORY"),
		options.Description("Path to the kmer index directory."))
}

// CLIIndexDirectory returns the --index directory path.
func CLIIndexDirectory() string {
	return _indexDirectory
}

// CLIIndexEntropyThreshold returns the entropy filter threshold for index building (0 = disabled).
func CLIIndexEntropyThreshold() float64 {
	return _indexEntropyThreshold
}

// CLIIndexEntropySize returns the entropy filter word size for index building.
func CLIIndexEntropySize() int {
	return _indexEntropySize
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
