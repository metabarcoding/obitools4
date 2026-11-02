package obialign

import (
	"os"
	"slices"
	"sort"
	"sync"
	"sync/atomic"

	align "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	sim "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisimilarity"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// Params are the parameters of the comparison.
type Params struct {
	Threshold       float64 // 0: no threshold
	Normalized      bool
	Distance        bool
	Reference       sim.Reference
	AcceptAmbiguous bool
	ExtraColumns    bool
	Workers         int
}

// ParamsFromCLI collects the parameters from the command line options.
func ParamsFromCLI() Params {
	return Params{
		Threshold:       CLIThreshold(),
		Normalized:      CLINormalized(),
		Distance:        CLIDistance(),
		Reference:       CLIReference(),
		AcceptAmbiguous: CLIAcceptAmbiguous(),
		ExtraColumns:    CLIExtraColumns(),
		Workers:         obidefault.ParallelWorkers(),
	}
}

// dataset is a set of usable sequences with their description.
type dataset struct {
	seqs    obiseq.BioSequenceSlice
	entries []sim.Entry
}

// readDataset keeps the sequences that can be compared.
func readDataset(db obiseq.BioSequenceSlice, name string, p Params) dataset {
	var d dataset
	ambiguous := make([]bool, 0, len(db))
	for _, s := range db {
		ok, amb := sim.Classify(s.Sequence(), p.AcceptAmbiguous)
		if !ok {
			continue
		}
		d.seqs = append(d.seqs, s)
		ambiguous = append(ambiguous, amb)
	}
	if discarded := len(db) - len(d.seqs); discarded > 0 {
		log.Warnf("Discarded %d sequences of %s that did not contain only IUPAC nucleotides (or only 'acgt' without --accept-ambiguous)",
			discarded, name)
	}
	if len(d.seqs) == 0 {
		log.Fatalf("No valid sequences in %s.", name)
	}
	d.entries = sim.BuildEntries(len(d.seqs), func(i int) ([]byte, int, bool) {
		return d.seqs[i].Sequence(), d.seqs[i].Count(), ambiguous[i]
	}, p.Workers)
	for i := range d.entries {
		if d.entries[i].Length > align.MaxBandedLCSLength {
			log.Fatalf("The sequence %s is too long (%d bp, maximum %d bp)",
				d.seqs[i].Id(), d.entries[i].Length, align.MaxBandedLCSLength)
		}
	}
	return d
}

// pair is a pair of sequences that reaches the threshold: the index of the
// second one, and the score.
type pair struct {
	j     int
	score float64
}

// aligner compares the sequences of a dataset with the ones of an index.
type aligner struct {
	queries []sim.Entry
	index   *sim.Index
	self    bool // the index holds the queries: only the pairs i < j are reported
	sc      *sim.Scoring
}

// row finds the sequences of the index that reach the threshold with the
// query i, and appends them to out in the order of the index.
//
// Only the buckets of compatible lengths are visited, and the 4-mer filter
// works on blocks of sequences.
func (a *aligner) row(i int, buf *align.BandedLCSBuffer, commons *sim.Commons, out []pair) []pair {
	q := &a.queries[i]
	out = out[:0]
	for _, length := range a.index.Lengths() {
		if !a.sc.LengthCompatible(q.Length, length) {
			continue
		}
		b := a.index.Bucket(length)
		commons.Reset()
		start := 0
		if a.self {
			// the members are in the order of the dataset
			start = sort.Search(len(b.Members), func(k int) bool { return b.Members[k].ID > i })
		}
		for k := start; k < len(b.Members); k++ {
			m := &b.Members[k]
			if s, ok := a.sc.PairScore(q, &m.Entry, b.Common(k, q.Sparse, commons), buf); ok && a.sc.Passes(s) {
				out = append(out, pair{m.ID, a.sc.Reported(s)})
			}
		}
	}
	slices.SortFunc(out, func(x, y pair) int { return x.j - y.j })
	return out
}

func progressBar(n int) *progressbar.ProgressBar {
	if !obiconvert.CLIProgressBar() {
		return nil
	}
	return progressbar.NewOptions(n,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("[Comparing sequences]"),
	)
}

// maxPairsWithoutThreshold is the number of pairs above which a missing
// threshold is reported, the table having then as many rows.
const maxPairsWithoutThreshold = 1e8

// maxRowsPerRun is the number of rows of the table that a run is expected to
// produce at most when every pair is reported.
const maxRowsPerRun = 100000

// runShape chooses how many sequences a run covers, and how many runs a
// worker may be ahead of the oldest unfinished one. Without threshold every
// pair is a row of the table, so the runs are shortened to keep their size
// reasonable; otherwise many pairs are rejected and the runs are longer, to
// reduce the synchronizations.
func runShape(n, targets int, noThreshold bool, workers int) (rowsPerRun, window int) {
	rowsPerRun = min(max(n/(workers*32), 1), 256)
	if noThreshold {
		rowsPerRun = min(rowsPerRun, max(maxRowsPerRun/max(targets, 1), 1))
	}
	return rowsPerRun, 4 * workers
}

// Compare compares every sequence of db1 with every sequence of db2, or with
// the following ones of db1 if db2 is nil, and returns the pairs that reach the
// threshold as a table with the columns id1, id2 and score (and count1,
// count2, length1 and length2 if p.ExtraColumns).
//
// The rows are in the order of the sequences of db1, then of the ones of
// db2, whatever the number of workers.
//
// The workers take runs of rows of the matrix of the pairs, as in sumatra,
// but each run is sent to the writer with its rank: the writer puts them back
// in order, so that the output neither depends on the timing nor needs a lock
// on the output.
func Compare(db1, db2 obiseq.BioSequenceSlice, source string, p Params, onProgress func(done int)) *obiitercsv.ICSVRecord {
	d1 := readDataset(db1, "the input", p)
	d2 := d1
	self := db2 == nil
	if !self {
		d2 = readDataset(db2, "the reference file", p)
	}
	log.Infof("%d sequences", len(d1.entries))
	if !self {
		log.Infof("%d reference sequences", len(d2.entries))
	}

	all := d1.entries
	if !self {
		all = append(slices.Clone(d1.entries), d2.entries...)
	}
	sc, err := sim.NewScoring(sim.Params{
		Threshold:  p.Threshold,
		Normalized: p.Normalized,
		Distance:   p.Distance,
		Reference:  p.Reference,
		Ratio:      1e300, // no limit on the counts
	}, all)
	if err != nil {
		log.Fatalf("Wrong threshold: %v", err)
	}

	index := sim.NewIndex()
	for j := range d2.entries {
		index.Add(d2.entries[j], j)
	}
	a := &aligner{queries: d1.entries, index: index, self: self, sc: sc}

	iterator := obiitercsv.NewICSVRecord()
	for _, f := range []string{"id1", "id2", "score"} {
		iterator.AppendField(f)
	}
	if p.ExtraColumns {
		for _, f := range []string{"count1", "count2", "length1", "length2"} {
			iterator.AppendField(f)
		}
	}

	n := len(d1.entries)
	workers := max(p.Workers, 1)

	pairs := float64(n) * float64(len(d2.entries))
	if self {
		pairs = float64(n) * float64(n-1) / 2
	}
	log.Infof("%.0f pairs of sequences to compare", pairs)
	if p.Threshold == 0 && pairs > maxPairsWithoutThreshold {
		log.Warnf("No threshold has been given: the table will have %.0f rows, "+
			"use the option --threshold to keep only the similar pairs", pairs)
	}

	rowsPerRun, window := runShape(n, len(d2.entries), p.Threshold == 0, workers)
	runs := (n + rowsPerRun - 1) / rowsPerRun

	iterator.Add(workers)
	go func() {
		iterator.WaitAndClose()
	}()

	// The runs are sent to the writer with their rank, and the writer holds
	// those that arrive before the previous ones. To bound that memory, a
	// worker does not start a run more than `window` runs after the oldest run
	// that is not finished.
	var (
		lock     sync.Mutex
		progress = sync.NewCond(&lock)
		finished = make([]bool, runs)
		oldest   int
		rows     int
	)
	var next atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer iterator.Done()
			var buf align.BandedLCSBuffer
			var commons sim.Commons
			var pairs []pair
			for {
				run := int(next.Add(1)) - 1
				if run >= runs {
					return
				}
				lock.Lock()
				for run-oldest >= window {
					progress.Wait()
				}
				lock.Unlock()

				from, to := run*rowsPerRun, min((run+1)*rowsPerRun, n)
				batch := make([]obiitercsv.CSVRecord, 0, to-from)
				for i := from; i < to; i++ {
					pairs = a.row(i, &buf, &commons, pairs)
					for _, pr := range pairs {
						rec := obiitercsv.CSVRecord{
							"id1":   d1.seqs[i].Id(),
							"id2":   d2.seqs[pr.j].Id(),
							"score": pr.score,
						}
						if p.ExtraColumns {
							rec["count1"], rec["count2"] = d1.entries[i].Count, d2.entries[pr.j].Count
							rec["length1"], rec["length2"] = d1.entries[i].Length, d2.entries[pr.j].Length
						}
						batch = append(batch, rec)
					}
				}
				iterator.Push(obiitercsv.MakeCSVRecordBatch(source, run, batch))

				lock.Lock()
				finished[run] = true
				for oldest < runs && finished[oldest] {
					oldest++
				}
				rows += to - from
				done := rows
				progress.Broadcast()
				lock.Unlock()
				if onProgress != nil {
					onProgress(done)
				}
			}
		}()
	}
	return iterator
}

// CLIOBIAlign compares the sequences as the command line asks, and returns the table of the pairs.
func CLIOBIAlign(iterator obiiter.IBioSequence) *obiitercsv.ICSVRecord {
	source, db := iterator.Load()
	log.Infof("Sequence dataset of %d sequences loaded", len(db))

	var db2 obiseq.BioSequenceSlice
	if CLIHasReferenceDB() {
		db2 = CLIReferenceDB()
		if db2 == nil {
			db2 = obiseq.BioSequenceSlice{}
		}
	}

	var bar *progressbar.ProgressBar
	barChecked := false // the bar is created once, even when it is disabled
	last := int64(0)
	var lock sync.Mutex
	return Compare(db, db2, source, ParamsFromCLI(), func(done int) {
		lock.Lock()
		defer lock.Unlock()
		if !barChecked {
			bar, barChecked = progressBar(len(db)), true
		}
		if bar != nil {
			bar.Add(int(int64(done) - last))
		}
		last = int64(done)
	})
}

// CLIWriteTable writes the table to the output file given by --out, or to the standard output.
func CLIWriteTable(table *obiitercsv.ICSVRecord) {
	obicsv.CLICSVWriter(table, true)
}
