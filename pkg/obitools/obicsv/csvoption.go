package obicsv

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
)

// __options__ holds configuration options for processing.
// Each field corresponds to a specific setting that can be adjusted.
type __options__ struct {
	with_progress_bar bool // Indicates whether to display a progress bar
	filename          string
	buffer_size       int      // Size of the buffer for processing
	batch_size        int      // Number of items to process in a batch
	full_file_batch   bool     // Indicates whether to process the full file in a batch
	parallel_workers  int      // Number of parallel workers to use
	no_order          bool     // Indicates whether to process items in no specific order
	closefile         bool     // Indicates whether to close the file after processing
	appendfile        bool     // Indicates whether to append to the file instead of overwriting
	compressed        bool     // Indicates whether the input data is compressed
	skip_empty        bool     // Indicates whether to skip empty entries
	csv_naomit        bool     // Indicates whether to omit NA values in CSV output
	csv_id            bool     // Indicates whether to include ID in CSV output
	csv_sequence      bool     // Indicates whether to include sequence in CSV output
	csv_quality       bool     // Indicates whether to include quality in CSV output
	csv_definition    bool     // Indicates whether to include definition in CSV output
	csv_count         bool     // Indicates whether to include count in CSV output
	csv_taxon         bool     // Indicates whether to include taxon in CSV output
	csv_keys          []string // List of keys to include in CSV output
	csv_separator     string   // Separator to use in CSV output
	csv_navalue       string   // Value to use for NA entries in CSV output
	csv_auto          bool     // Indicates whether to automatically determine CSV format
	source            string   // Source of the data
}

// Options wraps the __options__ struct to provide a pointer to the options.
type Options struct {
	pointer *__options__ // Pointer to the underlying options
}

// WithOption is a function type that takes an Options parameter and modifies it.
type WithOption func(Options)

