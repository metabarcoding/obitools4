package obiprimer

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func corpusOf(roles []Role, seqs ...string) *Corpus {
	c := &Corpus{Roles: roles}
	for i, s := range seqs {
		c.Seqs = append(c.Seqs, obiseq.NewBioSequence(fmt.Sprintf("s%d", i), []byte(s), ""))
	}
	return c
}

func canon(word string) uint64 { return obikmer.EncodeCanonicalKmer([]byte(word), len(word)) }

// The reference implementation is checked by hand first: everything else is
// compared with it.
func TestNaiveMatchSemantics(t *testing.T) {
	const word = "AACCGGTA" // not a palindrome; reverse complement TACCGGTT
	cases := []struct {
		name     string
		seq      string
		circular bool
		e        int
		want     []Match
	}{
		{"exact forward", "ggaaccggtacc", false, 0, []Match{{0, 2, 0, false}}},
		{"exact reverse strand", "ggtaccggttcc", false, 0, []Match{{0, 2, 0, true}}},
		{"one error is not found at e=0", "ggaaccgctacc", false, 0, nil},
		{"one error found at e=1 with its count", "ggaaccgctacc", false, 1, []Match{{0, 2, 1, false}}},
		{"an N is an error", "ggaaccgntacc", false, 1, []Match{{0, 2, 1, false}}},
		{"an N is never a match at e=0", "ggaaccgntacc", false, 0, nil},
		{"an IUPAC code is an error, not a wildcard", "ggaaccgrtacc", false, 0, nil},
		{"U reads as T", "ggaaccgguacc", false, 0, []Match{{0, 2, 0, false}}},
		{"junction ignored on a linear sequence", "ccggtagggaa", false, 0, nil},
		{"junction found on a circular sequence", "ccggtagggaa", true, 0, []Match{{0, 9, 0, false}}},
		{"overlapping windows are all reported", "aaccggtaaccggta", false, 0, []Match{{0, 0, 0, false}, {0, 7, 0, false}}},
		{"too short", "aaccgg", true, 3, nil},
	}
	for _, tc := range cases {
		c := corpusOf([]Role{Example}, tc.seq)
		got, err := matchNaive(c, []uint64{canon(word)}, MatchParams{K: len(word), MaxErrors: tc.e, Circular: tc.circular})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got[0].Matches, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got[0].Matches, tc.want)
		}
	}
}

func TestMatchParamsValidation(t *testing.T) {
	c := corpusOf([]Role{Example}, "acgtacgtacgtacgt")
	for _, p := range []MatchParams{{K: 1}, {K: 32}, {K: 8, MaxErrors: -1}, {K: 8, MaxErrors: 8}} {
		if _, err := matchNaive(c, nil, p); err == nil {
			t.Errorf("%+v must be rejected", p)
		}
		if _, err := matchApat(c, nil, p); err == nil {
			t.Errorf("%+v must be rejected by obiapat variant", p)
		}
	}
}

func randSeq(rng *rand.Rand, n int, alphabet string) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return b
}

func mutate(rng *rand.Rand, s []byte, n int) []byte {
	out := append([]byte(nil), s...)
	for _, i := range rng.Perm(len(out))[:n] {
		out[i] = "acgt"[rng.Intn(4)]
	}
	return out
}

// randomCorpus builds sequences with words planted on both strands, with 0..3
// substitutions, across the circular junction, and with ambiguity codes.
func randomCorpus(rng *rand.Rand, words []string, nseq int) *Corpus {
	c := &Corpus{}
	for i := 0; i < nseq; i++ {
		alphabet := "acgt"
		if i%3 == 0 {
			alphabet = "acgtacgtacgtacgtacgtnnrykm"
		}
		s := randSeq(rng, 30+rng.Intn(500), alphabet)
		for range 1 + rng.Intn(4) {
			w := []byte(words[rng.Intn(len(words))])
			if rng.Intn(2) == 0 {
				w = []byte(revcomp(string(w)))
			}
			w = mutate(rng, []byte(string(w)), rng.Intn(4))
			for j := range w {
				w[j] += 'a' - 'A'
			}
			at := rng.Intn(len(s) - len(w) + 1)
			if rng.Intn(5) == 0 { // across the junction
				s = append(s[:0:0], append(append([]byte(nil), w[len(w)/2:]...), append(s[:len(s)-len(w)/2], w[:len(w)/2]...)...)...)
				continue
			}
			copy(s[at:], w)
		}
		c.Seqs = append(c.Seqs, obiseq.NewBioSequence(fmt.Sprintf("r%d", i), s, ""))
		c.Roles = append(c.Roles, Example)
	}
	return c
}

