package obingslibrary

import (
	"fmt"
	"math"
	"slices"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type PrimerMatch struct {
	Begin      int
	End        int
	Mismatches int
	Marker     int
	Forward    bool
}

type TagMatcher func(
	sequence *obiseq.BioSequence,
	begin, end int, forward bool) (TagPair, error)

func Hamming(a, b string) int {

	if len(a) != len(b) {
		return max(len(a), len(b))
	}

	count := int(0)

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			count++
		}
	}

	return count
}

func Levenshtein(s1, s2 string) int {
	lenS1, lenS2 := len(s1), len(s2)
	if lenS1 == 0 {
		return lenS2
	}
	if lenS2 == 0 {
		return lenS1
	}

	// Create two slices to store the distances
	prev := make([]int, lenS2+1)
	curr := make([]int, lenS2+1)

	// Initialize the previous row of the matrix
	for j := 0; j <= lenS2; j++ {
		prev[j] = j
	}

	// Iterate over each character in s1
	for i := 1; i <= lenS1; i++ {
		curr[0] = i

		// Iterate over each character in s2
		for j := 1; j <= lenS2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			// Calculate the minimum cost for the current cell
			curr[j] = min(prev[j]+1,
				curr[j-1]+1,    // Insertion
				prev[j-1]+cost) // Substitution
		}
		// Copy current row to previous row for the next iteration
		prev, curr = curr, prev
	}

	// The last value in the previous row is the Levenshtein distance
	return prev[lenS2]
}

func lookForTag(seq string, delimiter byte) string {

	i := len(seq) - 1

	// obilog.Warnf("Provided fragment : %s", string(seq))

	for i >= 0 && seq[i] != delimiter {
		i--
	}

	for i >= 0 && seq[i] == delimiter {
		i--
	}

	end := i + 1

	for i >= 0 && seq[i] != delimiter {
		i--
	}

	begin := i + 1

	if i < 0 {
		return ""
	}

	// obilog.Warnf("extracted : %s", string(seq[begin:end]))
	return seq[begin:end]
}

func lookForRescueTag(seq string, delimiter byte, taglength, border, indel int) string {
	// log.Info("lookForRescueTag")
	// log.Infof("seq: %s", seq)

	i := len(seq) - 1

	// Skip the border part not corresponding to the tag delimiter
	for i >= 0 && seq[i] != delimiter {
		i--
	}

	delimlen := 0
	for i >= 0 && seq[i] == delimiter {
		i--
		delimlen++
	}

	if (border - delimlen) > indel {
		return ""
	}

	if delimlen > border {
		i += delimlen - border
	}

	// log.Infof("delimlen: %d", delimlen)

	end := i + 1

	i -= taglength - indel

	for i >= 0 && seq[i] != delimiter {
		i--
	}

	delimlen = 0
	for i >= 0 && seq[i] == delimiter {
		i--
		delimlen++
	}

	delimlen = min(delimlen, border)

	// log.Infof("delimlen: %d", delimlen)

	begin := i + delimlen + 1

	if i < 0 || obiutils.Abs(taglength-end+begin) > indel {
		return ""
	}

	// log.Infof("begin: %d, end: %d", begin, end)
	// log.Infof("seq[begin:end]: %s", seq[begin:end])

	return seq[begin:end]
}

func (marker *Marker) beginDelimitedTagExtractor(
	sequence *obiseq.BioSequence,
	begin int,
	forward bool) string {

	// log.Warn("beginDelimitedTagExtractor")
	taglength := 2*marker.Forward_spacer + marker.Forward_tag_length
	delimiter := marker.Forward_tag_delimiter

	if !forward {
		taglength = 2*marker.Reverse_spacer + marker.Reverse_tag_length
		delimiter = marker.Reverse_tag_delimiter
	}

	fb := begin - taglength*2
	if fb < 0 {
		fb = 0
	}

	return lookForTag(sequence.String()[fb:begin], delimiter)
}

func (marker *Marker) beginRescueTagExtractor(
	sequence *obiseq.BioSequence,
	begin int,
	forward bool) string {

	delimiter := marker.Forward_tag_delimiter
	border := marker.Forward_spacer
	taglength := marker.Forward_tag_length
	delta := marker.Forward_tag_indels

	if !forward {
		taglength = marker.Reverse_tag_length
		border = marker.Reverse_spacer
		delimiter = marker.Reverse_tag_delimiter
		delta = marker.Reverse_tag_indels
	}

	frglength := border + taglength

	fb := begin - frglength*2
	if fb < 0 {
		fb = 0
	}

	return lookForRescueTag(sequence.String()[fb:begin], delimiter, taglength, border, delta)
}

