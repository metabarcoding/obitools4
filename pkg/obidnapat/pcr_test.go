package obidnapat

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func revcompLower(s string) string {
	comp := map[byte]byte{'a': 't', 'c': 'g', 'g': 'c', 't': 'a', 'n': 'n'}
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		b[len(s)-1-i] = comp[s[i]]
	}
	return string(b)
}

// amplicon is what the tests compare: the bases, the orientation and the annotations
// that tell where the primers were found.
type amplicon struct {
	seq, dir, fmatch, rmatch string
	ferr, rerr               int
}

func (a amplicon) String() string {
	return fmt.Sprintf("%s %s [%s %d | %s %d]", a.dir, a.seq, a.fmatch, a.ferr, a.rmatch, a.rerr)
}

func fromSlice(res obiseq.BioSequenceSlice) []amplicon {
	var out []amplicon
	for _, s := range res {
		a := s.Annotations()
		out = append(out, amplicon{
			seq: s.String(), dir: a["direction"].(string),
			fmatch: a["forward_match"].(string), ferr: a["forward_error"].(int),
			rmatch: a["reverse_match"].(string), rerr: a["reverse_error"].(int),
		})
	}
	sortAmplicons(out)
	return out
}

func sortAmplicons(a []amplicon) {
	sort.Slice(a, func(i, j int) bool { return a[i].String() < a[j].String() })
}

func sameAmplicons(a, b []amplicon) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// pcrOracle is the in silico PCR written from its definition: the windows of each
// primer come from the naive search, the bases are read by indices modulo the length,
// and a circular sequence is read as an infinite one.
func pcrOracle(seq []byte, fwd, rev ApatPattern, o PCROptions) []amplicon {
	n, circ := len(seq), o.circular
	if n == 0 {
		return nil
	}
	at := func(i int) byte { return seq[((i%n)+n)%n] }
	text := func(from, to int) string {
		b := make([]byte, 0, max(to-from, 0))
		for i := from; i < to; i++ {
			b = append(b, at(i))
		}
		return string(b)
	}
	fRC, _ := fwd.ReverseComplement()
	rRC, _ := rev.ReverseComplement()

	var out []amplicon
	for _, forwardOrientation := range []bool{true, false} {
		first, second := fwd, rRC
		if !forwardOrientation {
			first, second = rev, fRC
		}
		for _, a := range oracle(first, seq, 0, -1, circ) {
			for _, b := range oracle(second, seq, 0, -1, circ) {
				var between, to int
				switch {
				case b[1] > a[0]:
					between, to = b[0]-a[1], b[0]
				case circ:
					between, to = b[0]+n-a[1], b[0]+n
				}
				if between <= 0 || o.minLength > 0 && between < o.minLength || o.maxLength > 0 && between > o.maxLength {
					continue
				}
				from := a[1]
				if o.extension >= 0 {
					if circ && between+(a[1]-a[0])+(b[1]-b[0])+2*o.extension > n {
						continue // longer than the molecule
					}
					from, to = a[0]-o.extension, to+(b[1]-b[0])+o.extension
					if !circ {
						if o.fullExtension && (from < 0 || to > n) {
							continue
						}
						from, to = max(from, 0), min(to, n)
					}
				}
				s := text(from, to)
				first, second := text(a[0], a[1]), text(b[0], b[1])
				x := amplicon{}
				if forwardOrientation {
					x = amplicon{seq: s, dir: "forward", fmatch: first, ferr: a[2], rmatch: revcompLower(second), rerr: b[2]}
				} else {
					x = amplicon{seq: revcompLower(s), dir: "reverse", fmatch: revcompLower(second), ferr: b[2], rmatch: first, rerr: a[2]}
				}
				out = append(out, x)
			}
		}
	}
	sortAmplicons(out)
	return out
}

// plant puts copies of the primers, of their reverse complements, with a few
// substitutions, in a random sequence.
func plantPrimers(rng *rand.Rand, n int, fwd, rev string) []byte {
	seq := randomSequence(rng, "acgt", n, "")
	sites := []string{strings.ToLower(fwd), revcompLower(strings.ToLower(fwd)), strings.ToLower(rev), revcompLower(strings.ToLower(rev))}
	for k := 0; k < 2+rng.Intn(8); k++ {
		site := []byte(sites[rng.Intn(4)])
		for i := range site {
			if rng.Intn(12) == 0 {
				site[i] = "acgt"[rng.Intn(4)]
			}
		}
		if len(site) < len(seq) {
			copy(seq[rng.Intn(len(seq)-len(site)+1):], site)
		}
	}
	return seq
}

func randomPCRCase(rng *rand.Rand, sameLength bool) (fwd, rev string, e1, e2 int, setters []PCROption, seq []byte, circ bool) {
	l1 := 8 + rng.Intn(10)
	l2 := l1
	if !sameLength {
		l2 = 8 + rng.Intn(10)
	}
	fwd, rev = randomACGT(rng, l1), randomACGT(rng, l2)
	e1, e2 = rng.Intn(3), rng.Intn(3)
	circ = rng.Intn(2) == 0
	setters = []PCROption{PCRPrimers(fwd, e1, rev, e2), PCRCircular(circ)}
	if rng.Intn(2) == 0 {
		setters = append(setters, PCRMinLength(rng.Intn(20)))
	}
	if rng.Intn(2) == 0 {
		setters = append(setters, PCRMaxLength(10+rng.Intn(100)))
	}
	if rng.Intn(3) == 0 {
		setters = append(setters, PCRExtension(rng.Intn(15)), PCRFullExtension(rng.Intn(2) == 0))
	}
	return fwd, rev, e1, e2, setters, plantPrimers(rng, 40+rng.Intn(260), fwd, rev), circ
}

// TestPCRAgainstOracle checks the amplicons, linear and circular, with lengths and
// extensions, primers of different lengths included, against the definition.
func TestPCRAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(71))
	total, wrapped := 0, 0
	for iter := 0; iter < iterations(5000); iter++ {
		fwd, rev, e1, e2, setters, seq, circ := randomPCRCase(rng, false)
		pcr, err := NewPCR(setters...)
		if err != nil {
			t.Fatal(err)
		}
		fp, _ := MakeApatPattern(fwd, e1, false)
		rp, _ := MakeApatPattern(rev, e2, false)
		if circ && (fp.Len() > len(seq) || rp.Len() > len(seq)) {
			continue
		}
		want := pcrOracle(seq, fp, rp, makePCROptions(setters))
		got := fromSlice(pcr.Amplicons(obiseq.NewBioSequence("s", append([]byte(nil), seq...), "")))
		if !sameAmplicons(got, want) {
			t.Fatalf("fwd %s/%d rev %s/%d circular=%v seq=%q options %+v\n got  %v\n want %v", fwd, e1, rev, e2, circ, seq, makePCROptions(setters), got, want)
		}
		total += len(got)
		if circ {
			wrapped += len(got)
		}
	}
	if total < 1000 || wrapped < 500 {
		t.Errorf("the draws are not varied enough: %d amplicons, %d on circular sequences", total, wrapped)
	}
}