// Every implementation must return exactly what the reference returns.
func TestMatchersAgreeWithReference(t *testing.T) {
	for _, tc := range []struct {
		k, e     int
		circular bool
	}{{18, 0, false}, {18, 1, true}, {18, 3, false}, {18, 3, true}, {12, 2, true}, {25, 4, false}} {
		t.Run(fmt.Sprintf("k%d-e%d-circ%v", tc.k, tc.e, tc.circular), func(t *testing.T) {
			rng := rand.New(rand.NewSource(int64(tc.k*100 + tc.e)))
			var wordStr []string
			var words []uint64
			for range 7 {
				w := string(randSeq(rng, tc.k, "ACGT"))
				wordStr = append(wordStr, w)
				words = append(words, canon(w))
			}
			c := randomCorpus(rng, wordStr, 40)
			p := MatchParams{K: tc.k, MaxErrors: tc.e, Circular: tc.circular}

			want, err := matchNaive(c, words, p)
			if err != nil {
				t.Fatal(err)
			}
			total := 0
			for _, w := range want {
				total += len(w.Matches)
			}
			if total < 10 {
				t.Fatalf("test data too poor: only %d reference matches", total)
			}

			got, err := matchApat(c, words, p)
			if err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if !reflect.DeepEqual(got[i].Matches, want[i].Matches) {
					t.Fatalf("word %d: obiapat gives %v\nreference gives %v", i, got[i].Matches, want[i].Matches)
				}
			}
		})
	}
}

func TestSequencesMatched(t *testing.T) {
	c := corpusOf([]Role{Example, Example, CounterExample, CounterExample}, "a", "a", "a", "a")
	w := WordMatches{Matches: []Match{{0, 1, 0, false}, {0, 9, 0, true}, {1, 3, 0, false}, {3, 0, 0, false}}}
	if e, x := w.SequencesMatched(c); e != 2 || x != 1 {
		t.Errorf("got %d examples, %d counterexamples", e, x)
	}
}

func readEcoCounts(t *testing.T, path string) map[string]int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	counts := map[string]int{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var w string
		var in, out int
		if _, err := fmt.Sscanf(sc.Text(), "%s\t%d\t%d", &w, &in, &out); err != nil {
			t.Fatal(err)
		}
		counts[w] = in
	}
	return counts
}

// Opt-in comparison with ecoPrimers on the tutorial data (same variable as the
// strict word test). For every strict word, ecoPrimers logs the number of
// example sequences in which the fuzzy search (-e 3, both strands, circular)
// finds it; testdata/ecoprimers_*.words.tsv holds those numbers.
func TestMatchApatAgainstEcoPrimers(t *testing.T) {
	dir := os.Getenv("OBIPRIMERS_TUTORIAL")
	if dir == "" {
		t.Skip("OBIPRIMERS_TUTORIAL not set")
	}
	if v := os.Getenv("OBIPRIMERS_WORKERS"); v != "" { // e.g. 1, to time a single core
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatal(err)
		}
		obidefault.SetMaxCPU(n)
	}
	taxonomy, err := obiformats.LoadTaxonomy(filepath.Join(dir, "ncbitaxo_20250211.tgz"), false, false)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct{ name, taxid string }{{"Cyprinidae", "7953"}, {"Teleostei", "32443"}} {
		t.Run(tc.name, func(t *testing.T) {
			restrict, err := BelongingTo(taxonomy, []string{tc.taxid})
			if err != nil {
				t.Fatal(err)
			}
			seqs, err := obiformats.ReadSequencesFromFile(filepath.Join(dir, "mito.vert.fasta.gz"))
			if err != nil {
				t.Fatal(err)
			}
			corpus, err := LoadCorpus(seqs, Selection{Restrict: restrict})
			if err != nil {
				t.Fatal(err)
			}
			sw, err := FindStrictWords(corpus, WordSearch{PrimerLength: 18, Quorum: 0.7, Circular: true, TmpDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			words := make([]uint64, len(sw.Words))
			for i, w := range sw.Words {
				words[i] = w.Kmer
			}

			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			start := time.Now()
			res, err := matchApat(corpus, words, MatchParams{K: 18, MaxErrors: 3, Circular: true})
			elapsed := time.Since(start)
			runtime.ReadMemStats(&after)
			if err != nil {
				t.Fatal(err)
			}
			nm := 0
			for _, r := range res {
				nm += len(r.Matches)
			}
			bases := 0
			for _, s := range corpus.Seqs {
				bases += s.Len()
			}
			t.Logf("obiapat: %d words x %d sequences (%.1f Mb) e=3 in %s (%.0f Mb/s per word-strand pass); %d matches; heap +%.0f MB, allocated %.0f MB",
				len(words), len(corpus.Seqs), float64(bases)/1e6, elapsed.Round(time.Millisecond),
				float64(bases)*float64(len(words))*2/1e6/elapsed.Seconds(), nm,
				float64(after.HeapAlloc-before.HeapAlloc)/1e6, float64(after.TotalAlloc-before.TotalAlloc)/1e6)

			eco := readEcoCounts(t, filepath.Join("testdata", "ecoprimers_"+tc.name+".words.tsv"))
			compared, differ := 0, 0
			for i, r := range res {
				w := strings.ToUpper(string(obikmer.DecodeKmer(words[i], 18, nil)))
				want, ok := eco[w]
				if !ok {
					t.Errorf("word %s is not in the ecoPrimers log", w)
					continue
				}
				got, _ := r.SequencesMatched(corpus)
				compared++
				if got != want {
					differ++
					t.Errorf("word %s: %d example sequences, ecoPrimers %d", w, got, want)
				}
			}
			t.Logf("%d words compared with ecoPrimers, %d differ", compared, differ)
		})
	}
}