func (marker *Marker) beginFixedTagExtractor(
	sequence *obiseq.BioSequence,
	begin int,
	forward bool) string {

	taglength := marker.Forward_tag_length
	spacer := marker.Forward_spacer

	if !forward {
		taglength = marker.Reverse_tag_length
		spacer = marker.Reverse_spacer
	}

	fb := begin - spacer - taglength
	if fb < 0 {
		return ""
	}

	return sequence.String()[fb:(begin - spacer)]
}

func (marker *Marker) endDelimitedTagExtractor(
	sequence *obiseq.BioSequence,
	end int,
	forward bool) string {
	// log.Warn("endDelimitedTagExtractor")

	taglength := marker.Reverse_spacer + marker.Reverse_tag_length
	delimiter := marker.Reverse_tag_delimiter

	if !forward {
		taglength = marker.Forward_spacer + marker.Forward_tag_length
		delimiter = marker.Forward_tag_delimiter
	}

	fb := end + taglength*2
	if fb > sequence.Len() {
		fb = sequence.Len()
	}

	if end >= fb {
		return ""
	}

	tag_seq, err := sequence.Subsequence(end, fb, false)

	if err != nil {
		log.Fatalf("Cannot extract sequence tag : %v", err)
	}

	return lookForTag(tag_seq.ReverseComplement(true).String(), delimiter)
}

func (marker *Marker) endRescueTagExtractor(
	sequence *obiseq.BioSequence,
	end int,
	forward bool) string {

	delimiter := marker.Reverse_tag_delimiter
	border := marker.Reverse_spacer
	taglength := marker.Reverse_tag_length
	delta := marker.Reverse_tag_indels

	if !forward {
		taglength = marker.Forward_tag_length
		border = marker.Forward_spacer
		delimiter = marker.Forward_tag_delimiter
		delta = marker.Forward_tag_indels
	}

	frglength := border + taglength

	fb := end + frglength*2

	if fb > sequence.Len() {
		fb = sequence.Len()
	}

	if end >= fb {
		return ""
	}

	tag_seq, err := sequence.Subsequence(end, fb, false)

	if err != nil {
		log.Fatalf("Cannot extract sequence tag : %v", err)
	}

	return lookForRescueTag(tag_seq.ReverseComplement(true).String(), delimiter, taglength, border, delta)
}
func (marker *Marker) endFixedTagExtractor(
	sequence *obiseq.BioSequence,
	end int,
	forward bool) string {

	taglength := marker.Reverse_tag_length
	spacer := marker.Reverse_spacer

	if !forward {
		taglength = marker.Forward_tag_length
		spacer = marker.Forward_spacer
	}

	fe := end + spacer + taglength
	if fe > sequence.Len() {
		return ""
	}

	tag_seq, err := sequence.Subsequence(end+spacer, fe, false)

	if err != nil {
		log.Fatalf("Cannot extract sequence tag : %v", err)
	}

	return tag_seq.ReverseComplement(true).String()
}

func (marker *Marker) beginTagExtractor(
	sequence *obiseq.BioSequence,
	begin int,
	forward bool) string {
	// obilog.Warnf("Forward : %v -> %d %c", forward, marker.Forward_spacer, marker.Forward_tag_delimiter)
	// obilog.Warnf("Forward : %v -> %d %c", forward, marker.Reverse_spacer, marker.Reverse_tag_delimiter)
	if forward {
		if marker.Forward_tag_length == 0 {
			return ""
		}

		if marker.Forward_tag_delimiter == 0 {
			return marker.beginFixedTagExtractor(sequence, begin, forward)
		} else {
			if marker.Forward_tag_indels == 0 {
				// obilog.Warnf("Delimited tag for forward primers %s", marker.forward.String())
				return marker.beginDelimitedTagExtractor(sequence, begin, forward)
			} else {
				// obilog.Warnf("Rescue tag for forward primers %s", marker.forward.String())
				return marker.beginRescueTagExtractor(sequence, begin, forward)
			}
		}
	} else {
		if marker.Reverse_tag_length == 0 {
			return ""
		}

		if marker.Reverse_tag_delimiter == 0 {
			return marker.beginFixedTagExtractor(sequence, begin, forward)
		} else {
			if marker.Reverse_tag_indels == 0 {
				// obilog.Warnf("Delimited tag for reverse/complement primers %s", marker.creverse.String())
				return marker.beginDelimitedTagExtractor(sequence, begin, forward)
			} else {
				// obilog.Warnf("Rescue tag for reverse/complement primers %s", marker.creverse.String())
				return marker.beginRescueTagExtractor(sequence, begin, forward)
			}
		}
	}
}

