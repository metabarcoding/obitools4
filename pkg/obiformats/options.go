package obiformats

import (
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type __options__ struct {
	fastseq_header_parser obiseq.SeqAnnotator
	fastseq_header_writer BioSequenceFormater
	seqBatchFormater      FormatSeqBatch
	with_progress_bar     bool
	buffer_size           int
	batch_size            int
	total_seq_size        int
	full_file_batch       bool
	parallel_workers      int
	no_order              bool
	closefile             bool
	appendfile            bool
	compressed            bool
	skip_empty            bool
	with_quality          bool
	csv_id                bool
	csv_sequence          bool
	csv_quality           bool
	csv_definition        bool
	csv_count             bool
	csv_taxon             bool
	csv_keys              []string
	csv_separator         string
	csv_navalue           string
	csv_auto              bool
	paired_filename       string
	source                string
	with_feature_table    bool
	with_pattern          bool
	with_parent           bool
	with_path             bool
	with_rank             bool
	with_taxid            bool
	with_scientific_name  bool
	without_root_path     bool
	raw_taxid             bool
	u_to_t                bool
	with_metadata         []string
}

type Options struct {
	pointer *__options__
}

type WithOption func(Options)

func MakeOptions(setters []WithOption) Options {
	o := __options__{
		fastseq_header_parser: ParseGuessedFastSeqHeader,
		fastseq_header_writer: FormatFastSeqJsonHeader,
		seqBatchFormater:      nil,
		with_progress_bar:     false,
		buffer_size:           2,
		parallel_workers:      obidefault.ReadParallelWorkers(),
		batch_size:            obidefault.BatchSize(),
		total_seq_size:        1024 * 1024 * 100, // 100 MB by default
		no_order:              false,
		full_file_batch:       false,
		closefile:             false,
		appendfile:            false,
		compressed:            false,
		skip_empty:            false,
		with_quality:          true,
		csv_id:                true,
		csv_definition:        false,
		csv_count:             false,
		csv_taxon:             false,
		csv_sequence:          true,
		csv_quality:           false,
		csv_separator:         ",",
		csv_navalue:           "NA",
		csv_keys:              make([]string, 0),
		csv_auto:              false,
		paired_filename:       "",
		source:                "unknown",
		with_feature_table:    false,
		with_pattern:          true,
		with_parent:           false,
		with_path:             false,
		with_rank:             true,
		with_taxid:            true,
		u_to_t:                false,
		with_scientific_name:  false,
		without_root_path:     false,
		raw_taxid:             false,
	}

	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

func (opt Options) BatchSize() int {
	return opt.pointer.batch_size
}

func (opt Options) TotalSeqSize() int {
	return opt.pointer.total_seq_size
}

func (opt Options) FullFileBatch() bool {
	return opt.pointer.full_file_batch
}

func (opt Options) ParallelWorkers() int {
	return opt.pointer.parallel_workers
}

func (opt Options) ParseFastSeqHeader() obiseq.SeqAnnotator {
	return opt.pointer.fastseq_header_parser
}

func (opt Options) FormatFastSeqHeader() func(*obiseq.BioSequence) string {
	return opt.pointer.fastseq_header_writer
}

func (opt Options) SequenceFormater() FormatSeqBatch {
	return opt.pointer.seqBatchFormater
}

func (opt Options) NoOrder() bool {
	return opt.pointer.no_order
}

func (opt Options) ProgressBar() bool {
	return opt.pointer.with_progress_bar
}

func (opt Options) CloseFile() bool {
	return opt.pointer.closefile
}

func (opt Options) AppendFile() bool {
	return opt.pointer.appendfile
}

func (opt Options) CompressedFile() bool {
	return opt.pointer.compressed
}

func (opt Options) SkipEmptySequence() bool {
	return opt.pointer.skip_empty
}

func (opt Options) ReadQualities() bool {
	return opt.pointer.with_quality
}

func (opt Options) CSVId() bool {
	return opt.pointer.csv_id
}

func (opt Options) CSVDefinition() bool {
	return opt.pointer.csv_definition
}

func (opt Options) CSVCount() bool {
	return opt.pointer.csv_count
}

func (opt Options) CSVTaxon() bool {
	return opt.pointer.csv_taxon
}

func (opt Options) CSVSequence() bool {
	return opt.pointer.csv_sequence
}

func (opt Options) CSVQuality() bool {
	return opt.pointer.csv_quality
}

func (opt Options) CSVKeys() []string {
	return opt.pointer.csv_keys
}

func (opt Options) CSVSeparator() string {
	return opt.pointer.csv_separator
}

func (opt Options) CSVNAValue() string {
	return opt.pointer.csv_navalue
}

func (opt Options) CSVAutoColumn() bool {
	return opt.pointer.csv_auto
}

func (opt Options) HaveToSavePaired() bool {
	return opt.pointer.paired_filename != ""
}

func (opt Options) PairedFileName() string {
	return opt.pointer.paired_filename
}

func (opt Options) HasSource() bool {
	return opt.pointer.source != ""
}

func (opt Options) Source() string {
	return opt.pointer.source
}

func (opt Options) WithFeatureTable() bool {
	return opt.pointer.with_feature_table
}

// WithPattern returns whether the pattern option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithPattern() bool {
	return o.pointer.with_pattern
}

// WithParent returns whether the parent option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithParent() bool {
	return o.pointer.with_parent
}