// MakeOptions creates an Options instance with default settings and applies any provided setters.
// It returns the configured Options.
//
// Parameters:
//   - setters: A slice of WithOption functions to customize the options.
//
// Returns:
//   - An Options instance with the specified settings.
func MakeOptions(setters []WithOption) Options {
	o := __options__{
		with_progress_bar: false,
		filename:          "-",
		buffer_size:       2,
		parallel_workers:  obidefault.ReadParallelWorkers(),
		batch_size:        obidefault.BatchSize(),
		no_order:          false,
		full_file_batch:   false,
		closefile:         false,
		appendfile:        false,
		compressed:        false,
		skip_empty:        false,
		csv_id:            true,
		csv_definition:    false,
		csv_count:         false,
		csv_taxon:         false,
		csv_sequence:      true,
		csv_quality:       false,
		csv_separator:     ",",
		csv_navalue:       "NA",
		csv_keys:          make(obiiter.CSVHeader, 0),
		csv_auto:          false,
		source:            "unknown",
	}
	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

// BatchSize returns the size of the batch to be processed.
// It retrieves the batch size from the underlying options.
func (opt Options) BatchSize() int {
	return opt.pointer.batch_size
}

func (opt Options) FileName() string {
	return opt.pointer.filename
}

// FullFileBatch returns whether the full file should be processed in a single batch.
// It retrieves the setting from the underlying options.
func (opt Options) FullFileBatch() bool {
	return opt.pointer.full_file_batch
}

// ParallelWorkers returns the number of parallel workers to be used for processing.
// It retrieves the number of workers from the underlying options.
func (opt Options) ParallelWorkers() int {
	return opt.pointer.parallel_workers
}

// NoOrder returns whether the processing should occur in no specific order.
// It retrieves the setting from the underlying options.
func (opt Options) NoOrder() bool {
	return opt.pointer.no_order
}

// ProgressBar returns whether a progress bar should be displayed during processing.
// It retrieves the setting from the underlying options.
func (opt Options) ProgressBar() bool {
	return opt.pointer.with_progress_bar
}

// CloseFile returns whether the file should be closed after processing.
// It retrieves the setting from the underlying options.
func (opt Options) CloseFile() bool {
	return opt.pointer.closefile
}

// AppendFile returns whether to append to the file instead of overwriting it.
// It retrieves the setting from the underlying options.
func (opt Options) AppendFile() bool {
	return opt.pointer.appendfile
}

// CompressedFile returns whether the input data is compressed.
// It retrieves the setting from the underlying options.
func (opt Options) CompressedFile() bool {
	return opt.pointer.compressed
}

// SkipEmptySequence returns whether empty sequences should be skipped during processing.
// It retrieves the setting from the underlying options.
func (opt Options) SkipEmptySequence() bool {
	return opt.pointer.skip_empty
}

// CSVId returns whether the ID should be included in the CSV output.
// It retrieves the setting from the underlying options.
func (opt Options) CSVId() bool {
	return opt.pointer.csv_id
}

// CSVDefinition returns whether the definition should be included in the CSV output.
// It retrieves the setting from the underlying options.
func (opt Options) CSVDefinition() bool {
	return opt.pointer.csv_definition
}

// CSVCount returns whether the count should be included in the CSV output.
// It retrieves the setting from the underlying options.
func (opt Options) CSVCount() bool {
	return opt.pointer.csv_count
}

// CSVTaxon returns whether the taxon should be included in the CSV output.
// It retrieves the setting from the underlying options.
func (opt Options) CSVTaxon() bool {
	return opt.pointer.csv_taxon
}

// CSVSequence returns whether the sequence should be included in the CSV output.
// It retrieves the setting from the underlying options.
func (opt Options) CSVSequence() bool {
	return opt.pointer.csv_sequence
}

// CSVQuality returns whether the quality should be included in the CSV output.
// It retrieves the setting from the underlying options.
func (opt Options) CSVQuality() bool {
	return opt.pointer.csv_quality
}

// CSVKeys returns the list of keys to include in the CSV output.
// It retrieves the keys from the underlying options.
func (opt Options) CSVKeys() []string {
	return opt.pointer.csv_keys
}

// CSVSeparator returns the separator used in the CSV output.
// It retrieves the separator from the underlying options.
func (opt Options) CSVSeparator() string {
	return opt.pointer.csv_separator
}

// CSVNAValue returns the value used for NA entries in the CSV output.
// It retrieves the NA value from the underlying options.
func (opt Options) CSVNAValue() string {
	return opt.pointer.csv_navalue
}

// CSVAutoColumn returns whether to automatically determine the CSV format.
// It retrieves the setting from the underlying options.
func (opt Options) CSVAutoColumn() bool {
	return opt.pointer.csv_auto
}

// HasSource returns whether a source has been specified.
// It checks if the source field in the underlying options is not empty.
func (opt Options) HasSource() bool {
	return opt.pointer.source != ""
}

// Source returns the source of the data.
// It retrieves the source from the underlying options.
func (opt Options) Source() string {
	return opt.pointer.source
}

func OptionFileName(filename string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.filename = filename
	})

	return f
}

// OptionCloseFile returns a WithOption function that sets the closefile option to true.
func OptionCloseFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.closefile = true
	})

	return f
}

// OptionDontCloseFile returns a WithOption function that sets the closefile option to false.
func OptionDontCloseFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.closefile = false
	})

	return f
}

// OptionsAppendFile returns a WithOption function that sets the appendfile option.
// Parameters:
//   - append: A boolean indicating whether to append to the file.
func OptionsAppendFile(append bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.appendfile = append
	})

	return f
}

// OptionNoOrder returns a WithOption function that sets the no_order option.
// Parameters:
//   - no_order: A boolean indicating whether to process items in no specific order.
func OptionNoOrder(no_order bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.no_order = no_order
	})

	return f
}

// OptionsCompressed returns a WithOption function that sets the compressed option.
// Parameters:
//   - compressed: A boolean indicating whether the input data is compressed.
func OptionsCompressed(compressed bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.compressed = compressed
	})

	return f
}

// OptionsSkipEmptySequence returns a WithOption function that sets the skip_empty option.
// Parameters:
//   - skip: A boolean indicating whether to skip empty sequences.
func OptionsSkipEmptySequence(skip bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.skip_empty = skip
	})

	return f
}

// OptionsNewFile returns a WithOption function that sets the appendfile option to false,
// indicating that a new file should be created instead of appending to an existing one.
func OptionsNewFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.appendfile = false
	})

	return f
}