func (marker *Marker) endTagExtractor(
	sequence *obiseq.BioSequence,
	end int,
	forward bool) string {
	if forward {
		if marker.Reverse_tag_length == 0 {
			return ""
		}

		if marker.Reverse_tag_delimiter == 0 {
			return marker.endFixedTagExtractor(sequence, end, forward)
		} else {
			if marker.Reverse_tag_indels == 0 {
				// obilog.Warnf("Delimited tag for reverse primers %s", marker.reverse.String())
				return marker.endDelimitedTagExtractor(sequence, end, forward)
			} else {
				// obilog.Warnf("Rescue tag for reverse primers %s", marker.reverse.String())
				return marker.endRescueTagExtractor(sequence, end, forward)
			}
		}
	} else {
		if marker.Forward_tag_length == 0 {
			return ""
		}

		if marker.Forward_tag_delimiter == 0 {
			return marker.endFixedTagExtractor(sequence, end, forward)
		} else {
			if marker.Forward_tag_indels == 0 {
				// obilog.Warnf("Delimited tag for forward/complement primers %s", marker.cforward.String())
				return marker.endDelimitedTagExtractor(sequence, end, forward)
			} else {
				// obilog.Warnf("Rescue tag for forward/complement primers %s", marker.cforward.String())
				return marker.endRescueTagExtractor(sequence, end, forward)
			}
		}
	}
}

func (library *NGSLibrary) TagExtractor(
	sequence *obiseq.BioSequence,
	annotations obiseq.Annotation,
	primers PrimerPair,
	begin, end int,
	forward bool) *TagPair {

	marker, ok := library.Markers[primers]

	if !ok {
		log.Fatalf("marker not found : %v", primers)
	}

	forward_tag := marker.beginTagExtractor(sequence, begin, forward)
	reverse_tag := marker.endTagExtractor(sequence, end, forward)

	if !forward {
		forward_tag, reverse_tag = reverse_tag, forward_tag
	}

	if forward_tag != "" {
		annotations["obimultiplex_forward_tag"] = forward_tag
	}

	if reverse_tag != "" {
		annotations["obimultiplex_reverse_tag"] = reverse_tag
	}

	return &TagPair{forward_tag, reverse_tag}
}

func (marker *Marker) ClosestForwardTag(
	tag string,
	dist func(string, string) int,
) (string, int) {
	mindist := math.MaxInt
	mintag := ""

	for ts := range marker.samples {
		d := dist(ts.Forward, tag)

		if d == mindist && mintag != "" && ts.Forward != mintag {
			mintag = ""
		}

		if d < mindist {
			mindist = d
			mintag = ts.Forward
		}
	}

	return mintag, mindist
}

func (marker *Marker) ClosestReverseTag(
	tag string,
	dist func(string, string) int,
) (string, int) {
	mindist := math.MaxInt
	mintag := ""

	for ts := range marker.samples {
		d := dist(ts.Reverse, tag)

		if d == mindist && mintag != "" && ts.Reverse != mintag {
			mintag = ""
		}

		if d < mindist {
			mindist = d
			mintag = ts.Reverse
		}
	}

	return mintag, mindist
}

