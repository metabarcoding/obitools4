package obiprimer

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// synthetic builds nseq example sequences sharing the block F + insert + rc(R),
// each with its own flanks and its own substitution inside the insert (so that
// every sequence has a distinct amplicon, hence Bs = 1, with one taxon each).
// extraF, if not empty, is appended to the flank of sequence extraIn: that
// sequence then holds the forward site twice.
type synthetic struct {
	corpus  *Corpus
	taxa    *Taxa
	f, r, i string // forward primer, reverse primer (5'->3'), insert of sequence 0
	pos     int    // 0-based start of the block in every sequence
}

func newSynthetic(nseq int, repeatForwardIn int, seed int64) synthetic {
	rng := rand.New(rand.NewSource(seed))
	f := randDNA(rng, 18)
	r := randDNA(rng, 18)
	ins := randDNA(rng, 40)
	const pos = 60

	c := &Corpus{}
	taxa := &Taxa{Rank: "species", ExampleTaxa: nseq}
	for k := 0; k < nseq; k++ {
		in := []byte(ins)
		in[3+k] = "ACGT"[(strings.IndexByte("ACGT", in[3+k])+1)%4] // one private substitution
		s := randDNA(rng, pos) + f + string(in) + revcomp(r) + randDNA(rng, 60)
		if k == repeatForwardIn {
			s += randDNA(rng, 20) + f // the forward site a second time
		}
		c.Seqs = append(c.Seqs, obiseq.NewBioSequence(fmt.Sprintf("s%d", k), []byte(strings.ToLower(s)), ""))
		c.Roles = append(c.Roles, Example)
		taxa.Of = append(taxa.Of, int32(k))
	}
	return synthetic{c, taxa, f, r, ins, pos}
}

func (s synthetic) params() SearchParams {
	p := DefaultSearchParams()
	p.MaxErrors = 1
	p.MinLength, p.MaxLength = 20, 100
	p.Partitions = 8
	p.TmpDir = ""
	return p
}

// find returns the pair made of the primers f and r, whatever their order.
func find(pairs []PairStats, f, r string) *PairStats {
	for i := range pairs {
		if (pairs[i].Primer1 == f && pairs[i].Primer2 == r) || (pairs[i].Primer1 == r && pairs[i].Primer2 == f) {
			return &pairs[i]
		}
	}
	return nil
}

func TestSearchFindsThePlantedPair(t *testing.T) {
	s := newSynthetic(8, -1, 11)
	pairs, info, err := Search(s.corpus, s.taxa, s.params())
	if err != nil {
		t.Fatal(err)
	}
	p := find(pairs, s.f, s.r)
	if p == nil {
		t.Fatalf("the planted pair was not found among %d pairs (%+v)", len(pairs), info)
	}
	if p.Examples != 8 || p.MinLen != 40 || p.MaxLen != 40 || p.Bs != 1 || p.Bc != 1 || p.Bd != 1 {
		t.Errorf("unexpected indices %+v", *p)
	}
	if p.RefLocation != "" {
		t.Error("no reference was asked: no location expected")
	}
}

func TestReferenceLocation(t *testing.T) {
	forwardFirst, reverseFirst := 0, 0
	for seed := int64(1); seed <= 24; seed++ {
		s := newSynthetic(8, -1, seed)
		p := s.params()
		p.Reference = "s3"
		pairs, _, err := Search(s.corpus, s.taxa, p)
		if err != nil {
			t.Fatal(err)
		}
		pair := find(pairs, s.f, s.r)
		if pair == nil {
			t.Fatalf("seed %d: planted pair not found", seed)
		}

		// 1-based sites: forward primer at 61..78, then 40 bases of insert, then the reverse site
		wantLoc := "join(61..78,119..136)"
		insert := strings.ToLower(string(s.corpus.Seqs[3].Sequence()[s.pos+18 : s.pos+58]))
		if pair.Primer1 == s.f { // the amplicon is read from the first primer
			forwardFirst++
			if pair.RefLocation != wantLoc || pair.RefAmplicon != insert {
				t.Errorf("seed %d, forward first: got %s %s", seed, pair.RefLocation, pair.RefAmplicon)
			}
		} else {
			reverseFirst++
			rc := strings.ToLower(revcomp(strings.ToUpper(insert)))
			if pair.RefLocation != "complement("+wantLoc+")" || pair.RefAmplicon != rc {
				t.Errorf("seed %d, reverse first: got %s %s", seed, pair.RefLocation, pair.RefAmplicon)
			}
		}
		if len(pair.RefAmplicon) != pair.MinLen {
			t.Errorf("seed %d: amplicon length %d, pair length %d", seed, len(pair.RefAmplicon), pair.MinLen)
		}
	}
	if forwardFirst == 0 || reverseFirst == 0 {
		t.Errorf("both orientations must be exercised: %d forward first, %d reverse first", forwardFirst, reverseFirst)
	}
}

