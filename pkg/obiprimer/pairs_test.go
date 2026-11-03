package obiprimer

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
)

// ecoRow is one result line of ecoPrimers.
type ecoRow struct {
	p1, p2         string
	tm1, mtm1      float64
	tm2, mtm2      float64
	gc1, gc2       int
	good           string
	in, out        int
	yule           float64
	intaxa, outtax int
	bc             float64
	well           int
	bs             float64
	mind, maxd     int
	avg            float64
}

func readEcoRows(t *testing.T, path string) []ecoRow {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var rows []ecoRow
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.Split(line, "\t")
		fl := func(i int) float64 {
			v, err := strconv.ParseFloat(strings.TrimSpace(f[i]), 64)
			if err != nil {
				t.Fatalf("field %d of %q: %v", i, line, err)
			}
			return v
		}
		in := func(i int) int { return int(fl(i)) }
		rows = append(rows, ecoRow{
			p1: f[1], p2: f[2], tm1: fl(3), mtm1: fl(4), tm2: fl(5), mtm2: fl(6),
			gc1: in(7), gc2: in(8), good: f[9], in: in(10), out: in(11), yule: fl(12),
			intaxa: in(13), outtax: in(14), bc: fl(15), well: in(16), bs: fl(17),
			mind: in(18), maxd: in(19), avg: fl(20),
		})
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return rows
}

