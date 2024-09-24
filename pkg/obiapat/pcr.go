package obiapat

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type _Options struct {
	minLength       int
	maxLength       int
	circular        bool
	forwardError    int
	reverseError    int
	extension       int
	fullExtension   bool
	batchSize       int
	parallelWorkers int
	forward         ApatPattern
	cfwd            ApatPattern
	reverse         ApatPattern
	crev            ApatPattern
}

// Options stores a set of option usable by the
// PCR simulation algotithm.
type Options struct {
	pointer *_Options
}

// WithOption is the standard type for function
// declaring options.
type WithOption func(Options)

func (options Options) HasExtension() bool {
	return options.pointer.extension > -1

}

func (options Options) Extension() int {
	return options.pointer.extension
}

func (options Options) OnlyFullExtension() bool {
	return options.pointer.fullExtension
}

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
		extension:       -1,
		fullExtension:   false,
		circular:        false,
		parallelWorkers: obioptions.CLIParallelWorkers(),
		batchSize:       100,
		forward:         NilApatPattern,
		cfwd:            NilApatPattern,
		reverse:         NilApatPattern,
		crev:            NilApatPattern,
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
func OptionForwardPrimer(primer string, max int) WithOption {

	f := WithOption(func(opt Options) {
		var err error

		opt.pointer.forward, err = MakeApatPattern(primer, max, false)
		if err != nil {
			log.Fatalf("error : %v\n", err)
		}

		opt.pointer.cfwd, err = opt.pointer.forward.ReverseComplement()
		if err != nil {
			log.Fatalf("error : %v\n", err)
		}
		opt.pointer.forwardError = max
	})

	return f
}

// OptionWithExtension sets the length of the extension to be added to the sequence.
//
// An negative value indicates that no extension is added
// The extension parameter is an integer that represents the extension value to be set.
// The returned function is of type WithOption.
func OptionWithExtension(extension int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.extension = extension
	})

	return f
}

func OptionOnlyFullExtension(full bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.fullExtension = full
	})

	return f
}

