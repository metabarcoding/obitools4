package obingslibrary

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

type DemultiplexMatch struct {
	ForwardMatch      string
	ReverseMatch      string
	ForwardTag        string
	ReverseTag        string
	BarcodeStart      int
	BarcodeEnd        int
	ForwardMismatches int
	ReverseMismatches int
	IsDirect          bool
	Pcr               *PCR
	ForwardPrimer     string
	ReversePrimer     string
	Error             error
}

func (library *NGSLibrary) Compile(maxError int, allowsIndel bool) error {
	for primers, marker := range *library {
		err := marker.Compile(primers.Forward,
			primers.Reverse,
			maxError, allowsIndel)
		if err != nil {
			return err
		}
	}
	return nil
}

func (library *NGSLibrary) Match(sequence *obiseq.BioSequence) *DemultiplexMatch {
	for primers, marker := range *library {
		m := marker.Match(sequence)
		if m != nil {
			m.ForwardPrimer = strings.ToLower(primers.Forward)
			m.ReversePrimer = strings.ToLower(primers.Reverse)
			return m
		}
	}
	return nil
}

func (library *NGSLibrary) ExtractBarcode(sequence *obiseq.BioSequence, inplace bool) (*obiseq.BioSequence, error) {
	match := library.Match(sequence)

	return match.ExtractBarcode(sequence, inplace)
}

func (marker *Marker) Compile(forward, reverse string, maxError int, allowsIndel bool) error {
	var err error
	marker.forward, err = obiapat.MakeApatPattern(forward, maxError, allowsIndel)
	if err != nil {
		return err
	}
	marker.reverse, err = obiapat.MakeApatPattern(reverse, maxError, allowsIndel)
	if err != nil {
		return err
	}

	marker.cforward, err = marker.forward.ReverseComplement()
	if err != nil {
		return err
	}
	marker.creverse, err = marker.reverse.ReverseComplement()
	if err != nil {
		return err
	}

	marker.taglength = 0
	for tags := range marker.samples {
		lf := len(tags.Forward)
		lr := len(tags.Reverse)

		l := lf
		if lf == 0 {
			l = lr
		}

		if lr != 0 && l != lr {
			return fmt.Errorf("forward tag (%s) and reverse tag (%s) do not have the same length",
				tags.Forward, tags.Reverse)
		}

		if marker.taglength != 0 && l != marker.taglength {
			return fmt.Errorf("tag pair (%s,%s) is not compatible with a tag length of %d",
				tags.Forward, tags.Reverse, marker.taglength)
		} else {
			marker.taglength = l
		}
	}

	return nil
}

