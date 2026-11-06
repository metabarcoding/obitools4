package obidnapat

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// PCR is an in silico PCR: the amplicons that a pair of primers gives on a sequence,
// with the primers on either strand. It is immutable and can be used by several
// goroutines at once.
type PCR struct {
	opt                    PCROptions
	forward, cfwd          ApatPattern // the forward primer, and its reverse complement
	reverse, crev          ApatPattern
	forwardLen, reverseLen int
}

// NewPCR compiles the primers of an in silico PCR.
func NewPCR(setters ...PCROption) (*PCR, error) {
	opt := makePCROptions(setters)
	if err := opt.validate(); err != nil {
		return nil, err
	}

	pcr := &PCR{opt: opt}
	var err error

	if pcr.forward, err = MakeApatPattern(opt.forwardPrimer, opt.forwardErrors, false); err != nil {
		return nil, fmt.Errorf("forward primer: %w", err)
	}
	if pcr.cfwd, err = pcr.forward.ReverseComplement(); err != nil {
		return nil, err
	}
	if pcr.reverse, err = MakeApatPattern(opt.reversePrimer, opt.reverseErrors, false); err != nil {
		return nil, fmt.Errorf("reverse primer: %w", err)
	}
	if pcr.crev, err = pcr.reverse.ReverseComplement(); err != nil {
		return nil, err
	}
	pcr.forwardLen, pcr.reverseLen = pcr.forward.Len(), pcr.reverse.Len()
	return pcr, nil
}

// pair is a window of the first primer and a window of the reverse complement of the
// second one that delimit an amplicon, and the part of the sequence that it covers.
type pair struct {
	first, second [3]int // [start, end, errors]
	from, to      int
}

// find returns the windows of a primer (every start where it matches).
func (pcr *PCR) find(p ApatPattern, seq *obiseq.BioSequence, begin, length int) [][3]int {
	if pcr.opt.circular {
		return p.AppendAllIndexCircular(nil, seq, begin, length)
	}
	return p.AppendAllIndex(nil, seq, begin, length)
}

// pairs lists the amplicons that the first primer, found on the sequence, and the
// reverse complement of the second one, found after it, delimit.
func (pcr *PCR) pairs(seq *obiseq.BioSequence, first, secondRC ApatPattern) []pair {
	opt, n := pcr.opt, seq.Len()

	firsts := pcr.find(first, seq, 0, -1)
	if len(firsts) == 0 {
		return nil
	}

	// A linear sequence: the second primer is looked for from the first window on,
	// and no farther than the longest amplicon allows. This only saves time.
	begin, length := firsts[0][0], n-firsts[0][0]
	if opt.maxLength > 0 {
		length = firsts[len(firsts)-1][1] - begin + opt.maxLength + secondRC.Len()
	}
	if opt.circular {
		begin, length = 0, -1
	}
	seconds := pcr.find(secondRC, seq, begin, length)

	var res []pair
	for _, fm := range firsts {
		for _, rm := range seconds {
			var between int // the length of the amplicon, the primers excluded
			switch {
			case rm[1] > fm[0]:
				between = rm[0] - fm[1]
			case opt.circular: // the second primer lies before the first one: across the junction
				between = rm[0] + n - fm[1]
			}
			// (a length of 0 or less: the primers touch or overlap)
			if between <= 0 || opt.minLength > 0 && between < opt.minLength ||
				opt.maxLength > 0 && between > opt.maxLength {
				continue
			}

			from, to := fm[1], rm[0]
			if opt.hasExtension() {
				from, to = fm[0]-opt.extension, rm[1]+opt.extension
				if !opt.fullExtension && !opt.circular {
					from, to = max(from, 0), min(to, n)
				}
				if !opt.circular && (from < 0 || to > n) {
					continue
				}
				if opt.circular && between+(fm[1]-fm[0])+(rm[1]-rm[0])+2*opt.extension > n {
					continue // the amplicon with its flanks would be longer than the molecule
				}
				if opt.circular && from < 0 { // the flank crosses the junction: Subsequence wants a start in the sequence
					shift := (-from + n - 1) / n * n
					from, to = from+shift, to+shift
				}
			}
			res = append(res, pair{first: fm, second: rm, from: from, to: to})
		}
	}
	return res
}

// window returns the bases of a window of the sequence, in upper or lower case as in
// the sequence.
func (pcr *PCR) window(seq *obiseq.BioSequence, w [3]int, reverseComplement bool) string {
	s, err := seq.Subsequence(w[0], w[1], pcr.opt.circular)
	if err != nil || s == nil {
		log.Fatalf("error in extracting %d..%d from %s (circular %v): %v", w[0], w[1], seq.Id(), pcr.opt.circular, err)
	}
	if reverseComplement {
		s = s.ReverseComplement(true)
	}
	text := s.String()
	s.Recycle()
	return text
}

// amplicons appends the amplicons of one orientation of the PCR: the first primer on
// the strand of the sequence and the second one on the other strand. In the
// orientation "forward" the first primer is the forward one; in "reverse" it is the
// reverse one and the amplicon is reported reverse complemented, so that it always
// reads from the forward primer to the reverse primer.
func (pcr *PCR) amplicons(results obiseq.BioSequenceSlice, seq *obiseq.BioSequence, forwardOrientation bool) obiseq.BioSequenceSlice {
	first, secondRC := pcr.forward, pcr.crev
	if !forwardOrientation {
		first, secondRC = pcr.reverse, pcr.cfwd
	}

	for _, p := range pcr.pairs(seq, first, secondRC) {
		amplicon, err := seq.Subsequence(p.from, p.to, pcr.opt.circular)
		if err != nil {
			log.Fatalf("error : %v\n", err)
		}
		if !forwardOrientation {
			amplicon = amplicon.ReverseComplement(true)
		}

		annot := amplicon.Annotations()
		obiutils.MustFillMap(annot, seq.Annotations())

		annot["forward_primer"] = pcr.forward.String()
		annot["reverse_primer"] = pcr.reverse.String()
		if forwardOrientation {
			annot["forward_match"] = pcr.window(seq, p.first, false)
			annot["forward_error"] = p.first[2]
			annot["reverse_match"] = pcr.window(seq, p.second, true)
			annot["reverse_error"] = p.second[2]
			annot["direction"] = "forward"
		} else {
			annot["forward_match"] = pcr.window(seq, p.second, true)
			annot["forward_error"] = p.second[2]
			annot["reverse_match"] = pcr.window(seq, p.first, false)
			annot["reverse_error"] = p.first[2]
			annot["direction"] = "reverse"
		}

		results = append(results, amplicon)
	}
	return results
}

// Amplicons returns the amplicons that the primers give on the sequence, those of the
// forward primer then those of the reverse primer.
func (pcr *PCR) Amplicons(seq *obiseq.BioSequence) obiseq.BioSequenceSlice {
	results := make(obiseq.BioSequenceSlice, 0, 10)
	results = pcr.amplicons(results, seq, true)
	return pcr.amplicons(results, seq, false)
}

// Slice returns the amplicons of every sequence of the slice.
func (pcr *PCR) Slice(sequences obiseq.BioSequenceSlice) obiseq.BioSequenceSlice {
	results := make(obiseq.BioSequenceSlice, 0, len(sequences))
	for _, seq := range sequences {
		results = append(results, pcr.Amplicons(seq)...)
	}
	return results
}

// SliceWorker is the worker of the PCR for obiseq.MakeISliceWorker.
func (pcr *PCR) SliceWorker() obiseq.SeqSliceWorker {
	return func(sequences obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {
		return pcr.Slice(sequences), nil
	}
}