// OptionForwardError sets the number of
// error allowed when matching the forward
// primer.
func OptionReversePrimer(primer string, max int) WithOption {
	f := WithOption(func(opt Options) {
		var err error

		opt.pointer.reverse, err = MakeApatPattern(primer, max, false)
		if err != nil {
			log.Fatalf("error : %v\n", err)
		}

		opt.pointer.crev, err = opt.pointer.reverse.ReverseComplement()
		if err != nil {
			log.Fatalf("error : %v\n", err)
		}
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

func _Pcr(seq ApatSequence,
	opt Options) obiseq.BioSequenceSlice {
	results := make(obiseq.BioSequenceSlice, 0, 10)

	forward := opt.pointer.forward
	cfwd := opt.pointer.cfwd
	reverse := opt.pointer.reverse
	crev := opt.pointer.crev

	forwardMatches := forward.FindAllIndex(seq, 0, -1)

	if len(forwardMatches) > 0 {

		begin := forwardMatches[0][0]
		length := seq.Len() - begin

		if opt.pointer.maxLength > 0 {
			length = forwardMatches[len(forwardMatches)-1][1] - begin + opt.MaxLength() + reverse.Len()
		}

		if opt.Circular() {
			begin = 0
			length = seq.Len() + _MaxPatLen
		}

		reverseMatches := crev.FindAllIndex(seq, begin, length)

		if reverseMatches != nil {
			for _, fm := range forwardMatches {

				posi := fm[0]

				if posi < seq.Len() {

					erri := fm[2]

					for _, rm := range reverseMatches {
						posj := rm[0]
						if posj < seq.Len() {
							posj := rm[1]
							errj := rm[2]
							length = 0

							if posj > posi {
								length = rm[0] - fm[1]
							} else {
								if opt.Circular() {
									length = rm[0] + seq.Len() - posi - forward.Len()
								}
							}
							if length > 0 && // For when primers touch or overlap
								(opt.MinLength() == 0 || length >= opt.MinLength()) &&
								(opt.MaxLength() == 0 || length <= opt.MaxLength()) {
								var from, to int
								if opt.HasExtension() {
									from = fm[0] - opt.Extension()
									to = rm[1] + opt.Extension()
								} else {
									from = fm[1]
									to = rm[0]
								}

								if opt.HasExtension() && !opt.OnlyFullExtension() && !opt.Circular() {
									if from < 0 {
										from = 0
									}
									if to > seq.Len() {
										to = seq.Len()
									}
								}

								if (opt.HasExtension() && ((from >= 0 && to <= seq.Len()) || opt.Circular())) ||
									!opt.HasExtension() {

									amplicon, error := seq.pointer.reference.Subsequence(from, to, opt.Circular())

									if error != nil {
										log.Fatalf("error : %v\n", error)
									}

									log.Debugf("seq length : %d capacity : %d", amplicon.Len(), cap(amplicon.Sequence()))
									annot := amplicon.Annotations()
									obiutils.MustFillMap(annot, seq.pointer.reference.Annotations())

									annot["forward_primer"] = forward.String()

									match, _ := seq.pointer.reference.Subsequence(fm[0], fm[1], opt.pointer.circular)
									annot["forward_match"] = match.String()
									match.Recycle()

									annot["forward_error"] = erri

									annot["reverse_primer"] = reverse.String()
									match, _ = seq.pointer.reference.Subsequence(rm[0], rm[1], opt.pointer.circular)
									if match == nil {
										log.Fatalf("error in extracting sequence from reference: %d:%d (%v)\n", rm[0], rm[1], opt.pointer.circular)
									}
									match = match.ReverseComplement(true)
									annot["reverse_match"] = match.String()
									match.Recycle()

									annot["reverse_error"] = errj
									annot["direction"] = "forward"

									// log.Debugf("amplicon sequence capacity : %d", cap(amplicon.Sequence()))

									results = append(results, amplicon)

								}
							}
						}
					}
				}
			}
		}
	}

	forwardMatches = reverse.FindAllIndex(seq, 0, -1)

	if forwardMatches != nil {

		begin := forwardMatches[0][0]
		length := seq.Len() - begin

		if opt.pointer.maxLength > 0 {
			length = forwardMatches[len(forwardMatches)-1][1] - begin + opt.MaxLength() + reverse.Len()
		}

		if opt.Circular() {
			begin = 0
			length = seq.Len() + _MaxPatLen
		}

		reverseMatches := cfwd.FindAllIndex(seq, begin, length)

		if reverseMatches != nil {
			for _, fm := range forwardMatches {

				posi := fm[0]

				if posi < seq.Len() {

					erri := fm[2]

					for _, rm := range reverseMatches {
						posj := rm[0]
						if posj < seq.Len() {
							posj := rm[1]
							errj := rm[2]
							length = 0

							if posj > posi {
								length = rm[0] - fm[1]
							} else {
								if opt.Circular() {
									length = rm[0] + seq.Len() - posi - forward.Len()
								}
							}
							if length > 0 && // For when primers touch or overlap
								(opt.MinLength() == 0 || length >= opt.MinLength()) &&
								(opt.MaxLength() == 0 || length <= opt.MaxLength()) {
								var from, to int
								if opt.HasExtension() {
									from = fm[0] - opt.Extension()
									to = rm[1] + opt.Extension()
								} else {
									from = fm[1]
									to = rm[0]
								}

								if opt.HasExtension() && !opt.OnlyFullExtension() && !opt.Circular() {
									if from < 0 {
										from = 0
									}
									if to > seq.Len() {
										to = seq.Len()
									}
								}

								if (opt.HasExtension() && ((from >= 0 && to <= seq.Len()) || opt.Circular())) ||
									!opt.HasExtension() {
									amplicon, error := seq.pointer.reference.Subsequence(from, to, opt.pointer.circular)

									if error != nil {
										log.Fatalf("error : %v\n", error)
									}

									amplicon = amplicon.ReverseComplement(true)

									annot := amplicon.Annotations()
									obiutils.MustFillMap(annot, seq.pointer.reference.Annotations())
									annot["forward_primer"] = forward.String()

									match, _ := seq.pointer.reference.Subsequence(rm[0], rm[1], opt.pointer.circular)
									match.ReverseComplement(true)
									annot["forward_match"] = match.String()
									match.Recycle()

									annot["forward_error"] = errj

									annot["reverse_primer"] = reverse.String()
									match, _ = seq.pointer.reference.Subsequence(fm[0], fm[1], opt.pointer.circular)
									annot["reverse_match"] = match.String()
									match.Recycle()

									annot["reverse_error"] = erri
									annot["direction"] = "reverse"

									results = append(results, amplicon)
									// log.Debugf("amplicon sequence capacity : %d", cap(amplicon.Sequence()))
								}
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
func PCRSim(sequence *obiseq.BioSequence, options ...WithOption) obiseq.BioSequenceSlice {

	opt := MakeOptions(options)

	seq, _ := MakeApatSequence(sequence, opt.Circular())
	defer seq.Free()

	results := _Pcr(seq, opt)

	return results
}

func _PCRSlice(sequences obiseq.BioSequenceSlice,
	options Options) obiseq.BioSequenceSlice {

	results := make(obiseq.BioSequenceSlice, 0, len(sequences))

	if len(sequences) > 0 {
		seq, _ := MakeApatSequence(sequences[0], options.Circular())

		// if AllocatedApaSequences() == 0 {
		// 	log.Panicln("Bizarre....")
		// }
		amplicons := _Pcr(seq, options)

		if len(amplicons) > 0 {
			results = append(results, amplicons...)
		}

		for _, sequence := range sequences[1:] {
			seq, _ = MakeApatSequence(sequence, options.Circular(), seq)
			amplicons = _Pcr(seq, options)

			if len(amplicons) > 0 {
				results = append(results, amplicons...)
			}
		}

		//log.Debugln(AllocatedApaSequences())

		// seq.Free()
	}

	return results
}

// PCRSlice runs the PCR simulation algorithm on a set of
// obiseq.BioSequence instances grouped in a obiseq.BioSequenceSlice.
// PCR parameters are
// specified using the corresponding Option functions
// defined for the PCR algorithm.
func PCRSlice(sequences obiseq.BioSequenceSlice,
	options ...WithOption) obiseq.BioSequenceSlice {

	opt := MakeOptions(options)
	return _PCRSlice(sequences, opt)
}

// PCRSliceWorker is a worker function builder which produce
// job function usable by the obiseq.MakeISliceWorker function.
func PCRSliceWorker(options ...WithOption) obiseq.SeqSliceWorker {

	opt := MakeOptions(options)
	worker := func(sequences obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {
		result := _PCRSlice(sequences, opt)
		return result, nil
	}

	return worker
}