// WithPath returns whether the path option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithPath() bool {
	return o.pointer.with_path
}

// WithRank returns whether the rank option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithRank() bool {
	return o.pointer.with_rank
}

func (o *Options) WithTaxid() bool {
	return o.pointer.with_taxid
}

// WithScientificName returns whether the scientific name option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithScientificName() bool {
	return o.pointer.with_scientific_name
}

// WithoutRootPath returns whether the root path option is enabled.
func (o *Options) WithoutRootPath() bool {
	return o.pointer.without_root_path
}

// RawTaxid returns whether the raw taxid option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) RawTaxid() bool {
	return o.pointer.raw_taxid
}

func (o *Options) UtoT() bool {
	return o.pointer.u_to_t
}

// WithMetadata returns a slice of strings containing the metadata
// associated with the Options instance. It retrieves the metadata
// from the pointer's with_metadata field.
func (o *Options) WithMetadata() []string {
	if o.WithPattern() {
		idx := slices.Index(o.pointer.with_metadata, "query")
		if idx >= 0 {
			o.pointer.with_metadata = slices.Delete(o.pointer.with_metadata, idx, idx+1)
		}
	}

	return o.pointer.with_metadata
}

func OptionCloseFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.closefile = true
	})

	return f
}

func OptionDontCloseFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.closefile = false
	})

	return f
}

func OptionsAppendFile(append bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.appendfile = append
	})

	return f
}

func OptionNoOrder(no_order bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.no_order = no_order
	})

	return f
}

func OptionsCompressed(compressed bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.compressed = compressed
	})

	return f
}

func OptionsSkipEmptySequence(skip bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.skip_empty = skip
	})

	return f
}

func OptionsReadQualities(read bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_quality = read
	})

	return f
}

func OptionsNewFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.appendfile = false
	})

	return f
}

func OptionsFastSeqHeaderParser(parser obiseq.SeqAnnotator) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.fastseq_header_parser = parser
	})

	return f
}

func OptionFastSeqDoNotParseHeader() WithOption {
	return OptionsFastSeqHeaderParser(nil)
}

func OptionsFastSeqDefaultHeaderParser() WithOption {
	return OptionsFastSeqHeaderParser(ParseGuessedFastSeqHeader)
}

// OptionsFastSeqHeaderFormat allows foor specifying the format
// used to write FASTA and FASTQ sequence.
func OptionsFastSeqHeaderFormat(format func(*obiseq.BioSequence) string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.fastseq_header_writer = format
	})

	return f
}

func OptionsSequenceFormater(formater FormatSeqBatch) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.seqBatchFormater = formater
	})

	return f
}

func OptionsParallelWorkers(nworkers int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.parallel_workers = nworkers
	})

	return f
}

func OptionsBatchSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batch_size = size
	})

	return f
}

func OptionsBatchSizeDefault(bp int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batch_size = bp
	})

	return f
}

func OptionsFullFileBatch(full bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.full_file_batch = full
	})

	return f
}

func OptionsSource(source string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.source = source
	})

	return f
}

func OptionsWithProgressBar() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_progress_bar = true
	})

	return f
}

func OptionsWithoutProgressBar() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_progress_bar = false
	})

	return f
}

func WritePairedReadsTo(filename string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.paired_filename = filename
	})

	return f
}

func CSVId(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_id = include
	})

	return f
}

func CSVSequence(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_sequence = include
	})

	return f
}

func CSVQuality(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_quality = include
	})

	return f
}

func CSVDefinition(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_definition = include
	})

	return f
}

func CSVCount(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_count = include
	})

	return f
}

func CSVTaxon(include bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_taxon = include
	})

	return f
}

func CSVKey(key string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_keys = append(opt.pointer.csv_keys, key)
	})

	return f
}

func CSVKeys(keys []string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_keys = append(opt.pointer.csv_keys, keys...)
	})

	return f
}

func CSVSeparator(separator string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_separator = separator
	})

	return f
}

func CSVNAValue(navalue string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_navalue = navalue
	})

	return f
}

func CSVAutoColumn(auto bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.csv_auto = auto
	})

	return f
}

func WithFeatureTable(with bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_feature_table = with
	})

	return f
}

func OptionsWithPattern(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_pattern = value
	})

	return f
}

func OptionsWithParent(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_parent = value
	})

	return f
}

func OptionsWithPath(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_path = value
	})

	return f
}

func OptionsWithRank(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_rank = value
	})

	return f
}

func OptionsWithTaxid(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_taxid = value
	})

	return f
}

func OptionsWithScientificName(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_scientific_name = value
	})

	return f
}

func OptionWithoutRootPath(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.without_root_path = value
	})
	return f
}

func OptionsRawTaxid(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.raw_taxid = value
	})

	return f
}

func OptionsUtoT(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.u_to_t = value
	})

	return f
}

func OptionsWithMetadata(values ...string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_metadata = values
	})
	return f
}