func TestReferenceErrors(t *testing.T) {
	s := newSynthetic(8, -1, 11)
	p := s.params()
	p.Reference = "nope"
	if _, _, err := Search(s.corpus, s.taxa, p); err == nil {
		t.Error("an unknown reference must be refused")
	}
	s.corpus.Roles[2] = CounterExample
	p.Reference = "s2"
	if _, _, err := Search(s.corpus, s.taxa, p); err == nil {
		t.Error("a counterexample cannot be the reference")
	}
}

func TestReferenceOnACircularJunction(t *testing.T) {
	// span writes a range running over the end of a circular sequence in two parts
	if got := span(95, 10, 100, true); got != "96..100,1..5" {
		t.Errorf("circular span %q", got)
	}
	if got := span(95, 10, 100, false); got != "96..105" {
		t.Errorf("linear span %q", got)
	}
	if got := span(10, 18, 100, true); got != "11..28" {
		t.Errorf("inner span %q", got)
	}
}

func TestNoMultiMatchDropsRepeatedPrimers(t *testing.T) {
	// the forward primer is found twice in sequence 2
	s := newSynthetic(8, 2, 11)

	with, infoWith, err := Search(s.corpus, s.taxa, s.params())
	if err != nil {
		t.Fatal(err)
	}
	if find(with, s.f, s.r) == nil {
		t.Fatalf("without -U the pair must be found (%+v)", infoWith)
	}

	p := s.params()
	p.NoMultiMatch = true
	without, info, err := Search(s.corpus, s.taxa, p)
	if err != nil {
		t.Fatal(err)
	}
	if find(without, s.f, s.r) != nil {
		t.Error("with -U the repeated forward primer must be discarded, and its pair with it")
	}
	if info.Repeated == 0 {
		t.Errorf("the repeated word was not counted: %+v", info)
	}
	if len(without) >= len(with) {
		t.Errorf("-U must not add pairs: %d with, %d without", len(with), len(without))
	}
}

func TestDropMultiMatch(t *testing.T) {
	wm := []WordMatches{
		{Word: 1, Matches: []Match{{0, 5, 0, false}, {1, 9, 0, true}}},   // one per sequence: kept
		{Word: 2, Matches: []Match{{0, 5, 0, false}, {0, 50, 1, false}}}, // two in sequence 0
		{Word: 3, Matches: []Match{{2, 5, 0, false}, {2, 5, 0, true}}},   // both strands of one sequence
		{Word: 4}, // no match at all: kept
		{Word: 5, Matches: []Match{{3, 1, 0, false}, {3, 8, 0, false}}}, // twice, but in a non-target: kept
	}
	c := corpusOf([]Role{Example, Example, CounterExample, Other}, "a", "a", "a", "a")
	kept, dropped := DropMultiMatch(wm, c)
	if dropped != 2 || len(kept) != 3 || kept[0].Word != 1 || kept[1].Word != 4 || kept[2].Word != 5 {
		t.Errorf("kept %v, dropped %d", kept, dropped)
	}
}