func (library *NGSLibrary) SampleIdentifier(
	primerseqs PrimerPair,
	tags *TagPair,
	annotations obiseq.Annotation) *PCR {

	marker, ok := library.Markers[primerseqs]

	if !ok {
		log.Fatalf("marker not found : %v", primerseqs)
	}

	forward := ""
	reverse := ""

	fdistance := int(0)
	rdistance := int(0)

	if tags.Forward != "" {
		switch marker.Forward_matching {
		case "strict":
			forward = tags.Forward
			fdistance = 0
			annotations["obimultiplex_forward_matching"] = "strict"
		case "hamming":
			forward, fdistance = marker.ClosestForwardTag(tags.Forward, Hamming)
			annotations["obimultiplex_forward_matching"] = "hamming"
		case "indel":
			forward, fdistance = marker.ClosestForwardTag(tags.Forward, Levenshtein)
			annotations["obimultiplex_forward_matching"] = "indel"
		}
		annotations["obimultiplex_forward_tag_dist"] = fdistance
		annotations["obimultiplex_forward_proposed_tag"] = forward
	}

	if tags.Reverse != "" {
		switch marker.Reverse_matching {
		case "strict":
			reverse = tags.Reverse
			rdistance = 0
			annotations["obimultiplex_reverse_matching"] = "strict"
		case "hamming":
			reverse, rdistance = marker.ClosestReverseTag(tags.Reverse, Hamming)
			annotations["obimultiplex_reverse_matching"] = "hamming"
		case "indel":
			reverse, rdistance = marker.ClosestReverseTag(tags.Reverse, Levenshtein)
			annotations["obimultiplex_reverse_matching"] = "indel"
		}
		annotations["obimultiplex_reverse_tag_dist"] = rdistance
		annotations["obimultiplex_reverse_proposed_tag"] = reverse
	}

	proposed := TagPair{forward, reverse}

	pcr, ok := marker.samples[proposed]

	if !ok {
		annotations["obimultiplex_error"] = fmt.Sprintf("Cannot associate sample to the tag pair (%s:%s)", forward, reverse)
		return nil
	}

	annotations["sample"] = pcr.Sample
	annotations["experiment"] = pcr.Experiment
	for k, v := range pcr.Annotations {
		annotations[k] = v
	}

	return pcr
}

