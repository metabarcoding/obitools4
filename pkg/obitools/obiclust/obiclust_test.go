package obiclust

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// readTestFasta reads a fasta file whose headers are "id count=N; key=V;".
func readTestFasta(t testing.TB, filename string) obiseq.BioSequenceSlice {
	t.Helper()
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var db obiseq.BioSequenceSlice
	var header string
	var seq strings.Builder
	flush := func() {
		if header == "" {
			return
		}
		fields := strings.Fields(header)
		s := obiseq.NewBioSequence(fields[0], []byte(seq.String()), "")
		for _, f := range fields[1:] {
			if k, v, ok := strings.Cut(strings.TrimSuffix(f, ";"), "="); ok {
				if n, err := strconv.Atoi(v); err == nil {
					s.SetAttribute(k, n)
				}
			}
		}
		db = append(db, s)
		seq.Reset()
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if line := sc.Text(); strings.HasPrefix(line, ">") {
			flush()
			header = line[1:]
		} else {
			seq.WriteString(strings.TrimSpace(line))
		}
	}
	flush()
	return db
}

// paramsFromSumaclust translates the options of sumaclust.
func paramsFromSumaclust(t testing.TB, args []string) Params {
	t.Helper()
	p := Params{Threshold: 0.97, Ratio: 1.0, Normalized: true, SortKey: "count", Workers: 4}
	for i := 0; i < len(args); i++ {
		var err error
		switch args[i] {
		case "-t":
			i++
			p.Threshold, err = strconv.ParseFloat(args[i], 64)
		case "-R":
			i++
			p.Ratio, err = strconv.ParseFloat(args[i], 64)
		case "-s":
			i++
			p.SortKey = args[i]
		case "-e":
			p.Exact = true
		case "-l":
			p.Reference = refShortest
		case "-L":
			p.Reference = refLongest
		case "-a":
			p.Reference = refAlignment
		case "-d":
			p.Distance = true
		case "-r":
			p.Normalized = false
		case "-o":
			p.Ascending = true
		case "-g":
			p.AcceptAmbiguous = true
		default:
			t.Fatalf("unknown option %s", args[i])
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return p
}

// TestSumaclustResults compares the clusterings with the outputs of
// sumaclust 1.0.36 (testdata/*.expected: id, cluster, score, center, weight
// and sequence of every record, in the order sumaclust writes them).
func TestSumaclustResults(t *testing.T) {
	cases, err := os.ReadFile(filepath.Join("testdata", "cases.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(cases)), "\n") {
		parts := strings.Split(line, "\t")
		data, name, opts := parts[0], parts[1], strings.Fields(parts[2])
		t.Run(data+"/"+name, func(t *testing.T) {
			db := readTestFasta(t, filepath.Join("testdata", data+".fasta"))
			res := Cluster(db, paramsFromSumaclust(t, opts), nil)

			expected, err := os.ReadFile(filepath.Join("testdata", data+"_"+name+".expected"))
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Split(strings.TrimSpace(string(expected)), "\n")
			got := res.Annotate()
			if len(got) != len(want) {
				t.Fatalf("%d records, want %d", len(got), len(want))
			}
			for i, s := range got {
				score, _ := s.GetFloatAttribute("cluster_score")
				center, _ := s.GetBoolAttribute("cluster_center")
				c := map[bool]string{true: "T", false: "F"}[center]
				cluster, _ := s.GetStringAttribute("cluster")
				weight, _ := s.GetIntAttribute("cluster_weight")
				line := fmt.Sprintf("%s\t%s\t%f\t%s\t%d\t%s", s.Id(), cluster, score, c, weight, s.String())
				if line != want[i] {
					t.Fatalf("record %d:\n got %s\nwant %s", i, line, want[i])
				}
			}
		})
	}
}

// referenceCluster transcribes the sequential algorithm of sumaclust.
func referenceCluster(entries []entry, sc *scoring, fast bool) (center []int, score []float64) {
	n := len(entries)
	center, score = make([]int, n), make([]float64, n)
	var centers []int
	var buf obialign.BandedLCSBuffer
	for i := range entries {
		scoremax, best := sc.worst, -1
		for _, c := range centers {
			s := sc.pairScore(&entries[i], &entries[c], -1, &buf)
			if sc.selects(s, scoremax) {
				scoremax, best = s, c
				if fast {
					break
				}
			}
		}
		if best >= 0 {
			center[i], score[i] = best, sc.memberScore(scoremax)
		} else if entries[i].ambiguous {
			center[i] = -1
		} else {
			center[i], score[i] = i, sc.centerScore(entries[i].length)
			centers = append(centers, i)
		}
	}
	return center, score
}

func randomEntries(rng *rand.Rand, n int) []entry {
	var roots [][]byte
	entries := make([]entry, 0, n)
	for len(entries) < n {
		var s []byte
		if len(roots) == 0 || rng.Intn(5) == 0 {
			s = make([]byte, 30+rng.Intn(50))
			for i := range s {
				s[i] = "acgt"[rng.Intn(4)]
			}
			roots = append(roots, s)
		} else {
			root := roots[rng.Intn(len(roots))]
			for _, c := range root {
				switch r := rng.Intn(40); {
				case r == 0:
					s = append(s, "acgt"[rng.Intn(4)])
				case r == 1:
				case r == 2:
					s = append(s, c, "acgt"[rng.Intn(4)])
				default:
					s = append(s, c)
				}
			}
		}
		tab := buildTable4(s)
		entries = append(entries, entry{seq: s, length: len(s), count: 1 + rng.Intn(20), table: tab, sparse: sparseWords(tab)})
	}
	return entries
}

// TestParallelClusteringIsSequential checks that the clustering by blocks
// gives, whatever the number of workers and the size of the blocks, the
// result of the sequential algorithm.
func TestParallelClusteringIsSequential(t *testing.T) {
	rng := rand.New(rand.NewSource(21))
	scorings := []scoring{
		{threshold: 0.9, normalize: true, lcsmode: true, ref: refAlignment, ratio: 1, worst: 0},
		{threshold: 0.85, normalize: true, lcsmode: true, ref: refShortest, ratio: 2, worst: 0},
		{threshold: 0.9, normalize: true, lcsmode: false, ref: refLongest, ratio: 1, worst: 0},
		{threshold: 35, normalize: false, lcsmode: true, ref: refAlignment, ratio: 1, worst: 0},
		{threshold: 5, normalize: false, lcsmode: false, ref: refAlignment, ratio: 1, worst: 80},
	}
	for round := 0; round < 4; round++ {
		entries := randomEntries(rng, 300)
		for si := range scorings {
			sc := scorings[si]
			for _, fast := range []bool{true, false} {
				wantCenter, wantScore := referenceCluster(entries, &sc, fast)
				for _, workers := range []int{1, 3, 8} {
					for _, blocks := range [][2]int{{0, 0}, {1, 1}, {2, 5}, {7, 7}, {50, 300}} {
						cl := newClustering(entries, &sc, fast, workers)
						cl.blockMin, cl.blockMax = blocks[0], blocks[1]
						cl.run(func(int) {})
						for i := range entries {
							if cl.center[i] != wantCenter[i] || cl.score[i] != wantScore[i] {
								t.Fatalf("scoring %d fast=%v workers=%d blocks=%v: entry %d is (%d, %v), want (%d, %v)",
									si, fast, workers, blocks, i, cl.center[i], cl.score[i], wantCenter[i], wantScore[i])
							}
						}
					}
				}
			}
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in        string
		accept    bool
		kept      bool
		ambiguous bool
	}{
		{"acgt", false, true, false}, {"acgt", true, true, false},
		{"acgtn", false, false, false}, {"acgtn", true, true, true},
		{"acgtx", true, false, false}, {"nacgt", false, false, false},
		{"acgt-", true, false, false}, {"acgu", true, true, true},
		{"ACGT", true, false, false}, {"", true, false, false},
		{"rykmswbdhvn", true, true, true}, {"rykmswbdhvn", false, false, false},
	}
	for _, c := range cases {
		got, amb := classify([]byte(c.in), c.accept)
		if (got != nil) != c.kept || amb != c.ambiguous {
			t.Errorf("classify(%q, %v) = (%q, %v), want kept=%v ambiguous=%v", c.in, c.accept, got, amb, c.kept, c.ambiguous)
		}
	}
}

func makeSeq(id, seq string, count int, key float64) *obiseq.BioSequence {
	s := obiseq.NewBioSequence(id, []byte(seq), "")
	s.SetAttribute("count", count)
	s.SetAttribute("key", key)
	return s
}

func ids(groups []group, recs []record) string {
	var out []string
	for _, g := range groups {
		out = append(out, recs[g.start].seq.Id())
	}
	return strings.Join(out, " ")
}

// TestPrepareOrder checks the order of the sequences, which does not depend
// on the stability of a sort: decreasing count, then position; the head of a
// group of identical sequences is the one with the extreme key, then the
// first one.
func TestPrepareOrder(t *testing.T) {
	db := obiseq.BioSequenceSlice{
		makeSeq("a", "aaaacccc", 5, 1),
		makeSeq("b", "gggg", 9, 2),
		makeSeq("c", "aaaacccc", 5, 3), // same sequence as a, same count: a is the head
		makeSeq("d", "tttt", 10, 4),
		makeSeq("e", "cccc", 10, 5), // same cumulated count as d: after d, which comes first
		makeSeq("f", "aaaacccc", 7, 6),
		makeSeq("g", "nnnn", 3, 7), // discarded
		makeSeq("h", "ccgg", 2, 8),
	}

	check := func(key string, ascending bool, want string) {
		recs, groups, discarded := prepare(db, key, ascending, false)
		if discarded != 1 {
			t.Errorf("discarded = %d, want 1", discarded)
		}
		if got := ids(groups, recs); got != want {
			t.Errorf("key=%s ascending=%v: got %s, want %s", key, ascending, got, want)
		}
	}
	// a, c and f are one group of count 17, headed by f (count 7)
	check("count", false, "f d e b h")
	check("count", true, "h b d e a") // the head of a group is then the record with the smallest count
	// on the key attribute, the groups are ranked on the key of their head, the record with the largest key
	check("key", false, "h f e d b")
	check("key", true, "a b d e h")
	// no sorting: first occurrence of each sequence
	check("None", false, "a b d e h")

	// the identical sequences follow their head, with the largest count first
	recs, groups, _ := prepare(db, "count", false, false)
	var members []string
	for r := groups[0].start; r < groups[0].end; r++ {
		members = append(members, recs[r].seq.Id())
	}
	if got := strings.Join(members, " "); got != "f a c" {
		t.Errorf("group of f: %s, want f a c", got)
	}
}

// TestAmbiguousSequences checks that a sequence with IUPAC letters is attached
// to a cluster when it matches a center, and is left without cluster otherwise.
func TestAmbiguousSequences(t *testing.T) {
	root := "acgtacgtaccgtagctagcatcgatcgatcgtagctagctagcatcg"
	db := obiseq.BioSequenceSlice{
		makeSeq("center", root, 10, 0),
		makeSeq("amb", root[:20]+"n"+root[21:], 5, 0), // matches the center
		makeSeq("far", "ttttgggggccccaaaaattttgggggccccaaaaattttggggg", 4, 0),
		makeSeq("ambfar", "nnnnttttggggcccccaaaa", 3, 0), // matches nothing, cannot be a center
	}
	p := Params{Threshold: 0.9, Ratio: 1, Normalized: true, SortKey: "count", AcceptAmbiguous: true, Workers: 2}
	res := Cluster(db, p, nil)
	out := res.Annotate()
	got := map[string]string{}
	for _, s := range out {
		c, ok := s.GetStringAttribute("cluster")
		if !ok {
			c = "-"
		}
		got[s.Id()] = c
	}
	want := map[string]string{"center": "center", "amb": "center", "far": "far", "ambfar": "-"}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("%s: cluster %q, want %q", id, got[id], w)
		}
	}
	if res.Clusters() != 2 {
		t.Errorf("%d clusters, want 2", res.Clusters())
	}
	for _, s := range out {
		if s.Id() == "ambfar" {
			if center, _ := s.GetBoolAttribute("cluster_center"); center || s.HasAttribute("cluster_weight") {
				t.Error("ambfar must have cluster_center=false and no weight")
			}
		}
	}
}

// TestLengthTable checks that the values precomputed by length give the
// scores of the direct computation, in every mode.
func TestLengthTable(t *testing.T) {
	rng := rand.New(rand.NewSource(33))
	entries := randomEntries(rng, 120)
	for i := 0; i < 20; i++ { // some very short sequences, which skip the filter
		s := entries[i].seq[:8+rng.Intn(10)]
		entries[i] = entry{seq: s, length: len(s), count: 1, table: buildTable4(s)}
	}
	lmax := 0
	for i := range entries {
		lmax = max(lmax, entries[i].length)
	}
	var buf obialign.BandedLCSBuffer
	for _, ref := range []reference{refAlignment, refLongest, refShortest} {
		for _, mode := range []scoring{
			{threshold: 0.9, normalize: true, lcsmode: true},
			{threshold: 0.8, normalize: true, lcsmode: false},
			{threshold: 30, normalize: false, lcsmode: true},
			{threshold: 6, normalize: false, lcsmode: false},
		} {
			direct := mode
			direct.ref, direct.ratio = ref, 3
			table := direct
			table.prepareLengths(lmax)
			if (ref == refShortest) != (table.byLength == nil) {
				t.Fatalf("ref %v: table presence is wrong", ref)
			}
			for i := range entries {
				for j := range entries {
					a, b := direct.pairScore(&entries[i], &entries[j], -1, &buf), table.pairScore(&entries[i], &entries[j], -1, &buf)
					if a != b {
						t.Fatalf("ref %v mode %+v pair %d,%d: %v with the table, %v directly", ref, mode, i, j, b, a)
					}
				}
			}
		}
	}
}

// TestBlockCommon checks the sums of a block of transposed tables against
// commonWords.
func TestBlockCommon(t *testing.T) {
	width := simdBlockWidth()
	if width == 0 {
		t.Skip("blocks of tables are used with GOEXPERIMENT=simd only")
	}
	rng := rand.New(rand.NewSource(8))
	seq := func(n int, alpha string) []byte {
		s := make([]byte, n)
		for i := range s {
			s[i] = alpha[rng.Intn(len(alpha))]
		}
		return s
	}
	for it := 0; it < 200; it++ {
		alpha := []string{"acgt", "ac", "a"}[it%3]
		block := newTableBlock(width)
		tabs := make([]*table4, width)
		for j := range tabs {
			tabs[j] = buildTable4(seq(rng.Intn(600), alpha))
			block.set(width, j, tabs[j])
		}
		q := buildTable4(seq(rng.Intn(260), alpha))
		words := sparseWords(q)
		if words == nil {
			continue
		}
		out := make([]uint8, 64)
		block.common(width, words, out)
		for j, tab := range tabs {
			want := commonWordsRef(q, tab) - min(q.overflow, tab.overflow)
			if int(out[j]) != want {
				t.Fatalf("lane %d: %d, want %d", j, out[j], want)
			}
		}
	}
}
