package obiapat

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type _Options struct {
	minLength       int
	maxLength       int
	circular        bool
	forwardError    int
	reverseError    int
	bufferSize      int
	batchSize       int
	parallelWorkers int
}

// Options stores a set of option usable by the
// PCR simulation algotithm.
type Options struct {
	pointer *_Options
}

// WithOption is the standard type for function
// declaring options.
type WithOption func(Options)

// MinLength method returns minimum length of
// the searched amplicon (length of the primers
// excluded)
func (options Options) MinLength() int {
	return options.pointer.minLength
}

// MaxLength method returns maximum length of
// the searched amplicon (length of the primers
// excluded)
func (options Options) MaxLength() int {
	return options.pointer.maxLength
}

// ForwardError method returns the number of
// error allowed when matching the forward
// primer.
func (options Options) ForwardError() int {
	return options.pointer.forwardError
}

// ReverseError method returns the number of
// error allowed when matching the reverse
// primer.
func (options Options) ReverseError() int {
	return options.pointer.reverseError
}

// Circular method returns the topology option.
// true for circular, false for linear
func (options Options) Circular() bool {
	return options.pointer.circular
}

// BufferSize returns the size of the channel
// buffer specified by the options
func (options Options) BufferSize() int {
	return options.pointer.bufferSize
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
		minLength:       0,
		maxLength:       0,
		forwardError:    0,
		reverseError:    0,
		circular:        false,
		parallelWorkers: 4,
		batchSize:       100,
		bufferSize:      100,
	}

	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

// OptionMinLength sets the minimum length of
// the searched amplicon (length of the primers
// excluded)
func OptionMinLength(minLength int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.minLength = minLength
	})

	return f
}

// OptionMaxLength sets the maximum length of
// the searched amplicon (length of the primers
// excluded)
func OptionMaxLength(maxLength int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.maxLength = maxLength
	})

	return f
}

// OptionForwardError sets the number of
// error allowed when matching the forward
// primer.
func OptionForwardError(max int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.forwardError = max
	})

	return f
}

// OptionReverseError sets the number of
// error allowed when matching the reverse
// primer.
func OptionReverseError(max int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.reverseError = max
	})

	return f
}

// OptionCircular sets the topology option.
// true for circular, false for linear
func OptionCircular(circular bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.circular = circular
	})

	return f
}

// OptionBufferSize sets the requested channel
// buffer size.
func OptionBufferSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.bufferSize = size
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

func _Pcr(seq ApatSequence, sequence obiseq.BioSequence,
	forward, cfwd, reverse, crev ApatPattern,
	opt Options) obiseq.BioSequenceSlice {
	results := make(obiseq.BioSequenceSlice, 0, 10)

	forwardMatches := forward.FindAllIndex(seq)

	if forwardMatches != nil {

		begin := forwardMatches[0][0]
		length := seq.Length() - begin

		if opt.pointer.maxLength > 0 {
			length = forwardMatches[len(forwardMatches)-1][2] - begin + opt.MaxLength() + reverse.Length()
		}

		if opt.Circular() {
			begin = 0
			length = seq.Length() + _MaxPatLen
		}

		reverseMatches := crev.FindAllIndex(seq, begin, length)

		if reverseMatches != nil {
			for _, fm := range forwardMatches {

				posi := fm[0]

				if posi < seq.Length() {

					erri := fm[2]

					for _, rm := range reverseMatches {
						posj := rm[0]
						if posj < seq.Length() {
							posj := rm[1]
							errj := rm[2]
							length = 0

							if posj > posi {
								length = rm[0] - fm[1]
							} else {
								if opt.Circular() {
									length = rm[0] + seq.Length() - posi - forward.Length()
								}
							}
							if length > 0 && // For when primers touch or overlap
								(opt.MinLength() == 0 || length >= opt.MinLength()) &&
								(opt.MaxLength() == 0 || length <= opt.MaxLength()) {
								amplicon, _ := sequence.Subsequence(fm[1], rm[0], opt.pointer.circular)
								annot := amplicon.Annotations()
								goutils.CopyMap(annot, sequence.Annotations())
								annot["forward_primer"] = forward.String()

								match, _ := sequence.Subsequence(fm[0], fm[1], opt.pointer.circular)
								annot["forward_match"] = match.String()
								(&match).Recycle()

								annot["forward_error"] = erri

								annot["reverse_primer"] = reverse.String()
								match, _ = sequence.Subsequence(rm[0], rm[1], opt.pointer.circular)
								match = match.ReverseComplement(true)
								annot["reverse_match"] = match.String()
								(&match).Recycle()

								annot["reverse_error"] = errj
								results = append(results, amplicon)
							}
						}
					}
				}
			}
		}
	}

	forwardMatches = reverse.FindAllIndex(seq)

	if forwardMatches != nil {

		begin := forwardMatches[0][0]
		length := seq.Length() - begin

		if opt.pointer.maxLength > 0 {
			length = forwardMatches[len(forwardMatches)-1][2] - begin + opt.MaxLength() + reverse.Length()
		}

		if opt.Circular() {
			begin = 0
			length = seq.Length() + _MaxPatLen
		}

		reverseMatches := cfwd.FindAllIndex(seq, begin, length)

		if reverseMatches != nil {
			for _, fm := range forwardMatches {

				posi := fm[0]

				if posi < seq.Length() {

					erri := fm[2]

					for _, rm := range reverseMatches {
						posj := rm[0]
						if posj < seq.Length() {
							posj := rm[1]
							errj := rm[2]
							length = 0

							if posj > posi {
								length = rm[0] - fm[1]
							} else {
								if opt.Circular() {
									length = rm[0] + seq.Length() - posi - forward.Length()
								}
							}
							if length > 0 && // For when primers touch or overlap
								(opt.MinLength() == 0 || length >= opt.MinLength()) &&
								(opt.MaxLength() == 0 || length <= opt.MaxLength()) {
								amplicon, _ := sequence.Subsequence(fm[1], rm[0], opt.pointer.circular)
								amplicon = amplicon.ReverseComplement(true)

								annot := amplicon.Annotations()
								goutils.CopyMap(annot, sequence.Annotations())
								annot["forward_primer"] = forward.String()

								match, _ := sequence.Subsequence(rm[0], rm[1], opt.pointer.circular)
								match.ReverseComplement(true)
								annot["forward_match"] = match.String()
								(&match).Recycle()

								annot["forward_error"] = errj

								annot["reverse_primer"] = reverse.String()
								match, _ = sequence.Subsequence(fm[0], fm[1], opt.pointer.circular)
								annot["reverse_match"] = match.String()
								(&match).Recycle()

								annot["reverse_error"] = erri
								results = append(results, amplicon)
							}
						}
					}
				}
			}
		}
	}
	return results
}