func (library *NGSLibrary) ExtractMultiBarcode(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
	i := 1
	markers := make([]*Marker, len(library.Markers)+1)
	primerseqs := make([]PrimerPair, len(library.Markers)+1)
	matches := make([]PrimerMatch, 0, len(library.Markers)+1)
	aseq, err := obiapat.MakeApatSequence(sequence, false)

	results := obiseq.MakeBioSequenceSlice()

	if err != nil {
		log.Fatalf("error in building apat sequence : %v\n", err)
	}

	for primers, marker := range library.Markers {
		markers[i] = marker
		primerseqs[i] = primers
		locs := marker.forward.AllMatches(aseq, 0, -1)
		if len(locs) > 0 {
			for _, loc := range locs {
				matches = append(matches, PrimerMatch{
					Begin:      loc[0],
					End:        loc[1],
					Mismatches: loc[2],
					Marker:     i,
					Forward:    true,
				})

			}

			locs = marker.creverse.AllMatches(aseq, locs[0][0]+1, -1)

			if len(locs) > 0 {
				for _, loc := range locs {
					matches = append(matches, PrimerMatch{
						Begin:      loc[0],
						End:        loc[1],
						Mismatches: loc[2],
						Marker:     -i,
						Forward:    true,
					})
				}
			}
		}

		locs = marker.reverse.AllMatches(aseq, 0, -1)
		if len(locs) > 0 {
			for _, loc := range locs {
				matches = append(matches, PrimerMatch{
					Begin:      loc[0],
					End:        loc[1],
					Mismatches: loc[2],
					Marker:     i,
					Forward:    false,
				})
			}

			locs = marker.cforward.AllMatches(aseq, locs[0][0]+1, -1)

			if len(locs) > 0 {
				for _, loc := range locs {
					matches = append(matches, PrimerMatch{
						Begin:      loc[0],
						End:        loc[1],
						Mismatches: loc[2],
						Marker:     -i,
						Forward:    false,
					})
				}
			}
		}
		i++
	}

	if len(matches) > 0 {
		slices.SortFunc(matches, func(a, b PrimerMatch) int { return a.Begin - b.Begin })

		state := 0
		var from PrimerMatch
		q := 0
		for _, match := range matches {

			switch state {
			case 0:
				if match.Marker > 0 {
					from = match
					state = 1
				}

			case 1:
				if match.Marker == -from.Marker && match.Forward == from.Forward {
					barcode_error := false
					annotations := obiseq.GetAnnotation()
					annotations["obimultiplex_forward_primer"] = primerseqs[from.Marker].Forward
					annotations["obimultiplex_reverse_primer"] = primerseqs[from.Marker].Reverse

					if from.Forward {
						// With have a barcode in the orientation from the forward primer to the reverse

						// Try to extract the forward primer match
						if from.Begin < 0 || from.End > sequence.Len() {
							barcode_error = true
							annotations["obimultiplex_error"] = "Cannot extract forward primer match"
						} else {
							annotations["obimultiplex_forward_match"] = sequence.String()[from.Begin:from.End]
						}

						// Try to extract the reverse primer match
						sseq, err := sequence.Subsequence(match.Begin, match.End, false)

						if err != nil {
							barcode_error = true
							annotations["obimultiplex_error"] = "Cannot extract reverse primer match"
						} else {
							annotations["obimultiplex_reverse_match"] = sseq.ReverseComplement(true).String()
						}

						annotations["obimultiplex_forward_error"] = from.Mismatches
						annotations["obimultiplex_reverse_error"] = match.Mismatches
					} else {
						// With have a barcode in the orientation from the reverse primer to the forward

						// Try to extract the reverse primer match
						if from.Begin < 0 || from.End > sequence.Len() {
							barcode_error = true
							annotations["obimultiplex_error"] = "Cannot extract reverse primer match"
						} else {
							annotations["obimultiplex_reverse_match"] = sequence.String()[from.Begin:from.End]
						}

						// Try to extract the forward primer match
						sseq, err := sequence.Subsequence(match.Begin, match.End, false)

						if err != nil {
							barcode_error = true
							annotations["obimultiplex_error"] = "Cannot extract forward primer match"
						} else {
							annotations["obimultiplex_forward_match"] = sseq.ReverseComplement(true).String()
						}

						annotations["obimultiplex_reverse_error"] = from.Mismatches
						annotations["obimultiplex_forward_error"] = match.Mismatches
					}

					// if we were  able to extract the primer matches we can extract the barcode
					if !barcode_error {
						tags := library.TagExtractor(sequence, annotations, primerseqs[from.Marker], from.Begin, match.End, from.Forward)

						barcode, err := sequence.Subsequence(from.End, match.Begin, false)

						if err == nil {
							annotations["obimultiplex_direction"] = map[bool]string{true: "forward", false: "reverse"}[from.Forward]

							if !match.Forward {
								barcode = barcode.ReverseComplement(true)
							}

							if tags != nil {
								library.SampleIdentifier(primerseqs[from.Marker], tags, annotations)
							}

							barcode.AnnotationsLock()
							obiutils.MustFillMap(barcode.Annotations(), annotations)
							barcode.AnnotationsUnlock()

							if barcode.Len() > 0 {
								results = append(results, barcode)
								q++
							}
						}

					}

					state = 0
				} else if match.Marker > 0 {
					log.Debugf("Marker mismatch : %d %d", match.Marker, from.Marker)
					from = match
				} else {
					log.Debugf("Marker mismatch : %d %d", match.Marker, from.Marker)
					state = 0
				}
			}
		}
	}

	if len(results) == 0 {
		sequence.SetAttribute("obimultiplex_error", "No barcode identified")
		results = append(results, sequence)
	} else {
		for i, result := range results {
			result.SetAttribute("obimultiplex_amplicon_rank", fmt.Sprintf("%d/%d", i+1, len(results)))
		}
	}

	if len(results) == 0 {
		log.Fatalf("ExtractMultiBarcode: No barcode found in sequence %s", sequence.Id())
	}

	return results, nil
}

func (library *NGSLibrary) ExtractMultiBarcodeSliceWorker(options ...WithOption) obiseq.SeqSliceWorker {
	opt := MakeOptions(options)

	if opt.AllowsIndels() {
		library.SetAllowsIndels(true)
	}

	if opt.AllowedMismatches() > 0 {
		library.SetAllowedMismatches(opt.AllowedMismatches())
	}

	library.Compile2()

	worker := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		res, err := library.ExtractMultiBarcode(sequence)

		if err != nil {
			log.Panic(err)
		}
		if res.Len() == 0 {
			log.Panicf("No barcode found in sequence %s", sequence.Id())
		}
		return res, err
	}

	return obiseq.SeqToSliceWorker(worker, true)
}
