package obingslibrary

import (
	"fmt"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type Marker struct {
	forward               obiapat.ApatPattern
	cforward              obiapat.ApatPattern
	reverse               obiapat.ApatPattern
	creverse              obiapat.ApatPattern
	Forward_tag_length    int
	Reverse_tag_length    int
	Forward_spacer        int
	Reverse_spacer        int
	Forward_tag_delimiter byte
	Reverse_tag_delimiter byte
	samples               map[TagPair]*PCR
}

func (marker *Marker) Compile(forward, reverse string, maxError int, allowsIndel bool) error {
	var err error

	marker.CheckTagLength()

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
	return nil
}

// Match finds the best matching demultiplex for a given sequence.
//
// Parameters:
//
//	marker - a pointer to a Marker struct that contains the forward and reverse primers.
//	sequence - a pointer to a BioSequence struct that represents the input sequence.
//
// Returns:
//
//	A pointer to a DemultiplexMatch struct that contains the best matching demultiplex.
func (marker *Marker) Match(sequence *obiseq.BioSequence) *DemultiplexMatch {
	aseq, _ := obiapat.MakeApatSequence(sequence, false)

	start, end, nerr, matched := marker.forward.BestMatch(aseq, marker.Forward_tag_length, -1)

	if matched {
		if start < 0 {
			start = 0
		}

		if end > sequence.Len() {
			end = sequence.Len()
		}

		sseq := sequence.String()
		direct := sseq[start:end]
		tagstart := max(start-marker.Forward_tag_length, 0)
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
			endtag := min(end+marker.Reverse_tag_length, sequence.Len())
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

	//
	// At this point the forward primer didn't match.
	// We try now with the reverse primer
	//

	start, end, nerr, matched = marker.reverse.BestMatch(aseq, marker.Reverse_tag_length, -1)

	if matched {
		if start < 0 {
			start = 0
		}

		if end > sequence.Len() {
			end = sequence.Len()
		}

		sseq := sequence.String()

		reverse := strings.ToLower(sseq[start:end])
		tagstart := max(start-marker.Reverse_tag_length, 0)
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

			endtag := min(end+marker.Forward_tag_length, sequence.Len())
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

func (marker *Marker) GetPCR(forward, reverse string) (*PCR, bool) {
	pair := TagPair{forward, reverse}
	pcr, ok := marker.samples[pair]

	if ok {
		return pcr, ok
	}

	ipcr := PCR{}
	marker.samples[pair] = &ipcr

	return &ipcr, false
}

func (marker *Marker) CheckTagLength() error {
	forward_length := make(map[int]int)
	reverse_length := make(map[int]int)

	marker.Forward_tag_length = -1
	marker.Reverse_tag_length = -1

	for tags := range marker.samples {
		forward_length[len(tags.Forward)]++
		reverse_length[len(tags.Reverse)]++
	}

	maxfl, _ := obiutils.MaxMap(forward_length)

	if len(forward_length) > 1 {
		others := make([]int, 0)
		for k := range forward_length {
			if k != maxfl {
				others = append(others, k)
			}
		}
		return fmt.Errorf("forward tag length %d is not the same for all the PCRs : %v\n", maxfl, others)
	}

	maxrl, _ := obiutils.MaxMap(reverse_length)

	if len(reverse_length) > 1 {
		others := make([]int, 0)
		for k := range reverse_length {
			if k != maxrl {
				others = append(others, k)
			}
		}
		return fmt.Errorf("reverse tag length %d is not the same for all the PCRs : %v\n", maxrl, others)
	}

	marker.Forward_tag_length = maxfl
	marker.Reverse_tag_length = maxrl
	return nil
}

func (marker *Marker) SetForwardTagSpacer(spacer int) {
	marker.Forward_spacer = spacer
}

func (marker *Marker) SetReverseTagSpacer(spacer int) {
	marker.Reverse_spacer = spacer
}

func (marker *Marker) SetTagSpacer(spacer int) {
	marker.SetForwardTagSpacer(spacer)
	marker.SetReverseTagSpacer(spacer)
}

func (marker *Marker) SetForwardTagDelimiter(delim byte) {
	marker.Forward_tag_delimiter = delim
}

func (marker *Marker) SetReverseTagDelimiter(delim byte) {
	marker.Reverse_tag_delimiter = delim
}

func (marker *Marker) SetTagDelimiter(delim byte) {
	marker.SetForwardTagDelimiter(delim)
	marker.SetReverseTagDelimiter(delim)
}
