package obingslibrary

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type _Options struct {
	discardErrors   bool
	unidentified    string
	allowedMismatch int
	withProgressBar bool
	parallelWorkers int
	batchSize       int
}

// Options stores a set of option usable by the
// PCR simulation algotithm.
type Options struct {
	pointer *_Options
}

// WithOption is the standard type for function
// declaring options.
type WithOption func(Options)

func OptionDiscardErrors(yes bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.discardErrors = yes
	})

	return f
}

func OptionUnidentified(filename string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.unidentified = filename
	})

	return f
}

func OptionWithProgressBar(yes bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.withProgressBar = yes
	})

	return f
}

func OptionAllowedMismatches(count int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.allowedMismatch = count
	})

	return f
}

// OptionParallelWorkers sets how many search
// jobs will be run in parallel.
func OptionParallelWorkers(nworkers int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.parallelWorkers = nworkers
	})

	return f
}

// OptionBatchSize sets the requested sequence
// batch size.
func OptionBatchSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batchSize = size
	})

	return f
}

func (options Options) DiscardErrors() bool {
	return options.pointer.unidentified == "" || options.pointer.discardErrors
}

func (options Options) Unidentified() string {
	return options.pointer.unidentified
}

func (options Options) AllowedMismatch() int {
	return options.pointer.allowedMismatch
}

func (options Options) WithProgressBar() bool {
	return options.pointer.withProgressBar
}

// BatchSize returns the size of the
// sequence batch used by the PCR algorithm
func (options Options) BatchSize() int {
	return options.pointer.batchSize
}

// ParallelWorkers returns how many search
// jobs will be run in parallel.
func (options Options) ParallelWorkers() int {
	return options.pointer.parallelWorkers
}

// MakeOptions buils a new default option set for
// the PCR simulation algoithm.
func MakeOptions(setters []WithOption) Options {
	o := _Options{
		discardErrors:   true,
		unidentified:    "",
		allowedMismatch: 0,
		withProgressBar: false,
		parallelWorkers: 4,
		batchSize:       1000,
	}

	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

func _ExtractBarcodeSlice(ngslibrary NGSLibrary,
	sequences obiseq.BioSequenceSlice,
	options Options) obiseq.BioSequenceSlice {
	newSlice := make(obiseq.BioSequenceSlice, 0, len(sequences))

	for _, seq := range sequences {
		s, err := ngslibrary.ExtractBarcode(seq, true)
		if err == nil || !options.pointer.discardErrors {
			newSlice = append(newSlice, s)
		}
	}

	return newSlice
}

func ExtractBarcodeSlice(ngslibrary NGSLibrary,
	sequences obiseq.BioSequenceSlice,
	options ...WithOption) obiseq.BioSequenceSlice {

	opt := MakeOptions(options)

	ngslibrary.Compile(opt.AllowedMismatch())

	return _ExtractBarcodeSlice(ngslibrary, sequences, opt)
}

func ExtractBarcodeSliceWorker(ngslibrary NGSLibrary,
	options ...WithOption) obiseq.SeqSliceWorker {

	opt := MakeOptions(options)

	ngslibrary.Compile(opt.AllowedMismatch())

	worker := func(sequences obiseq.BioSequenceSlice) obiseq.BioSequenceSlice {
		return _ExtractBarcodeSlice(ngslibrary, sequences, opt)
	}

	return worker
}
