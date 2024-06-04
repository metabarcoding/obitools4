package obingslibrary

import (
	"sort"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
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

// func (library *NGSLibrary) MakeTagMatcherFixedLength() TagMatcher {
// 	return func(sequence *obiseq.BioSequence, begin, end int, forward bool) (TagPair, error) {
// 		fb := 0
// 		fe := 0
// 		if forward {
// 			fb = begin - library.Forward_spacer - library.Forward_tag_length
// 		} else {
// 			fb = begin - library.Reverse_spacer - library.Reverse_tag_length
// 		}

// 		if fb < 0 {
// 			return TagPair{}, fmt.Errorf("begin too small")
// 		}
// 		if forward {
// 			fe = end + library.Reverse_tag_length + library.Reverse_spacer
// 		} else {
// 			fe = end + library.Forward_tag_length + library.Forward_spacer
// 		}

// 		if fe > len(sequence.String()) {
// 			return TagPair{}, fmt.Errorf("end too large")
// 		}

// 		ftag := sequence.String()[fb:begin]
// 		rtag, err := sequence.Subsequence(end, fe, true)

// 		if err != nil {
// 			return TagPair{}, fmt.Errorf("error in subsequence : %v", err)
// 		}

// 		return TagPair{Forward: ftag, Reverse: rtag.String()}, nil
// 	}

// }

func (library *NGSLibrary) ExtractMultiBarcode(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
	i := 1
	markers := make([]*Marker, len(library.Markers)+1)
	primerseqs := make([]PrimerPair, len(library.Markers)+1)
	matches := make([]PrimerMatch, 0, len(library.Markers)+1)
	aseq, err := obiapat.MakeApatSequence(sequence, false)

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
					q++
					log.Infof("%d -- %s [%s:%s] %s : %d -> %d mismatches : %d:%d",
						q,
						sequence.Id(),
						primerseqs[from.Marker].Forward,
						primerseqs[from.Marker].Reverse,
						map[bool]string{true: "forward", false: "reverse"}[from.Forward],
						from.End,
						match.Begin-1,
						from.Mismatches,
						match.Mismatches,
					)
					state = 0
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

	return nil, nil
}

func (library *NGSLibrary) ExtractMultiBarcodeSliceWorker(options ...WithOption) obiseq.SeqSliceWorker {
	opt := MakeOptions(options)

	library.Compile(opt.AllowedMismatch(), opt.AllowsIndel())

	worker := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		return library.ExtractMultiBarcode(sequence)
	}

	return obiseq.SeqToSliceWorker(worker, true)
}
