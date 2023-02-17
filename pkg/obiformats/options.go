package obiformats

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type __options__ struct {
	fastseq_header_parser obiseq.SeqAnnotator
	fastseq_header_writer func(*obiseq.BioSequence) string
	with_progress_bar     bool
	buffer_size           int
	batch_size            int
	quality_shift         int
	parallel_workers      int
	closefile             bool
	appendfile            bool
	compressed            bool
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
		quality_shift:         33,
		parallel_workers:      4,
		batch_size:            5000,
		closefile:             false,
		appendfile:            false,
		compressed:            false,
	}

	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

func (opt Options) QualityShift() int {
	return opt.pointer.quality_shift
}

func (opt Options) BufferSize() int {
	return opt.pointer.buffer_size
}

func (opt Options) BatchSize() int {
	return opt.pointer.batch_size
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

func OptionsBufferSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.buffer_size = size
	})

	return f
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

func OptionsAppendFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.appendfile = true
	})

	return f
}

func OptionsCompressed() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.compressed = true
	})

	return f
}

func OptionsNewFile() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.appendfile = false
	})

	return f
}

// Allows to specify the ascii code corresponding to
// a quality of 0 in fastq encoded quality scores.
func OptionsQualityShift(shift int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.quality_shift = shift
	})

	return f
}

// Allows to specify a quality shift of 33, corresponding
// to a FastQ file qualities encoded following Sanger
// convention. This corresponds to Illumina produced FastQ
// files.
func OptionsQualitySanger() WithOption {
	return OptionsQualityShift(33)
}

// Allows to specify a quality shift of 64, corresponding
// to a FastQ file qualities encoded following the Solexa
// convention.
func OptionsQualitySolexa() WithOption {
	return OptionsQualityShift(64)
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