// PCR runs the PCR simulation algorithm on a single
// obiseq.BioSequence instance. PCR parameters are
// specified using the corresponding Option functions
// defined for the PCR algorithm.
func PCR(sequence obiseq.BioSequence,
	forward, reverse string, options ...WithOption) obiseq.BioSequenceSlice {

	opt := MakeOptions(options)

	seq, _ := MakeApatSequence(sequence, opt.Circular())

	fwd, _ := MakeApatPattern(forward, opt.ForwardError())
	rev, _ := MakeApatPattern(reverse, opt.ReverseError())
	cfwd, _ := fwd.ReverseComplement()
	crev, _ := rev.ReverseComplement()

	results := _Pcr(seq, sequence,
		fwd, cfwd, rev, crev,
		opt)

	seq.Free()

	fwd.Free()
	rev.Free()
	cfwd.Free()
	crev.Free()

	return results
}

// PCRSlice runs the PCR simulation algorithm on a set of
// obiseq.BioSequence instances grouped in a obiseq.BioSequenceSlice.
// PCR parameters are
// specified using the corresponding Option functions
// defined for the PCR algorithm.
func PCRSlice(sequences obiseq.BioSequenceSlice,
	forward, reverse string, options ...WithOption) obiseq.BioSequenceSlice {

	results := make(obiseq.BioSequenceSlice, 0, len(sequences))

	opt := MakeOptions(options)

	fwd, _ := MakeApatPattern(forward, opt.ForwardError())
	rev, _ := MakeApatPattern(reverse, opt.ReverseError())
	cfwd, _ := fwd.ReverseComplement()
	crev, _ := rev.ReverseComplement()

	if len(sequences) > 0 {
		seq, _ := MakeApatSequence(sequences[0], opt.Circular())
		amplicons := _Pcr(seq, sequences[0],
			fwd, cfwd, rev, crev,
			opt)

		if len(amplicons) > 0 {
			results = append(results, amplicons...)
		}

		for _, sequence := range sequences[1:] {
			seq, _ := MakeApatSequence(sequence, opt.Circular(), seq)
			amplicons = _Pcr(seq, sequence,
				fwd, cfwd, rev, crev,
				opt)
			if len(amplicons) > 0 {
				results = append(results, amplicons...)
			}
		}

		seq.Free()
	}

	fwd.Free()
	rev.Free()
	cfwd.Free()
	crev.Free()

	return results
}

// PCRSliceWorker is a worker function builder which produce
// job function usable by the obiseq.MakeISliceWorker function.
func PCRSliceWorker(forward, reverse string,
	options ...WithOption) obiseq.SeqSliceWorker {

	worker := func(sequences obiseq.BioSequenceSlice) obiseq.BioSequenceSlice {
		return PCRSlice(sequences, forward, reverse, options...)
	}

	return worker
}
