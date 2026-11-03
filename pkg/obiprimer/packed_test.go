package obiprimer

import (
	"bytes"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func TestPackedSeqRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	alphabet := []byte("acgtacgtacgtacgtnnrykmu-.")
	for trial := 0; trial < 500; trial++ {
		n := rng.Intn(70) // all the lengths modulo 4, the empty sequence included
		raw := make([]byte, n)
		for i := range raw {
			raw[i] = alphabet[rng.Intn(len(alphabet))]
		}
		if trial%5 == 0 && n > 10 { // a long run of N, and one at each end
			copy(raw[2:9], "nnnnnnn")
			raw[0], raw[n-1] = 'n', 'n'
		}
		p := packBases(raw)
		buf := make([]byte, 3, 8) // dirty, too small or too big: whatever
		if got := p.unpack(buf); !bytes.Equal(got, raw) {
			t.Fatalf("trial %d: %q unpacked as %q", trial, raw, got)
		}
		if len(p.bits) != (n+3)/4 {
			t.Fatalf("trial %d: %d bytes for %d bases", trial, len(p.bits), n)
		}
	}
	if p := packBases([]byte("acgtacgt")); p.exc != nil {
		t.Errorf("no exception expected: %v", p.exc)
	}
}

// Packing the corpus must change nothing to the search, ambiguous bases and
// reference included.
func TestPackedCorpusGivesTheSameSearch(t *testing.T) {
	build := func() synthetic {
		s := newSynthetic(10, -1, 7)
		rng := rand.New(rand.NewSource(3))
		for k, seq := range s.corpus.Seqs { // sprinkle ambiguous bases, some inside the sites
			raw := append([]byte(nil), seq.Sequence()...)
			for range 3 { // in the flanks, and in the insert past the first bases
				at := rng.Intn(len(raw))
				if at >= s.pos-2 && at < s.pos+18+2 || at >= s.pos+18+40-2 && at < s.pos+18+40+18+2 {
					continue
				}
				raw[at] = "nrykm"[rng.Intn(5)]
			}
			if k < 2 { // one mismatch inside a site, which the search tolerates
				raw[s.pos+5] = 'n'
			}
			s.corpus.Seqs[k] = obiseq.NewBioSequence(seq.Id(), raw, "")
		}
		return s
	}
	run := func(pack bool, circular bool) []PairStats {
		s := build()
		if pack {
			s.corpus.Pack()
			for i := range s.corpus.Seqs {
				if s.corpus.Seqs[i] != nil {
					t.Fatal("Pack must free the sequences")
				}
			}
		}
		p := s.params()
		p.Reference = "s3"
		p.ThreePrime = 3
		p.Circular = circular
		p.MinSpecificity = 0
		pairs, _, err := Search(s.corpus, s.taxa, p)
		if err != nil {
			t.Fatal(err)
		}
		return pairs
	}
	for _, circular := range []bool{false, true} {
		want := run(false, circular)
		if len(want) == 0 {
			t.Fatal("the test needs pairs")
		}
		if got := run(true, circular); !reflect.DeepEqual(got, want) {
			t.Errorf("circular=%v: packing changes the result (%d pairs, want %d)", circular, len(got), len(want))
		}
	}
}

func TestPackedCorpusAccessors(t *testing.T) {
	c := &Corpus{
		Seqs:  []*obiseq.BioSequence{obiseq.NewBioSequence("a", []byte(strings.ToLower("ACGTNNAC")), ""), obiseq.NewBioSequence("b", []byte("gg"), "")},
		Roles: []Role{Example, Other},
	}
	c.Pack()
	if c.Len() != 2 || c.SeqLen(0) != 8 || c.SeqLen(1) != 2 || c.ID(1) != "b" {
		t.Errorf("accessors: %d %d %d %q", c.Len(), c.SeqLen(0), c.SeqLen(1), c.ID(1))
	}
	if raw, _ := c.bases(0, nil); string(raw) != "acgtnnac" {
		t.Errorf("bases: %q", raw)
	}
	if i, err := c.IndexOfExample("a"); err != nil || i != 0 {
		t.Errorf("IndexOfExample: %d %v", i, err)
	}
}
