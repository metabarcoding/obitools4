package obiprimer

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

// baseCode maps a nucleotide to 0..3 (A,C,G,T; U is T) and everything else to 4,
// which never equals a word base.
var baseCode = func() (t [256]byte) {
	for i := range t {
		t[i] = 4
	}
	for i, c := range "ACGT" {
		t[c], t[c+'a'-'A'] = byte(i), byte(i)
	}
	t['U'], t['u'] = 3, 3
	return
}()

func codesOf(s string) []byte {
	out := make([]byte, len(s))
	for i := range out {
		out[i] = baseCode[s[i]]
	}
	return out
}

// matchNaive is the reference implementation: every window of every sequence
// is compared with every word, base by base. It is slow and exists to define
// the exact semantics (see MatchParams) that faster implementations must
// reproduce.
func matchNaive(c *Corpus, words []uint64, p MatchParams) ([]WordMatches, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	fwd, rev := wordStrings(words, p.K)
	fc, rc := make([][]byte, len(words)), make([][]byte, len(words))
	for i := range words {
		fc[i], rc[i] = codesOf(fwd[i]), codesOf(rev[i])
	}

	mk := func() scanFunc {
		var codes []byte
		return func(seq *obiseq.BioSequence, emit func(word, pos, errors int, reverse bool)) {
			raw := seq.Sequence()
			n := len(raw)
			if n < p.K {
				return
			}
			codes = codes[:0]
			for _, b := range raw {
				codes = append(codes, baseCode[b])
			}
			windows := n - p.K + 1
			if p.Circular {
				codes = append(codes, codes[:p.K-1]...)
				windows = n
			}
			for pos := range windows {
				win := codes[pos : pos+p.K]
				for w := range words {
					for strand, pat := range [2][]byte{fc[w], rc[w]} {
						errs := 0
						for j, b := range pat {
							if win[j] != b {
								if errs++; errs > p.MaxErrors {
									break
								}
							}
						}
						if errs <= p.MaxErrors {
							emit(w, pos, errs, strand == 1)
						}
					}
				}
			}
		}
	}
	return matchCorpus(c, words, mk), nil
}
