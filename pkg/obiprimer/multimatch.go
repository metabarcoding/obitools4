package obiprimer

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"

// DropRepeatedWords removes the strict words found more than once (on either
// strand) in at least one example sequence. It is the first half of the
// ecoPrimers -U option (no multi-match).
func DropRepeatedWords(words []obikmer.SeqKmerCount) (kept []obikmer.SeqKmerCount, dropped int) {
	kept = make([]obikmer.SeqKmerCount, 0, len(words))
	for _, w := range words {
		if w.Repeated {
			dropped++
			continue
		}
		kept = append(kept, w)
	}
	return kept, dropped
}

// DropMultiMatch removes the words having two or more approximate occurrences,
// both strands counted, in at least one sequence of the corpus, whatever its
// role. It is the second half of the ecoPrimers -U option: the primers that
// can anneal at several places of a sequence are discarded.
func DropMultiMatch(wm []WordMatches) (kept []WordMatches, dropped int) {
	kept = make([]WordMatches, 0, len(wm))
	for _, w := range wm {
		multi := false
		for i := 1; i < len(w.Matches); i++ {
			if w.Matches[i].Seq == w.Matches[i-1].Seq {
				multi = true
				break
			}
		}
		if multi {
			dropped++
			continue
		}
		kept = append(kept, w)
	}
	return kept, dropped
}
