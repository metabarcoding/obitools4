package obingslibrary

import (
	"fmt"
	"sort"

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

func (library *NGSLibrary) TagExtractorFixedLength(
	sequence *obiseq.BioSequence,
	annotations obiseq.Annotation,
	primers PrimerPair,
	begin, end int,
	forward bool) *TagPair {

	marker, ok := library.Markers[primers]

	forward_tag := ""
	reverse_tag := ""

	if !ok {
		log.Fatalf("marker not found : %v", primers)
	}

	fb := 0

	if forward {
		annotations["direction"] = "direct"
		fb = begin - marker.Forward_spacer - marker.Forward_tag_length
		if fb < 0 {
			annotations["demultiplex_error"] = "Cannot extract forward tag"
			return nil
		}
		forward_tag = sequence.String()[fb:(fb + marker.Forward_tag_length)]

		fb = end + marker.Reverse_spacer
		if (fb + marker.Reverse_tag_length) > sequence.Len() {
			annotations["demultiplex_error"] = "Cannot extract reverse tag"
			return nil
		}

		seq_tag, err := sequence.Subsequence(fb, fb+marker.Forward_tag_length, false)

		if err != nil {
			annotations["demultiplex_error"] = "Cannot extract forward tag"
			return nil
		}

		reverse_tag = seq_tag.ReverseComplement(true).String()

	} else {
		annotations["direction"] = "reverse"
		fb = end + marker.Forward_spacer
		if (fb + marker.Forward_tag_length) > sequence.Len() {
			annotations["demultiplex_error"] = "Cannot extract forward tag"
			return nil
		}

		seq_tag, err := sequence.Subsequence(fb, fb+marker.Forward_tag_length, false)

		if err != nil {
			annotations["demultiplex_error"] = "Cannot extract forward tag"
			return nil
		}

		forward_tag = seq_tag.ReverseComplement(true).String()

		fb = begin - marker.Reverse_spacer - marker.Reverse_tag_length
		if fb < 0 {
			sequence.SetAttribute("demultiplex_error", "Cannot extract reverse tag")
			return nil
		}
		reverse_tag = sequence.String()[fb:(fb + marker.Reverse_tag_length)]
	}

	annotations["forward_tag"] = forward_tag
	annotations["reverse_tag"] = reverse_tag

	return &TagPair{
		Forward: forward_tag,
		Reverse: reverse_tag,
	}
}

func (library *NGSLibrary) StrictSampleIdentifier(primerseqs PrimerPair, tags *TagPair, annotations obiseq.Annotation) *PCR {
	marker := library.Markers[primerseqs]
	pcr, ok := marker.samples[*tags]

	if !ok {
		return nil
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
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Begin < matches[j].Begin
		})

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

					annotations := obiseq.GetAnnotation()
					annotations["forward_primer"] = primerseqs[from.Marker].Forward
					annotations["reverse_primer"] = primerseqs[from.Marker].Reverse

					if from.Forward {

						annotations["forward_match"] = sequence.String()[from.Begin:from.End]
						sseq, err := sequence.Subsequence(match.Begin, match.End, false)

						if err != nil {
							annotations["multiplex_error"] = "Cannot extract reverse match"
						}
						annotations["reverse_match"] = sseq.ReverseComplement(true).String()

						annotations["forward_error"] = from.Mismatches
						annotations["reverse_error"] = match.Mismatches
					} else {
						annotations["reverse_match"] = sequence.String()[from.Begin:from.End]
						sseq, err := sequence.Subsequence(match.Begin, match.End, false)

						if err != nil {
							annotations["multiplex_error"] = "Cannot extract forward match"
						}

						annotations["forward_match"] = sseq.ReverseComplement(true).String()

						annotations["reverse_error"] = from.Mismatches
						annotations["forward_error"] = match.Mismatches
					}

					tags := library.TagExtractorFixedLength(sequence, annotations, primerseqs[from.Marker], from.Begin, match.End, from.Forward)

					barcode, err := sequence.Subsequence(from.End, match.Begin, false)

					if err != nil {
						return nil, fmt.Errorf("%s [%s] : Cannot extract barcode %d : %v", sequence.Id(), sequence.Source(), q, err)
					}

					if !match.Forward {
						barcode = barcode.ReverseComplement(true)
					}

					if tags != nil {
						pcr := library.StrictSampleIdentifier(primerseqs[from.Marker], tags, annotations)

						if pcr == nil {
							annotations["demultiplex_error"] = "Cannot associate sample to the tag pair"
						} else {
							annotations["sample"] = pcr.Sample
							annotations["experiment"] = pcr.Experiment
							for k, v := range pcr.Annotations {
								annotations[k] = v
							}
						}
					}

					barcode.AnnotationsLock()
					obiutils.MustFillMap(barcode.Annotations(), annotations)
					barcode.AnnotationsUnlock()

					results = append(results, barcode)
					state = 0
					q++
				} else if match.Marker > 0 {
					log.Warnf("Marker mismatch : %d %d", match.Marker, from.Marker)
					from = match
				} else {
					log.Warnf("Marker mismatch : %d %d", match.Marker, from.Marker)
					state = 0
				}
			}
		}
	}

	if len(results) == 0 {
		sequence.SetAttribute("demultiplex_error", "No barcode identified")
		results = append(results, sequence)
	} else {
		for i, result := range results {
			result.SetAttribute("amplicon_rank", fmt.Sprintf("%d/%d", i+1, len(results)))
		}
	}

	return results, nil
}

func (library *NGSLibrary) ExtractMultiBarcodeSliceWorker(options ...WithOption) obiseq.SeqSliceWorker {
	opt := MakeOptions(options)

	library.Compile(opt.AllowedMismatch(), opt.AllowsIndel())

	worker := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		return library.ExtractMultiBarcode(sequence)
	}

	return obiseq.SeqToSliceWorker(worker, true)
}