func TestDropRepeatedWords(t *testing.T) {
	words := []obikmer.SeqKmerCount{{Kmer: 1}, {Kmer: 2, Repeated: true}, {Kmer: 3}}
	kept, dropped := DropRepeatedWords(words)
	if dropped != 1 || len(kept) != 2 || kept[0].Kmer != 1 || kept[1].Kmer != 3 {
		t.Errorf("kept %v, dropped %d", kept, dropped)
	}
}

// A non-target amplified with the very barcode of a target taxon makes this
// target badly identified, although its barcode is unique among the targets;
// it counts in neither the sensitivity nor the false positive quorums.
func TestNonTargetSharingABarcodeLowersBsAndBd(t *testing.T) {
	s := newSynthetic(8, -1, 11)

	alone, _, err := Search(s.corpus, s.taxa, s.params())
	if err != nil {
		t.Fatal(err)
	}
	p0 := find(alone, s.f, s.r)
	if p0 == nil || p0.Bs != 1 || p0.Bd != 1 {
		t.Fatalf("without the non-target: %+v", p0)
	}

	// a copy of sequence 3, of another taxon, that is not a target
	copy3 := append([]byte(nil), s.corpus.Seqs[3].Sequence()...)
	s.corpus.Seqs = append(s.corpus.Seqs, obiseq.NewBioSequence("other", copy3, ""))
	s.corpus.Roles = append(s.corpus.Roles, Other)
	s.taxa.Of = append(s.taxa.Of, 100)

	pairs, _, err := Search(s.corpus, s.taxa, s.params())
	if err != nil {
		t.Fatal(err)
	}
	p := find(pairs, s.f, s.r)
	if p == nil {
		t.Fatal("the planted pair must still be found")
	}
	if p.Examples != 8 || p.ExampleTaxa != 8 || p.Bc != 1 {
		t.Errorf("the non-target must not change the quorums nor Bc: %+v", *p)
	}
	if p.Identified != 7 || p.Bs != 0.875 {
		t.Errorf("target 3 shares its barcode with the non-target: identified %d, Bs %v (want 7, 0.875)", p.Identified, p.Bs)
	}
	if p.Bd != 0.875 {
		t.Errorf("Bd %v, want 7 discriminant barcodes of 8", p.Bd)
	}
	if p.CounterTaxa != 0 || p.CounterExamples != 0 {
		t.Errorf("a non-target is not a counterexample: %+v", *p)
	}

	// Now counterexamples: two copies (of sequences 3 and 5) that the pair amplifies and
	// one unrelated sequence that it does not. 2 of 3 amplified is under the 0.9 quorum.
	s.corpus.Roles[len(s.corpus.Roles)-1] = CounterExample
	copy5 := append([]byte(nil), s.corpus.Seqs[5].Sequence()...)
	s.corpus.Seqs = append(s.corpus.Seqs, obiseq.NewBioSequence("counter5", copy5, ""))
	s.corpus.Roles = append(s.corpus.Roles, CounterExample)
	s.taxa.Of = append(s.taxa.Of, 101)
	unrelated := []byte(strings.ToLower(randDNA(rand.New(rand.NewSource(5)), 300)))
	s.corpus.Seqs = append(s.corpus.Seqs, obiseq.NewBioSequence("unrelated", unrelated, ""))
	s.corpus.Roles = append(s.corpus.Roles, CounterExample)
	s.taxa.Of = append(s.taxa.Of, 102)

	pq := s.params()
	pq.FalsePositive = 0.9
	pairs, _, err = Search(s.corpus, s.taxa, pq)
	if err != nil {
		t.Fatal(err)
	}
	pc := find(pairs, s.f, s.r)
	if pc == nil {
		t.Fatal("the pair amplifies 2 of 3 counterexamples (< 0.9): it must be retained")
	}
	if pc.CounterExamples != 2 || pc.CounterTaxa != 2 {
		t.Errorf("counterexamples amplified %d (taxa %d), want 2 and 2", pc.CounterExamples, pc.CounterTaxa)
	}
	if pc.Identified != 6 || pc.Bd != 0.75 {
		t.Errorf("targets 3 and 5 share their barcode with a counterexample: identified %d (want 6), Bd %v (want 0.75)",
			pc.Identified, pc.Bd)
	}
}
