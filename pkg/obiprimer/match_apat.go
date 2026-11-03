package obiprimer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// apatSafe returns seq when it only holds a, c, g, t; otherwise a copy where u
// becomes t and every other symbol becomes z. obiapat reads IUPAC codes in the
// sequence (an n matches every base); z is the symbol it matches with nothing,
// which gives back the ecoPrimers rule "an ambiguous base is an error".
func apatSafe(seq *obiseq.BioSequence) *obiseq.BioSequence {
	raw := seq.Sequence()
	clean := true
	for _, b := range raw {
		if b != 'a' && b != 'c' && b != 'g' && b != 't' {
			clean = false
			break
		}
	}
	if clean {
		return seq
	}
	out := make([]byte, len(raw))
	for i, b := range raw {
		switch b {
		case 'a', 'c', 'g', 't':
			out[i] = b
		case 'A', 'C', 'G', 'T':
			out[i] = b + 'a' - 'A'
		case 'u', 'U':
			out[i] = 't'
		default:
			out[i] = 'z'
		}
	}
	return obiseq.NewBioSequence(seq.Id(), out, "")
}

// matchApat searches the words with obiapat, the bit-parallel Baeza-Yates /
// Manber matcher of OBITools (the engine of ecoPrimers and obipcr).
func matchApat(c *Corpus, words []uint64, p MatchParams) ([]WordMatches, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	fwd, rev := wordStrings(words, p.K)
	pats := make([][2]obiapat.ApatPattern, len(words))
	for i := range words {
		var err error
		if pats[i][0], err = obiapat.MakeApatPattern(fwd[i], p.MaxErrors, false); err != nil {
			return nil, fmt.Errorf("obiprimer: pattern %s: %w", fwd[i], err)
		}
		if pats[i][1], err = obiapat.MakeApatPattern(rev[i], p.MaxErrors, false); err != nil {
			return nil, fmt.Errorf("obiprimer: pattern %s: %w", rev[i], err)
		}
	}

	mk := func() scanFunc {
		var recycle []obiapat.ApatSequence
		return func(seq *obiseq.BioSequence, emit func(word, pos, errors int, reverse bool)) {
			n := seq.Len()
			if n < p.K {
				return
			}
			aseq, err := obiapat.MakeApatSequence(apatSafe(seq), p.Circular, recycle...)
			if err != nil {
				panic(fmt.Sprintf("obiprimer: cannot encode sequence %s: %v", seq.Id(), err))
			}
			recycle = append(recycle[:0], aseq)

			for w := range pats {
				for strand := range 2 {
					for _, m := range pats[w][strand].FindAllIndex(aseq, 0, -1) {
						// On circular sequences obiapat also scans a copy of the
						// start: those hits repeat the ones at start-n.
						if m[0] >= 0 && m[0] < n {
							emit(w, m[0], m[2], strand == 1)
						}
					}
				}
			}
		}
	}
	return matchCorpus(c, words, mk), nil
}