// Opt-in end-to-end comparison with the complete ecoPrimers result tables of
// the tutorial: same parameters (-e 3 -3 2 -l 30 -L 150 -c, -r taxid).
//
// Cyprinidae is reproduced exactly: its 3448 pairs agree on every column
// (primers, good flags, sequence and taxon counts, Bc, Bs, lengths, GC, Tm and
// lowest Tm). Teleostei finds the same 407 pairs, and 6 of them count one
// sequence less than ecoPrimers (NC_033548, whose downstream site is
// NGGGGAGGCAAGTCGTAA). The 3' check of ecoPrimers reads an N through the
// shift register of extractSite, where the code 4 of N spills on the next base
// and ~4&3 reads as T, so an N can be accepted by accident. Emulating that
// register exactly in the 3' check was verified to reproduce both tables on
// every column; it is not kept: an ambiguous base cannot be a verified 3' match.
func TestPairsAgainstEcoPrimers(t *testing.T) {
	dir := os.Getenv("OBIPRIMERS_TUTORIAL")
	if dir == "" {
		t.Skip("OBIPRIMERS_TUTORIAL not set")
	}
	taxonomy, err := obiformats.LoadTaxonomy(filepath.Join(dir, "ncbitaxo_20250211.tgz"), false, false)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name, taxid string
		species     int
		divergent   int
	}{{"Cyprinidae", "7953", 316, 0}, {"Teleostei", "32443", 3876, 6}} {
		t.Run(tc.name, func(t *testing.T) {
			var stages []string
			tick := time.Now()
			lap := func(name string) {
				stages = append(stages, fmt.Sprintf("%s %s", name, time.Since(tick).Round(time.Millisecond)))
				tick = time.Now()
			}
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
			lap("read+select")
			taxa, err := corpus.AssignTaxa(taxonomy, "species")
			if err != nil {
				t.Fatal(err)
			}
			lap("taxa")
			if taxa.ExampleTaxa != tc.species {
				t.Errorf("%d example species, ecoPrimers reports %d", taxa.ExampleTaxa, tc.species)
			}

			sw, err := FindStrictWords(corpus, WordSearch{PrimerLength: 18, Quorum: 0.7, Circular: true, TmpDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			lap("strict words")
			words := make([]uint64, len(sw.Words))
			for i, w := range sw.Words {
				words[i] = w.Kmer
			}
			wm, err := MatchWords(corpus, words, MatchParams{K: 18, MaxErrors: 3, Circular: true})
			if err != nil {
				t.Fatal(err)
			}
			lap("approximate search")
			got, err := FindPairs(corpus, taxa, wm, PairParams{
				PrimerLength: 18, MinLength: 30, MaxLength: 150, ThreePrime: 2,
				Sensitivity: 0.9, FalsePositive: 0.1, MinSpecificity: 0.6, Circular: true,
			}, obithermo.DefaultModel())
			if err != nil {
				t.Fatal(err)
			}

			lap("pairs+indices+Tm")
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			t.Logf("stages: %s; Go heap in use at the end %.0f MB (includes the taxonomy)", strings.Join(stages, ", "), float64(ms.HeapAlloc)/1e6)

			want := readEcoRows(t, filepath.Join(dir, tc.name+".ecoprimers"))
			t.Logf("%d pairs, ecoPrimers %d", len(got), len(want))

			byKey := map[string]PairStats{}
			for _, g := range got {
				byKey[g.Primer1+"/"+g.Primer2] = g
			}
			bad := map[string]int{}
			missing, divergent := 0, 0
			note := func(field string, w ecoRow, format string, args ...any) {
				if bad[field]++; bad[field] <= 3 {
					t.Errorf("%s/%s %s: "+format, append([]any{w.p1, w.p2, field}, args...)...)
				}
			}
			near := func(a, b, tol float64) bool { return math.Abs(a-b) <= tol }
			for _, w := range want {
				g, ok := byKey[w.p1+"/"+w.p2]
				if !ok {
					if missing++; missing <= 5 {
						t.Errorf("ecoPrimers pair %s/%s not found", w.p1, w.p2)
					}
					continue
				}
				delete(byKey, w.p1+"/"+w.p2)
				good := string("bG"[b2i(g.Good1)]) + string("bG"[b2i(g.Good2)])
				if good != w.good {
					note("good", w, "got %s want %s", good, w.good)
				}
				// Known divergence, see the comment of the test: ecoPrimers accepts one
				// amplicon (NC_033548) whose 3' primer site ends with an N. We reject it,
				// so the counts can be exactly one lower and nothing else may differ.
				lower := func(got, want int) bool { return got == want || got == want-1 }
				diverges := g.Examples != w.in || g.ExampleTaxa != w.intaxa || g.Identified != w.well
				if diverges {
					divergent++
				}
				if !lower(g.Examples, w.in) || g.CounterExamples != w.out {
					note("sequences", w, "got %d/%d want %d/%d", g.Examples, g.CounterExamples, w.in, w.out)
				}
				if !lower(g.ExampleTaxa, w.intaxa) || g.CounterTaxa != w.outtax {
					note("taxa", w, "got %d/%d want %d/%d", g.ExampleTaxa, g.CounterTaxa, w.intaxa, w.outtax)
				}
				if !lower(g.Identified, w.well) {
					note("identified", w, "got %d want %d", g.Identified, w.well)
				}
				tol := 0.0006
				if diverges {
					tol = 0.0021
				}
				if !near(float64(g.Yule), w.yule, tol) || !near(float64(g.Bc), w.bc, tol) || !near(float64(g.Bs), w.bs, tol) {
					note("indices", w, "yule/bc/bs got %.4f/%.4f/%.4f want %.3f/%.3f/%.3f", g.Yule, g.Bc, g.Bs, w.yule, w.bc, w.bs)
				}
				if g.MinLen != w.mind || g.MaxLen != w.maxd || !near(g.MeanLen, w.avg, 0.006) {
					note("lengths", w, "got %d/%d/%.2f want %d/%d/%.2f", g.MinLen, g.MaxLen, g.MeanLen, w.mind, w.maxd, w.avg)
				}
				if g.GC1 != w.gc1 || g.GC2 != w.gc2 {
					note("gc", w, "got %d/%d want %d/%d", g.GC1, g.GC2, w.gc1, w.gc2)
				}
				// ecoPrimers prints Kelvin-273.0; obithermo uses 273.15
				if !near(g.Tm1+0.15, w.tm1, 0.051) || !near(g.Tm2+0.15, w.tm2, 0.051) {
					note("tm", w, "got %.2f/%.2f want %.1f/%.1f", g.Tm1+0.15, g.Tm2+0.15, w.tm1, w.tm2)
				}
				if !near(g.MinTm1+0.15, w.mtm1, 0.051) || !near(g.MinTm2+0.15, w.mtm2, 0.051) {
					note("mintm", w, "got %.2f/%.2f want %.1f/%.1f", g.MinTm1+0.15, g.MinTm2+0.15, w.mtm1, w.mtm2)
				}
			}
			t.Logf("missing from ours: %d; extra in ours: %d; mismatching fields: %v; pairs one sequence below ecoPrimers: %d",
				missing, len(byKey), bad, divergent)
			if divergent != tc.divergent {
				t.Errorf("%d pairs differ from ecoPrimers by one sequence, expected %d", divergent, tc.divergent)
			}
			for k := range byKey {
				if len(byKey) > 0 {
					t.Errorf("extra pair %s", k)
				}
				break
			}
		})
	}
	_ = obikmer.DecodeKmer
}
