package obiformats

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type __options__ struct {
	fastseq_header_parser obiseq.SeqAnnotator
	fastseq_header_writer func(*obiseq.BioSequence) string
	with_progress_bar     bool
	buffer_size           int
	batch_size            int
	total_seq_size        int
	full_file_batch       bool
	parallel_workers      int
	closefile             bool
	appendfile            bool
	compressed            bool
	skip_empty            bool
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
}

type Options struct {
	pointer *__options__
}

type WithOption func(Options)

func MakeOptions(setters []WithOption) Options {
	o := __options__{
		fastseq_header_parser: ParseGuessedFastSeqHeader,
		fastseq_header_writer: FormatFastSeqJsonHeader,
		with_progress_bar:     false,
		buffer_size:           2,
		parallel_workers:      obioptions.CLIReadParallelWorkers(),
		batch_size:            obioptions.CLIBatchSize(),
		total_seq_size:        1024 * 1024 * 100, // 100 MB by default
		full_file_batch:       false,
		closefile:             false,
		appendfile:            false,
		compressed:            false,
		skip_empty:            false,
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
		source:                "",
		with_feature_table:    false,
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