// OptionsParallelWorkers returns a WithOption function that sets the number of parallel workers.
// Parameters:
//   - nworkers: An integer specifying the number of parallel workers to use.
func OptionsParallelWorkers(nworkers int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.parallel_workers = nworkers
	})

	return f
}

// OptionsBatchSize returns a WithOption function that sets the batch_size option.
// Parameters:
//   - size: An integer specifying the size of the batch to be processed.
func OptionsBatchSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batch_size = size
	})

	return f
}

// OptionsBatchSizeDefault returns a WithOption function that sets the default batch_size option.
// Parameters:
//   - bp: An integer specifying the default size of the batch to be processed.
func OptionsBatchSizeDefault(bp int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batch_size = bp
	})

	return f
}

// OptionsFullFileBatch returns a WithOption function that sets the full_file_batch option.
// Parameters:
//   - full: A boolean indicating whether to process the full file in a single batch.
func OptionsFullFileBatch(full bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.full_file_batch = full
	})

	return f
}

// OptionsSource returns a WithOption function that sets the source option.
// Parameters:
//   - source: A string specifying the source of the data.
func OptionsSource(source string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.source = source
	})

	return f
}

// OptionsWithProgressBar returns a WithOption function that sets the with_progress_bar option to true,
// indicating that a progress bar should be displayed during processing.
func OptionsWithProgressBar() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_progress_bar = true
	})

	return f
}

// OptionsWithoutProgressBar returns a WithOption function that sets the with_progress_bar option to false,
// indicating that a progress bar should not be displayed during processing.
func OptionsWithoutProgressBar() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_progress_bar = false
	})

	return f
}

// CSVId returns a WithOption function that sets the csv_id option.
// Parameters:
//   - include: A boolean indicating whether to include the ID in the CSV output.
func CSVId(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_id = include
	})

	return f
}

// CSVSequence returns a WithOption function that sets the csv_sequence option.
// Parameters:
//   - include: A boolean indicating whether to include the sequence in the CSV output.
func CSVSequence(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_sequence = include
	})

	return f
}

// CSVQuality returns a WithOption function that sets the csv_quality option.
// Parameters:
//   - include: A boolean indicating whether to include the quality in the CSV output.
func CSVQuality(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_quality = include
	})

	return f
}

// CSVDefinition returns a WithOption function that sets the csv_definition option.
// Parameters:
//   - include: A boolean indicating whether to include the definition in the CSV output.
func CSVDefinition(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_definition = include
	})

	return f
}

// CSVCount returns a WithOption function that sets the csv_count option.
// Parameters:
//   - include: A boolean indicating whether to include the count in the CSV output.
func CSVCount(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_count = include
	})

	return f
}

// CSVTaxon returns a WithOption function that sets the csv_taxon option.
// Parameters:
//   - include: A boolean indicating whether to include the taxon in the CSV output.
func CSVTaxon(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_taxon = include
	})

	return f
}

// CSVKey returns a WithOption function that adds a key to the list of keys to include in the CSV output.
// Parameters:
//   - key: A string specifying the key to include in the CSV output.
func CSVKey(key string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_keys = append(opt.pointer.csv_keys, key)
	})

	return f
}

// CSVKeys returns a WithOption function that adds multiple keys to the list of keys to include in the CSV output.
// Parameters:
//   - keys: A slice of strings specifying the keys to include in the CSV output.
func CSVKeys(keys []string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_keys = append(opt.pointer.csv_keys, keys...)
	})

	return f
}

// CSVSeparator returns a WithOption function that sets the separator used in the CSV output.
// Parameters:
//   - separator: A string specifying the separator to use in the CSV output.
func CSVSeparator(separator string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_separator = separator
	})

	return f
}

// CSVNAValue returns a WithOption function that sets the value used for NA entries in the CSV output.
// Parameters:
//   - navalue: A string specifying the value to use for NA entries in the CSV output.
func CSVNAValue(navalue string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_navalue = navalue
	})

	return f
}

// CSVAutoColumn returns a WithOption function that sets the csv_auto option.
// Parameters:
//   - auto: A boolean indicating whether to automatically determine the CSV format.
func CSVAutoColumn(auto bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_auto = auto
	})

	return f
}
