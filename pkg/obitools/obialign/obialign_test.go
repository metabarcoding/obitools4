package obialign

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	align "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	sim "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisimilarity"
)

// readTestFasta reads a fasta file whose headers are "id count=N; ...".
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

// collect reads the whole table, whatever the order the batches arrive in,
// and returns its rows in the order of the output file.
func collect(it *obiitercsv.ICSVRecord) []obiitercsv.CSVRecord {
	batches := map[int][]obiitercsv.CSVRecord{}
	for it.Next() {
		b := it.Get()
		batches[b.Order()] = b.Slice()
	}
	orders := make([]int, 0, len(batches))
	for o := range batches {
		orders = append(orders, o)
	}
	sort.Ints(orders)
	var rows []obiitercsv.CSVRecord
	for k, o := range orders {
		if o != k {
			panic(fmt.Sprintf("batch %d is missing", k))
		}
		rows = append(rows, batches[o]...)
	}
	return rows
}

// sumatraParams translates the options of sumatra.
func sumatraParams(t testing.TB, args []string) Params {
	t.Helper()
	p := Params{Normalized: true, Workers: 4}
	for i := 0; i < len(args); i++ {
		var err error
		switch args[i] {
		case "-t":
			i++
			p.Threshold, err = strconv.ParseFloat(args[i], 64)
		case "-l":
			p.Reference = sim.RefShortest
		case "-L":
			p.Reference = sim.RefLongest
		case "-d":
			p.Distance = true
		case "-r":
			p.Normalized = false
		case "-x":
			p.ExtraColumns = true
		default:
			t.Fatalf("unknown option %s", args[i])
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return p
}

// TestSumatraResults compares the tables with the outputs of sumatra 1.0.36
// (testdata/*.expected), except that the identical sequences obey the
// threshold of a raw similarity, which sumatra does not check.
func TestSumatraResults(t *testing.T) {
	cases, err := os.ReadFile(filepath.Join("testdata", "cases.txt"))
	if err != nil {
		t.Fatal(err)
	}
	a := readTestFasta(t, filepath.Join("testdata", "a.fasta"))
	b := readTestFasta(t, filepath.Join("testdata", "b.fasta"))

	for _, line := range strings.Split(strings.TrimSpace(string(cases)), "\n") {
		parts := strings.Split(line, "\t")
		kind, name, opts := parts[0], parts[1], strings.Fields(parts[2])
		t.Run(kind+"/"+name, func(t *testing.T) {
			var db2 obiseq.BioSequenceSlice
			if kind == "two" {
				db2 = b
			}
			p := sumatraParams(t, opts)
			rows := collect(Compare(a, db2, "test", p, nil))

			expected, err := os.ReadFile(filepath.Join("testdata", kind+"_"+name+".expected"))
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Split(strings.TrimSuffix(string(expected), "\n"), "\n")
			if len(expected) == 0 {
				want = nil
			}
			if len(rows) != len(want) {
				t.Fatalf("%d rows, want %d", len(rows), len(want))
			}
			for i, r := range rows {
				got := fmt.Sprintf("%s\t%s\t%f", r["id1"], r["id2"], r["score"])
				if p.ExtraColumns {
					got += fmt.Sprintf("\t%d\t%d\t%d\t%d", r["count1"], r["count2"], r["length1"], r["length2"])
				}
				if got != want[i] {
					t.Fatalf("row %d:\n got %s\nwant %s", i, got, want[i])
				}
			}
		})
	}
}

func randomDataset(rng *rand.Rand, n int, ambiguousRate float64) obiseq.BioSequenceSlice {
	var roots [][]byte
	db := make(obiseq.BioSequenceSlice, 0, n)
	for len(db) < n {
		var s []byte
		if len(roots) == 0 || rng.Intn(4) == 0 {
			s = make([]byte, 30+rng.Intn(60))
			for i := range s {
				s[i] = "acgt"[rng.Intn(4)]
			}
			roots = append(roots, s)
		} else {
			for _, c := range roots[rng.Intn(len(roots))] {
				switch r := rng.Intn(30); {
				case r == 0:
					s = append(s, "acgt"[rng.Intn(4)])
				case r == 1:
				default:
					s = append(s, c)
				}
			}
		}
		if rng.Float64() < ambiguousRate && len(s) > 3 {
			s[rng.Intn(len(s))] = "nrykm"[rng.Intn(5)]
		}
		seq := obiseq.NewBioSequence(fmt.Sprintf("s%d", len(db)), s, "")
		seq.SetAttribute("count", 1+rng.Intn(9))
		db = append(db, seq)
	}
	return db
}

// naive compares every pair one by one, with the scoring and no index.
func naive(db1, db2 obiseq.BioSequenceSlice, p Params) []string {
	self := db2 == nil
	if self {
		db2 = db1
	}
	d1, d2 := readDataset(db1, "1", p), readDataset(db2, "2", p)
	all := append(append([]sim.Entry{}, d1.entries...), d2.entries...)
	sc, err := sim.NewScoring(sim.Params{Threshold: p.Threshold, Normalized: p.Normalized, Distance: p.Distance,
		Reference: p.Reference, Ratio: 1e300}, all)
	if err != nil {
		panic(err)
	}
	var buf align.BandedLCSBuffer
	var out []string
	for i := range d1.entries {
		for j := range d2.entries {
			if self && j <= i {
				continue
			}
			if s, ok := sc.PairScore(&d1.entries[i], &d2.entries[j], -1, &buf); ok && sc.Passes(s) {
				out = append(out, fmt.Sprintf("%s\t%s\t%f", d1.seqs[i].Id(), d2.seqs[j].Id(), sc.Reported(s)))
			}
		}
	}
	return out
}

// TestCompareMatchesNaive checks the scan by length and by blocks against the
// comparison of all the pairs, in every mode, with and without ambiguous
// sequences, whatever the number of workers.
func TestCompareMatchesNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	for _, ambiguous := range []float64{0, 0.2} {
		db1 := randomDataset(rng, 90, ambiguous)
		db2 := randomDataset(rng, 40, ambiguous)
		for _, ref := range []sim.Reference{sim.RefAlignment, sim.RefLongest, sim.RefShortest} {
			for _, base := range []Params{
				{Threshold: 0.9, Normalized: true},
				{Threshold: 0.7, Normalized: true, Distance: true},
				{Threshold: 30},
				{Threshold: 6, Distance: true},
				{Normalized: true},
				{Distance: true},
			} {
				base.Reference, base.AcceptAmbiguous = ref, ambiguous > 0
				for _, second := range []obiseq.BioSequenceSlice{nil, db2} {
					want := naive(db1, second, base)
					for _, workers := range []int{1, 3, 8} {
						p := base
						p.Workers = workers
						rows := collect(Compare(db1, second, "test", p, nil))
						if len(rows) != len(want) {
							t.Fatalf("%+v two=%v workers=%d: %d rows, want %d", base, second != nil, workers, len(rows), len(want))
						}
						for i, r := range rows {
							if got := fmt.Sprintf("%s\t%s\t%f", r["id1"], r["id2"], r["score"]); got != want[i] {
								t.Fatalf("%+v two=%v workers=%d row %d: %s, want %s", base, second != nil, workers, i, got, want[i])
							}
						}
					}
				}
			}
		}
	}
}

func TestRunShape(t *testing.T) {
	// with a threshold, the runs are long
	if rows, window := runShape(40000, 40000, false, 16); rows != 78 || window != 64 {
		t.Errorf("with a threshold: %d rows, window %d", rows, window)
	}
	// without threshold, a run produces at most about maxRowsPerRun rows
	if rows, _ := runShape(40000, 40000, true, 16); rows*40000 > 3*maxRowsPerRun {
		t.Errorf("without threshold: %d rows per run give too large a table", rows)
	}
	if rows, _ := runShape(10, 10, true, 4); rows < 1 {
		t.Errorf("a run must cover one sequence at least, got %d", rows)
	}
}
