package obiprimer

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func idPrefix(prefix string) obiseq.SequencePredicate {
	return func(s *obiseq.BioSequence) bool { return strings.HasPrefix(s.Id(), prefix) }
}

func TestSelectionRoles(t *testing.T) {
	seq := func(id string) *obiseq.BioSequence { return obiseq.NewBioSequence(id, []byte("acgt"), "") }
	// ids: "r" restricted clade, "x" exception inside it, "i" ignored clade, "o" other
	cases := []struct {
		name string
		sel  Selection
		want map[string]Role
	}{
		{"nothing set: everything is an example",
			Selection{}, map[string]Role{"r1": Example, "x1": Example, "i1": Example, "o1": Example}},
		{"restrict only: the rest is dropped",
			Selection{Restrict: idPrefix("r")},
			map[string]Role{"r1": Example, "x1": Ignored, "i1": Ignored, "o1": Ignored}},
		{"restrict + ignore",
			Selection{Restrict: idPrefix("r"), Ignore: idPrefix("i")},
			map[string]Role{"r1": Example, "i1": CounterExample, "o1": Ignored}},
		{"exception inside restrict becomes a counterexample",
			Selection{Restrict: idPrefix("r").Or(idPrefix("x")), Except: idPrefix("x")},
			map[string]Role{"r1": Example, "x1": CounterExample, "o1": Ignored}},
		{"example wins over ignore when both match",
			Selection{Restrict: idPrefix("r"), Ignore: idPrefix("r").Or(idPrefix("i"))},
			map[string]Role{"r1": Example, "i1": CounterExample, "o1": Ignored}},
	}
	for _, tc := range cases {
		for id, want := range tc.want {
			if got := tc.sel.Role(seq(id)); got != want {
				t.Errorf("%s: %s: role %d, want %d", tc.name, id, got, want)
			}
		}
	}
}

func TestMinSequences(t *testing.T) {
	for _, tc := range []struct {
		n    int
		q    float64
		want int
	}{{10, 0.7, 7}, {3909, 0.7, 2736}, {325, 0.7, 227}, {100, 0.9, 90}, {7, 0.9, 6}, {1, 1, 1}} {
		if got := MinSequences(tc.n, tc.q); got != tc.want {
			t.Errorf("MinSequences(%d,%g) = %d, want %d", tc.n, tc.q, got, tc.want)
		}
	}
}

func revcomp(s string) string {
	c := map[byte]byte{'A': 'T', 'C': 'G', 'G': 'C', 'T': 'A'}
	out := make([]byte, len(s))
	for i := range s {
		out[len(s)-1-i] = c[s[i]]
	}
	return string(out)
}

func randDNA(rng *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = "ACGT"[rng.Intn(4)]
	}
	return string(b)
}

