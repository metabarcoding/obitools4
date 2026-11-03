package obiprimer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

// loadTutorial reads the Cyprinidae examples of the tutorial data (the opt-in
// tests need OBIPRIMERS_TUTORIAL, see TestStrictWordsAgainstEcoPrimers).
func loadTutorial(t *testing.T) (*Corpus, *Taxa, *obitax.Taxonomy) {
	t.Helper()
	dir := os.Getenv("OBIPRIMERS_TUTORIAL")
	if dir == "" {
		t.Skip("OBIPRIMERS_TUTORIAL not set")
	}
	taxonomy, err := obiformats.LoadTaxonomy(filepath.Join(dir, "ncbitaxo_20250211.tgz"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	restrict, err := BelongingTo(taxonomy, []string{"7953"})
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
	taxa, err := corpus.AssignTaxa(taxonomy, "species")
	if err != nil {
		t.Fatal(err)
	}
	return corpus, taxa, taxonomy
}

func tutorialParams(t *testing.T) SearchParams {
	p := DefaultSearchParams()
	p.MinLength, p.MaxLength, p.ThreePrime, p.Circular = 30, 150, 2, true
	p.TmpDir = t.TempDir()
	return p
}

// ecoPrimers -U on the Cyprinidae tutorial case (-e 3 -3 2 -l 30 -L 150 -c -r 7953 -U)
// keeps 1366 pairs. testdata/ecoprimers_Cyprinidae_U.pairs.tsv holds, for each,
// the primers, the amplified example and counterexample sequences, the amplified
// example and counterexample taxa, and the identified taxa.
func TestNoMultiMatchAgainstEcoPrimers(t *testing.T) {
	corpus, taxa, _ := loadTutorial(t)
	p := tutorialParams(t)
	p.NoMultiMatch = true
	pairs, info, err := Search(corpus, taxa, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", *info)

	f, err := os.Open(filepath.Join("testdata", "ecoprimers_Cyprinidae_U.pairs.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	want := map[string][5]int{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var p1, p2 string
		var v [5]int
		if _, err := fmt.Sscanf(sc.Text(), "%s\t%s\t%d\t%d\t%d\t%d\t%d", &p1, &p2, &v[0], &v[1], &v[2], &v[3], &v[4]); err != nil {
			t.Fatal(err)
		}
		want[p1+"/"+p2] = v
	}

	if len(pairs) != len(want) {
		t.Errorf("%d pairs, ecoPrimers -U reports %d", len(pairs), len(want))
	}
	bad := 0
	for _, g := range pairs {
		v, ok := want[g.Primer1+"/"+g.Primer2]
		if !ok {
			t.Errorf("pair %s/%s is not in the ecoPrimers -U result", g.Primer1, g.Primer2)
			continue
		}
		got := [5]int{g.Examples, g.CounterExamples, g.ExampleTaxa, g.CounterTaxa, g.Identified}
		if got != v {
			if bad++; bad <= 3 {
				t.Errorf("pair %s/%s: got %v, ecoPrimers %v", g.Primer1, g.Primer2, got, v)
			}
		}
	}
}

var locRE = regexp.MustCompile(`(\d+)\.\.(\d+)`)

// ecoPrimers -R is only right for the references that happen to come first: it
// indexes the amplicons of a pair by sequence number. On NC_083213 only 382 of
// its 3405 locations are real primer sites, against all 3405 for obiprimers.
// Here the locations are checked against the sequence itself, so no
// ecoPrimers output is needed: each location must give two real primer sites
// (at most 3 substitutions) and the amplicon must be the region between them,
// read from the first primer.
func TestReferenceSitesAreRealSites(t *testing.T) {
	corpus, taxa, _ := loadTutorial(t)

	for _, ref := range []string{"NC_022183", "NC_083213"} {
		t.Run(ref, func(t *testing.T) {
			p := tutorialParams(t)
			p.Reference = ref
			pairs, _, err := Search(corpus, taxa, p)
			if err != nil {
				t.Fatal(err)
			}
			idx, err := corpus.IndexOfExample(ref)
			if err != nil {
				t.Fatal(err)
			}
			seq := strings.ToLower(string(corpus.Seqs[idx].Sequence()))

			hamming := func(a, b string) int {
				d := 0
				for i := range a {
					if a[i] != b[i] {
						d++
					}
				}
				return d
			}
			located, bad := 0, 0
			for _, pr := range pairs {
				if pr.RefLocation == "" {
					continue
				}
				located++
				m := locRE.FindAllStringSubmatch(pr.RefLocation, -1)
				if len(m) != 2 {
					t.Fatalf("location %q", pr.RefLocation)
				}
				var a1, b1, a2, b2 int
				fmt.Sscanf(m[0][0], "%d..%d", &a1, &b1)
				fmt.Sscanf(m[1][0], "%d..%d", &a2, &b2)
				left, right := seq[a1-1:b1], seq[a2-1:b2]
				between := seq[b1 : a2-1]

				p1, p2 := strings.ToLower(pr.Primer1), strings.ToLower(pr.Primer2)
				var ok bool
				if strings.HasPrefix(pr.RefLocation, "complement(") {
					ok = hamming(strings.ToLower(revcomp(strings.ToUpper(right))), p1) <= 3 && hamming(left, p2) <= 3 &&
						pr.RefAmplicon == strings.ToLower(revcomp(strings.ToUpper(between)))
				} else {
					ok = hamming(left, p1) <= 3 && hamming(strings.ToLower(revcomp(strings.ToUpper(right))), p2) <= 3 &&
						pr.RefAmplicon == between
				}
				if !ok {
					if bad++; bad <= 3 {
						t.Errorf("pair %s/%s: %s is not a real location", pr.Primer1, pr.Primer2, pr.RefLocation)
					}
				}
			}
			t.Logf("%d of %d pairs amplify %s; %d wrong locations", located, len(pairs), ref, bad)
			if located == 0 {
				t.Error("no pair amplifies the reference")
			}
		})
	}
}
