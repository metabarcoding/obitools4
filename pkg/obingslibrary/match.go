package obingslibrary

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
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

func (library *NGSLibrary) Compile(maxError int) error {
	for primers, marker := range *library {
		err := marker.Compile(primers.Forward,
			primers.Reverse,
			maxError)
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

func (marker *Marker) Compile(forward, reverse string, maxError int) error {
	var err error
	marker.forward, err = obiapat.MakeApatPattern(forward,
		maxError)
	if err != nil {
		return err
	}
	marker.reverse, err = obiapat.MakeApatPattern(reverse,
		maxError)
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
	match := marker.forward.FindAllIndex(aseq, marker.taglength)

	if len(match) > 0 {
		sseq := sequence.String()
		direct := sseq[match[0][0]:match[0][1]]
		ftag := sseq[(match[0][0] - marker.taglength):match[0][0]]

		m := DemultiplexMatch{
			ForwardMatch:      direct,
			ForwardTag:        ftag,
			BarcodeStart:      match[0][1],
			ForwardMismatches: match[0][2],
			IsDirect:          true,
			Error:             nil,
		}

		rmatch := marker.creverse.FindAllIndex(aseq, match[0][1])

		if len(rmatch) > 0 {

			// extracting primer matches
			reverse, _ := sequence.Subsequence(rmatch[0][0], rmatch[0][1], false)
			defer reverse.Recycle()
			reverse = reverse.ReverseComplement(true)
			rtag, err := sequence.Subsequence(rmatch[0][1], rmatch[0][1]+marker.taglength, false)
			defer rtag.Recycle()
			srtag := ""

			if err != nil {
				rtag = nil
			} else {
				rtag.ReverseComplement(true)
				srtag = strings.ToLower(rtag.String())
			}

			m.ReverseMatch = strings.ToLower(reverse.String())
			m.ReverseMismatches = rmatch[0][2]
			m.BarcodeEnd = rmatch[0][0]
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

	match = marker.reverse.FindAllIndex(aseq, marker.taglength)

	if len(match) > 0 {
		sseq := sequence.String()

		reverse := strings.ToLower(sseq[match[0][0]:match[0][1]])
		rtag := strings.ToLower(sseq[(match[0][0] - marker.taglength):match[0][0]])

		m := DemultiplexMatch{
			ReverseMatch:      reverse,
			ReverseTag:        rtag,
			BarcodeStart:      match[0][1],
			ReverseMismatches: match[0][2],
			IsDirect:          false,
			Error:             nil,
		}

		rmatch := marker.cforward.FindAllIndex(aseq, match[0][1])

		if len(rmatch) > 0 {

			direct, _ := sequence.Subsequence(rmatch[0][0], rmatch[0][1], false)
			defer direct.Recycle()
			direct = direct.ReverseComplement(true)

			ftag, err := sequence.Subsequence(rmatch[0][1], rmatch[0][1]+marker.taglength, false)
			defer ftag.Recycle()
			sftag := ""
			if err != nil {
				ftag = nil

			} else {
				ftag = ftag.ReverseComplement(true)
				sftag = ftag.String()
			}

			m.ForwardMatch = direct.String()
			m.ForwardTag = sftag
			m.ReverseMismatches = rmatch[0][2]
			m.BarcodeEnd = rmatch[0][0]

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

	if match.Error != nil {
		annot["demultiplex_error"] = fmt.Sprintf("%v", match.Error)
	}

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

	return sequence, match.Error
}