func TestFindStrictWords(t *testing.T) {
	const k = 18
	rng := rand.New(rand.NewSource(1))
	motif := randDNA(rng, 30)

	var data obiseq.BioSequenceSlice
	add := func(id, s string) { data = append(data, obiseq.NewBioSequence(id, []byte(s), "")) }

	// 10 examples; the motif is in 8 of them, 3 on the reverse strand. One
	// example carries it twice. Another has an ambiguity breaking it.
	for i := 0; i < 8; i++ {
		m := motif
		if i%3 == 0 {
			m = revcomp(motif)
		}
		s := randDNA(rng, 120) + m + randDNA(rng, 120)
		if i == 5 {
			s += randDNA(rng, 50) + motif
		}
		add("ex"+string(rune('a'+i)), s)
	}
	broken := motif[:12] + "N" + motif[13:]
	add("exh", randDNA(rng, 100)+broken+randDNA(rng, 100))
	add("exi", randDNA(rng, 300))
	// counterexamples and ignored sequences must not count
	add("cx1", randDNA(rng, 50)+motif)
	add("zz1", randDNA(rng, 50)+motif)

	sel := Selection{Restrict: idPrefix("ex").Or(idPrefix("cx")), Ignore: idPrefix("cx"), Except: idPrefix("cx")}
	res, err := FindStrictWords(obiiter.IBatchOver("t", data, 3), sel,
		WordSearch{PrimerLength: k, Quorum: 0.7, Partitions: 8, TmpDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	if res.Counts != (Counts{Examples: 10, CounterExamples: 1, Ignored: 1}) {
		t.Fatalf("counts = %+v", res.Counts)
	}
	if res.MinSequences != 7 {
		t.Fatalf("quorum = %d sequences, want 7", res.MinSequences)
	}

	want := map[uint64]bool{}
	for i := 0; i+k <= len(motif); i++ {
		want[obikmer.EncodeCanonicalKmer([]byte(motif[i:i+k]), k)] = true
	}
	// the N breaks the windows that cover it, but 8 sequences still carry all of them
	if len(res.Words) != len(want) {
		t.Fatalf("%d words, want %d (the motif windows)", len(res.Words), len(want))
	}
	for _, w := range res.Words {
		if !want[w.Kmer] {
			t.Errorf("unexpected word %d", w.Kmer)
		}
		if w.Sequences != 8 {
			t.Errorf("word %d in %d sequences, want 8", w.Kmer, w.Sequences)
		}
	}
	repeated := 0
	for _, w := range res.Words {
		if w.Repeated {
			repeated++
		}
	}
	if repeated != len(want) {
		t.Errorf("%d words flagged Repeated, want all %d (motif twice in one sequence)", repeated, len(want))
	}
}

func TestFindStrictWordsParameters(t *testing.T) {
	it := func() obiiter.IBioSequence {
		return obiiter.IBatchOver("t", obiseq.BioSequenceSlice{obiseq.NewBioSequence("a", []byte("acgtacgtacgtacgtacgtacgt"), "")}, 1)
	}
	if _, err := FindStrictWords(it(), Selection{}, WordSearch{PrimerLength: 1, Quorum: 0.7}); err == nil {
		t.Error("primer length 1 must fail")
	}
	if _, err := FindStrictWords(it(), Selection{}, WordSearch{PrimerLength: 18, Quorum: 0}); err == nil {
		t.Error("quorum 0 must fail")
	}
	if _, err := FindStrictWords(it(), Selection{Restrict: idPrefix("zz")},
		WordSearch{PrimerLength: 18, Quorum: 0.7}); err == nil {
		t.Error("no example must fail")
	}
}

// Opt-in end-to-end check against ecoPrimers on the "Designing fish primers"
// tutorial data. Set OBIPRIMERS_TUTORIAL to the directory holding
// ncbitaxo_20250211.tgz and mito.vert.fasta.gz.
//
// ecoPrimers (-e 3 -3 2 -l 30 -L 150 -c, k=18, strict quorum 0.7) reports
// "Strict primer count": 1208 for Cyprinidae (-r 7953, 325 examples) and 255
// for Teleostei (-r 32443, 3909 examples).
//
// Teleostei matches exactly. For Cyprinidae we find 1207: ecoPrimers keeps
// TTTATGAAAACAAATAAA, which a brute-force scan (both strands, circular
// junction) finds in only 225 of the 325 examples, below its own quorum of
// 227. That word is in none of the 3448 pairs ecoPrimers finally retains, so
// the divergence has no effect on the reference results.
func TestStrictWordsAgainstEcoPrimers(t *testing.T) {
	dir := os.Getenv("OBIPRIMERS_TUTORIAL")
	if dir == "" {
		t.Skip("OBIPRIMERS_TUTORIAL not set")
	}

	taxonomy, err := obiformats.LoadTaxonomy(filepath.Join(dir, "ncbitaxo_20250211.tgz"), false, false)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		taxid    string
		examples int
		words    int
	}{
		{"Cyprinidae", "7953", 325, 1207}, // ecoPrimers: 1208, see above
		{"Teleostei", "32443", 3909, 255},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restrict, err := BelongingTo(taxonomy, []string{tc.taxid})
			if err != nil {
				t.Fatal(err)
			}
			// The predicates read the taxon through the taxid annotation.
			seqs, err := obiformats.ReadSequencesFromFile(filepath.Join(dir, "mito.vert.fasta.gz"))
			if err != nil {
				t.Fatal(err)
			}
			res, err := FindStrictWords(seqs, Selection{Restrict: restrict},
				WordSearch{PrimerLength: 18, Quorum: 0.7, Circular: true, TmpDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%+v min=%d words=%d", res.Counts, res.MinSequences, len(res.Words))
			if res.Counts.Examples != tc.examples {
				t.Errorf("%d examples, ecoPrimers reports %d", res.Counts.Examples, tc.examples)
			}
			if len(res.Words) != tc.words {
				t.Errorf("%d strict words, ecoPrimers reports %d", len(res.Words), tc.words)
			}
		})
	}
}