func (marker *Marker) Match(sequence *obiseq.BioSequence) *DemultiplexMatch {
	aseq, _ := obiapat.MakeApatSequence(sequence, false)

	start, end, nerr, matched := marker.forward.BestMatch(aseq, marker.taglength, -1)
	if matched {
		sseq := sequence.String()
		direct := sseq[start:end]
		tagstart := obiutils.MaxInt(start-marker.taglength, 0)
		ftag := strings.ToLower(sseq[tagstart:start])

		m := DemultiplexMatch{
			ForwardMatch:      direct,
			ForwardTag:        ftag,
			BarcodeStart:      end,
			ForwardMismatches: nerr,
			IsDirect:          true,
			Error:             nil,
		}

		start, end, nerr, matched = marker.creverse.BestMatch(aseq, start, -1)

		if matched {

			// extracting primer matches
			reverse, _ := sequence.Subsequence(start, end, false)
			defer reverse.Recycle()
			reverse = reverse.ReverseComplement(true)
			endtag := obiutils.MinInt(end+marker.taglength, sequence.Len())
			rtag, err := sequence.Subsequence(end, endtag, false)
			defer rtag.Recycle()
			srtag := ""

			if err != nil {
				rtag = nil
			} else {
				rtag.ReverseComplement(true)
				srtag = strings.ToLower(rtag.String())
			}

			m.ReverseMatch = strings.ToLower(reverse.String())
			m.ReverseMismatches = nerr
			m.BarcodeEnd = start
			m.ReverseTag = srtag

			sample, ok := marker.samples[TagPair{ftag, srtag}]

			if ok {
				m.Pcr = sample
			}

			return &m

		}

		m.Error = fmt.Errorf("cannot locates reverse priming site")

		return &m
	}

	start, end, nerr, matched = marker.reverse.BestMatch(aseq, marker.taglength, -1)

	if matched {
		sseq := sequence.String()

		reverse := strings.ToLower(sseq[start:end])
		tagstart := obiutils.MaxInt(start-marker.taglength, 0)
		rtag := strings.ToLower(sseq[tagstart:start])

		m := DemultiplexMatch{
			ReverseMatch:      reverse,
			ReverseTag:        rtag,
			BarcodeStart:      end,
			ReverseMismatches: nerr,
			IsDirect:          false,
			Error:             nil,
		}

		start, end, nerr, matched = marker.cforward.BestMatch(aseq, end, -1)

		if matched {

			direct, _ := sequence.Subsequence(start, end, false)
			defer direct.Recycle()
			direct = direct.ReverseComplement(true)

			endtag := obiutils.MinInt(end+marker.taglength, sequence.Len())
			ftag, err := sequence.Subsequence(end, endtag, false)
			defer ftag.Recycle()
			sftag := ""
			if err != nil {
				ftag = nil

			} else {
				ftag = ftag.ReverseComplement(true)
				sftag = strings.ToLower(ftag.String())
			}

			m.ForwardMatch = direct.String()
			m.ForwardTag = sftag
			m.ForwardMismatches = nerr
			m.BarcodeEnd = start

			sample, ok := marker.samples[TagPair{sftag, rtag}]

			if ok {
				m.Pcr = sample
			}

			return &m
		}

		m.Error = fmt.Errorf("cannot locates forward priming site")

		return &m
	}

	return nil
}

func (match *DemultiplexMatch) ExtractBarcode(sequence *obiseq.BioSequence, inplace bool) (*obiseq.BioSequence, error) {
	if !inplace {
		sequence = sequence.Copy()
	}

	if match == nil {
		annot := sequence.Annotations()
		annot["demultiplex_error"] = "cannot match any primer pair"
		return sequence, errors.New("cannot match any primer pair")
	}

	if match.ForwardMatch != "" && match.ReverseMatch != "" {
		var err error

		if match.BarcodeStart < match.BarcodeEnd {
			sequence, err = sequence.Subsequence(match.BarcodeStart, match.BarcodeEnd, false)
			if err != nil {
				log.Fatalf("cannot extract sub sequence %d..%d %v", match.BarcodeStart, match.BarcodeEnd, *match)
			}
		} else {
			annot := sequence.Annotations()
			annot["demultiplex_error"] = "read correponding to a primer dimer"
			return sequence, errors.New("read correponding to a primer dimer")
		}
	}

	if !match.IsDirect {
		sequence.ReverseComplement(true)
	}

	annot := sequence.Annotations()

	if annot == nil {
		log.Fatalf("nil annot %v", sequence)
	}
	annot["forward_primer"] = match.ForwardPrimer
	annot["reverse_primer"] = match.ReversePrimer

	if match.IsDirect {
		annot["direction"] = "direct"
	} else {
		annot["direction"] = "reverse"
	}

	if match.ForwardMatch != "" {
		annot["forward_match"] = match.ForwardMatch
		annot["forward_mismatches"] = match.ForwardMismatches
		annot["forward_tag"] = match.ForwardTag
	}

	if match.ReverseMatch != "" {
		annot["reverse_match"] = match.ReverseMatch
		annot["reverse_mismatches"] = match.ReverseMismatches
		annot["reverse_tag"] = match.ReverseTag
	}

	if match.Error == nil {
		if match.Pcr != nil {
			annot["sample"] = match.Pcr.Sample
			annot["experiment"] = match.Pcr.Experiment
			for k, val := range match.Pcr.Annotations {
				annot[k] = val
			}
		} else {
			annot["demultiplex_error"] = "cannot assign the sequence to a sample"
			match.Error = errors.New("cannot assign the sequence to a sample")
		}
	} else {
		annot["demultiplex_error"] = fmt.Sprintf("%v", match.Error)
	}

	return sequence, match.Error
}
