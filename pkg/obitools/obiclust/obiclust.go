package obiclust

import (
	"os"
	"sync"
	"sync/atomic"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// buildEntries describes the distinct sequences of the groups, the 4-mer
// tables being computed in parallel.
func buildEntries(recs []record, groups []group, workers int) []entry {
	entries := make([]entry, len(groups))
	tables := make([]table4, len(groups)) // contiguous, in the order the centers are scanned

	var next atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < max(workers, 1); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				g := int(next.Add(1)) - 1
				if g >= len(groups) {
					return
				}
				head := &recs[groups[g].start]
				var sparse []uint16
				if !head.ambiguous {
					tables[g].fill(head.clean)
					sparse = sparseWords(&tables[g])
				}
				entries[g] = entry{
					seq:       head.clean,
					length:    len(head.clean),
					count:     groups[g].count,
					table:     &tables[g],
					sparse:    sparse,
					ambiguous: head.ambiguous,
				}
			}
		}()
	}
	wg.Wait()
	return entries
}

// Params are the parameters of the clustering.
type Params struct {
	Threshold       float64
	Ratio           float64
	Exact           bool // attach a sequence to the best center, not to the first matching one
	Distance        bool // the score is a distance, not a similarity
	Normalized      bool // the score is normalized by the reference length
	Reference       reference
	SortKey         string // "count", "None", or an attribute
	Ascending       bool
	AcceptAmbiguous bool // keep the sequences with IUPAC letters other than acgt (they cannot be centers)
	Workers         int
}

// ParamsFromCLI collects the parameters from the command line options.
func ParamsFromCLI() Params {
	return Params{
		Threshold:       CLIThreshold(),
		Ratio:           CLIRatio(),
		Exact:           CLIExact(),
		Distance:        CLIDistance(),
		Normalized:      CLINormalized(),
		Reference:       CLIReference(),
		SortKey:         CLISortKey(),
		Ascending:       CLIAscending(),
		AcceptAmbiguous: CLIAcceptAmbiguous(),
		Workers:         obidefault.ParallelWorkers(),
	}
}

// newScoring builds the scoring from the parameters, checking them as sumaclust does.
func newScoring(p Params, entries []entry) *scoring {
	threshold := p.Threshold

	if threshold <= 0 || (p.Normalized && threshold > 1.0) {
		log.Fatal("Please specify a threshold > 0, and < 1 when scores are normalized.")
	}

	sc := &scoring{
		normalize: p.Normalized,
		lcsmode:   !p.Distance,
		ref:       p.Reference,
		ratio:     p.Ratio,
	}
	if !sc.lcsmode && sc.normalize {
		threshold = 1.0 - threshold
	}
	sc.threshold = threshold

	lmax := 0
	for i := range entries {
		lmax = max(lmax, entries[i].length)
	}
	if sc.similarityLike() {
		sc.worst = 0.0
	} else {
		sc.worst = float64(lmax)
	}
	sc.prepareLengths(lmax)
	return sc
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
		progressbar.OptionSetDescription("[Clustering sequences]"),
	)
}

// Result is a clustering of a dataset.
type Result struct {
	recs      []record
	groups    []group
	clustered *clustering
	Discarded int // number of sequences that were not usable
}

// Clusters returns the number of clusters.
func (r *Result) Clusters() int { return len(r.clustered.centers) }

// Cluster clusters the sequences with the star clustering algorithm of
// sumaclust (the same as UCLUST and CD-HIT). progress, if not nil, is called
// with the number of distinct sequences processed.
func Cluster(db obiseq.BioSequenceSlice, p Params, progress func(done, total int)) *Result {
	recs, groups, discarded := prepare(db, p.SortKey, p.Ascending, p.AcceptAmbiguous)

	if discarded > 0 {
		log.Warnf("Discarded %d sequences that did not contain only IUPAC nucleotides (or only 'acgt' without --accept-ambiguous)", discarded)
	}
	if len(groups) == 0 {
		log.Fatal("No valid sequences.")
	}
	log.Infof("%d distinct sequences", len(groups))

	entries := buildEntries(recs, groups, p.Workers)
	for i := range entries {
		if entries[i].length > obialign.MaxBandedLCSLength {
			log.Fatalf("The sequence %s is too long (%d bp, maximum %d bp)",
				recs[groups[i].start].seq.Id(), entries[i].length, obialign.MaxBandedLCSLength)
		}
	}

	sc := newScoring(p, entries)
	if sc.similarityLike() {
		log.Infof("Clustering sequences when similarity >= %f", sc.threshold)
	} else {
		log.Infof("Clustering sequences when distance <= %f", sc.threshold)
	}

	cl := newClustering(entries, sc, !p.Exact, p.Workers)
	cl.run(func(done int) {
		if progress != nil {
			progress(done, len(entries))
		}
	})
	log.Infof("%d clusters created", len(cl.centers))

	return &Result{recs: recs, groups: groups, clustered: cl, Discarded: discarded}
}

// CLIOBIClust clusters the sequences and returns them annotated.
func CLIOBIClust(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	source, db := iterator.Load()
	log.Infof("Sequence dataset of %d sequences loaded", len(db))

	var bar *progressbar.ProgressBar
	last := 0
	res := Cluster(db, ParamsFromCLI(), func(done, total int) {
		if bar == nil {
			bar = progressBar(total)
		}
		if bar != nil {
			bar.Add(done - last)
		}
		last = done
	})

	if CLIOTUMap() != "" {
		table, err := obicsv.WriteCSVToFile(
			res.OTUMapIterator(source, obidefault.BatchSize()),
			CLIOTUMap(),
			obicsv.OptionsCompressed(obidefault.CompressOutput()))
		if err != nil {
			log.Fatalf("Cannot write the OTU map %s: %v", CLIOTUMap(), err)
		}
		table.Consume()
	}

	if CLINoSequenceOutput() {
		return emptyIterator()
	}
	return obiiter.IBatchOver(source, res.Annotate(), obidefault.BatchSize())
}

// emptyIterator returns an iterator that is already exhausted.
func emptyIterator() obiiter.IBioSequence {
	iterator := obiiter.MakeIBioSequence()
	iterator.Add(1)
	go func() {
		iterator.WaitAndClose()
	}()
	iterator.Done()
	return iterator
}
