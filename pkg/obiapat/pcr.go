package obiapat

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type __options__ struct {
	min_length       int
	max_length       int
	circular         bool
	forward_error    int
	reverse_error    int
	buffer_size      int
	batch_size       int
	parallel_workers int
}

type Options struct {
	pointer *__options__
}

type WithOption func(Options)

func (options Options) MinLength() int {
	return options.pointer.min_length
}

func (options Options) MaxLength() int {
	return options.pointer.max_length
}

func (options Options) ForwardError() int {
	return options.pointer.forward_error
}

func (options Options) ReverseError() int {
	return options.pointer.reverse_error
}

func (options Options) Circular() bool {
	return options.pointer.circular
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

func MakeOptions(setters []WithOption) Options {
	o := __options__{
		min_length:       0,
		max_length:       0,
		forward_error:    0,
		reverse_error:    0,
		circular:         false,
		parallel_workers: 4,
		batch_size:       100,
		buffer_size:      100,
	}

	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

func OptionMinLength(min_length int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.min_length = min_length
	})

	return f
}

func OptionMaxLength(max_length int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.max_length = max_length
	})

	return f
}

func OptionForwardError(max int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.forward_error = max
	})

	return f
}

func OptionReverseError(max int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.reverse_error = max
	})

	return f
}

func OptionCircular(circular bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.circular = circular
	})

	return f
}

func OptionBufferSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.buffer_size = size
	})

	return f
}

func OptionParallelWorkers(nworkers int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.parallel_workers = nworkers
	})

	return f
}

func OptionBatchSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batch_size = size
	})

	return f
}

func __pcr__(seq ApatSequence, sequence obiseq.BioSequence,
	forward, cfwd, reverse, crev ApatPattern,
	opt Options) obiseq.BioSequenceSlice {
	results := make(obiseq.BioSequenceSlice, 0, 10)

	forward_matches := forward.FindAllIndex(seq)

	if forward_matches != nil {

		begin := forward_matches[0][0]
		length := seq.Length() - begin

		if opt.pointer.max_length > 0 {
			length = forward_matches[len(forward_matches)-1][2] - begin + opt.MaxLength() + reverse.Length()
		}

		if opt.Circular() {
			begin = 0
			length = seq.Length() + MAX_PAT_LEN
		}

		reverse_matches := crev.FindAllIndex(seq, begin, length)

		if reverse_matches != nil {
			for _, fm := range forward_matches {

				posi := fm[0]

				if posi < seq.Length() {

					erri := fm[2]

					for _, rm := range reverse_matches {
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
								match.Revoke()

								annot["forward_error"] = erri

								annot["reverse_primer"] = reverse.String()
								match, _ = sequence.Subsequence(rm[0], rm[1], opt.pointer.circular)
								match = match.ReverseComplement(true)
								annot["reverse_match"] = match.String()
								match.Revoke()

								annot["reverse_error"] = errj
								results = append(results, amplicon)
							}
						}
					}
				}
			}
		}
	}

	forward_matches = reverse.FindAllIndex(seq)

	if forward_matches != nil {

		begin := forward_matches[0][0]
		length := seq.Length() - begin

		if opt.pointer.max_length > 0 {
			length = forward_matches[len(forward_matches)-1][2] - begin + opt.MaxLength() + reverse.Length()
		}

		if opt.Circular() {
			begin = 0
			length = seq.Length() + MAX_PAT_LEN
		}

		reverse_matches := cfwd.FindAllIndex(seq, begin, length)

		if reverse_matches != nil {
			for _, fm := range forward_matches {

				posi := fm[0]

				if posi < seq.Length() {

					erri := fm[2]

					for _, rm := range reverse_matches {
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
								match.Revoke()

								annot["forward_error"] = errj

								annot["reverse_primer"] = reverse.String()
								match, _ = sequence.Subsequence(fm[0], fm[1], opt.pointer.circular)
								annot["reverse_match"] = match.String()
								match.Revoke()

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

func PCR(sequence obiseq.BioSequence,
	forward, reverse string, options ...WithOption) obiseq.BioSequenceSlice {

	opt := MakeOptions(options)

	seq, _ := MakeApatSequence(sequence, opt.Circular())

	fwd, _ := MakeApatPattern(forward, opt.ForwardError())
	rev, _ := MakeApatPattern(reverse, opt.ReverseError())
	cfwd, _ := fwd.ReverseComplement()
	crev, _ := rev.ReverseComplement()

	results := __pcr__(seq, sequence,
		fwd, cfwd, rev, crev,
		opt)

	seq.Free()

	fwd.Free()
	rev.Free()
	cfwd.Free()
	crev.Free()

	return results
}

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
		amplicons := __pcr__(seq, sequences[0],
			fwd, cfwd, rev, crev,
			opt)

		if len(amplicons) > 0 {
			results = append(results, amplicons...)
		}

		for _, sequence := range sequences[1:] {
			seq, _ := MakeApatSequence(sequence, opt.Circular(), seq)
			amplicons = __pcr__(seq, sequence,
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

func PCRSliceWorker(forward, reverse string,
	options ...WithOption) obiseq.SeqSliceWorker {

	worker := func(sequences obiseq.BioSequenceSlice) obiseq.BioSequenceSlice {
		return PCRSlice(sequences, forward, reverse, options...)
	}

	return worker
}
