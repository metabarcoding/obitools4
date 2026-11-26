package obingslibrary

import (
	"fmt"
	"slices"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidnapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type Marker struct {
	forward               obidnapat.ApatPattern
	cforward              obidnapat.ApatPattern
	reverse               obidnapat.ApatPattern
	creverse              obidnapat.ApatPattern
	Forward_tag_length    int
	Reverse_tag_length    int
	Forward_spacer        int
	Reverse_spacer        int
	Forward_error         int
	Reverse_error         int
	Forward_allows_indels bool
	Reverse_allows_indels bool
	Forward_matching      string
	Reverse_matching      string
	Forward_tag_delimiter byte
	Reverse_tag_delimiter byte
	Forward_tag_indels    int
	Reverse_tag_indels    int
	// Forward_tag_mismatches and Reverse_tag_mismatches are the largest distance
	// between a tag found and the closest tag of the library for which the tag is
	// kept (matching hamming or indel); a negative value is no limit.
	Forward_tag_mismatches int
	Reverse_tag_mismatches int
	samples                map[TagPair]*PCR
	tags                   atomic.Pointer[tagSets] // the distinct tags of samples, built when needed
}

func (marker *Marker) Compile(forward, reverse string, maxError int, allowsIndel bool) error {
	var err error

	marker.CheckTagLength()

	marker.forward, err = obidnapat.MakeApatPattern(forward, maxError, allowsIndel)
	if err != nil {
		return err
	}
	marker.reverse, err = obidnapat.MakeApatPattern(reverse, maxError, allowsIndel)
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

func (marker *Marker) Compile2(forward, reverse string) error {
	var err error

	marker.CheckTagLength()

	marker.forward, err = obidnapat.MakeApatPattern(
		forward,
		marker.Forward_error,
		marker.Forward_allows_indels)
	if err != nil {
		return err
	}
	marker.reverse, err = obidnapat.MakeApatPattern(
		reverse,
		marker.Reverse_error,
		marker.Reverse_allows_indels)
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
	start, end, nerr, matched := marker.forward.BestMatch(sequence, marker.Forward_tag_length, -1)

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

		start, end, nerr, matched = marker.creverse.BestMatch(sequence, start, -1)

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

	start, end, nerr, matched = marker.reverse.BestMatch(sequence, marker.Reverse_tag_length, -1)

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

		start, end, nerr, matched = marker.cforward.BestMatch(sequence, end, -1)

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
	pair := TagPair{strings.ToLower(forward),
		strings.ToLower(reverse)}
	pcr, ok := marker.samples[pair]

	if ok {
		return pcr, ok
	}

	ipcr := PCR{}
	marker.samples[pair] = &ipcr
	marker.tags.Store(nil) // the distinct tags have to be listed again

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

	maxfl, _, err := obiutils.MaxMap(forward_length)

	if err != nil {
		return err
	}

	if len(forward_length) > 1 {
		others := make([]int, 0)
		for k := range forward_length {
			if k != maxfl {
				others = append(others, k)
			}
		}
		return fmt.Errorf("forward tag length %d is not the same for all the PCRs : %v", maxfl, others)
	}

	maxrl, _, err := obiutils.MaxMap(reverse_length)

	if err != nil {
		return err
	}

	if len(reverse_length) > 1 {
		others := make([]int, 0)
		for k := range reverse_length {
			if k != maxrl {
				others = append(others, k)
			}
		}
		return fmt.Errorf("reverse tag length %d is not the same for all the PCRs : %v", maxrl, others)
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

func normalizeTagDelimiter(delim byte) byte {
	if delim == '0' || delim == 0 {
		delim = 0
	} else {
		if delim >= 'A' && delim <= 'Z' {
			delim = delim - 'A' + 'a'
		}
		if delim != 'a' && delim != 'c' && delim != 'g' && delim != 't' {
			log.Fatalf("invalid reverse tag delimiter: %c, only 'a', 'c', 'g', 't' and '0' are allowed", delim)
		}
	}

	return delim
}

func (marker *Marker) SetForwardTagDelimiter(delim byte) {
	marker.Forward_tag_delimiter = normalizeTagDelimiter(delim)
}

func (marker *Marker) SetReverseTagDelimiter(delim byte) {
	marker.Reverse_tag_delimiter = normalizeTagDelimiter(delim)
}

func (marker *Marker) SetTagDelimiter(delim byte) {
	marker.SetForwardTagDelimiter(delim)
	marker.SetReverseTagDelimiter(delim)
}

func (marker *Marker) SetForwardTagIndels(indels int) {
	marker.Forward_tag_indels = indels
}

func (marker *Marker) SetReverseTagIndels(indels int) {
	marker.Reverse_tag_indels = indels
}

func (marker *Marker) SetTagIndels(indels int) {
	marker.SetForwardTagIndels(indels)
	marker.SetReverseTagIndels(indels)
}

func (marker *Marker) SetForwardTagMismatches(mismatches int) {
	marker.Forward_tag_mismatches = mismatches
}

func (marker *Marker) SetReverseTagMismatches(mismatches int) {
	marker.Reverse_tag_mismatches = mismatches
}

func (marker *Marker) SetTagMismatches(mismatches int) {
	marker.SetForwardTagMismatches(mismatches)
	marker.SetReverseTagMismatches(mismatches)
}

func (marker *Marker) SetForwardAllowedMismatches(allowed_mismatches int) {
	marker.Forward_error = allowed_mismatches
}

func (marker *Marker) SetReverseAllowedMismatches(allowed_mismatches int) {
	marker.Reverse_error = allowed_mismatches
}

func (marker *Marker) SetAllowedMismatch(allowed_mismatches int) {
	marker.SetForwardAllowedMismatches(allowed_mismatches)
	marker.SetReverseAllowedMismatches(allowed_mismatches)
}

func (marker *Marker) SetForwardAllowsIndels(allows_indel bool) {
	marker.Forward_allows_indels = allows_indel
}

func (marker *Marker) SetReverseAllowsIndels(allows_indel bool) {
	marker.Reverse_allows_indels = allows_indel
}

func (marker *Marker) SetAllowsIndel(allows_indel bool) {
	marker.SetForwardAllowsIndels(allows_indel)
	marker.SetReverseAllowsIndels(allows_indel)
}

func (marker *Marker) SetForwardMatching(matching string) error {
	switch matching {
	case "strict", "hamming", "indel": // Valid matching strategies
		marker.Forward_matching = matching

	default:
		return fmt.Errorf("invalid matching : %s", matching)
	}

	return nil
}

func (marker *Marker) SetReverseMatching(matching string) error {
	switch matching {
	case "strict", "hamming", "indel": // Valid matching strategies
		marker.Reverse_matching = matching

	default:
		return fmt.Errorf("invalid matching : %s", matching)
	}

	return nil
}

func (marker *Marker) SetMatching(matching string) error {
	if err := marker.SetForwardMatching(matching); err != nil {
		return err
	}
	if err := marker.SetReverseMatching(matching); err != nil {
		return err
	}
	return nil
}

// tagSets holds the distinct forward tags and the distinct reverse tags of the
// samples of a marker. There are far fewer of them than samples (a 96 well plate
// has 12 tags of each side for 96 samples), and only they are compared to the tags
// observed on the reads.
type tagSets struct {
	forward []string
	reverse []string
}

// distinctTags lists the distinct tags of the samples, building the lists the first
// time they are needed (and again once a sample has been added). It can be called by
// several goroutines at once; they may build the lists twice, which does no harm.
func (marker *Marker) distinctTags() *tagSets {
	if t := marker.tags.Load(); t != nil {
		return t
	}

	forward := make(map[string]struct{})
	reverse := make(map[string]struct{})
	for tags := range marker.samples {
		forward[tags.Forward] = struct{}{}
		reverse[tags.Reverse] = struct{}{}
	}

	t := &tagSets{forward: sortedKeys(forward), reverse: sortedKeys(reverse)}
	marker.tags.Store(t)
	return t
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
