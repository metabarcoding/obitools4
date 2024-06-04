package obingslibrary

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
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
	for primers, marker := range library.Markers {
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
	for primers, marker := range library.Markers {
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

// ExtractBarcode extracts the barcode from the given biosequence.
//
// Parameters:
// - sequence: The biosequence from which to extract the barcode.
// - inplace: A boolean indicating whether the barcode should be extracted in-place or not.
//
// Returns:
// - The biosequence with the extracted barcode.
// - An error indicating any issues encountered during the extraction process.
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
		annot["forward_error"] = match.ForwardMismatches
		annot["forward_tag"] = match.ForwardTag
	}

	if match.ReverseMatch != "" {
		annot["reverse_match"] = match.ReverseMatch
		annot["reverse_error"] = match.ReverseMismatches
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
