package obiclust

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	sim "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisimilarity"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// Params are the parameters of the clustering.
type Params struct {
	Threshold       float64
	Ratio           float64
	Exact           bool // attach a sequence to the best center, not to the first matching one
	Distance        bool // the score is a distance, not a similarity
	Normalized      bool // the score is normalized by the reference length
	Reference       sim.Reference
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
func newScoring(p Params, entries []sim.Entry) *sim.Scoring {
	if p.Threshold <= 0 {
		log.Fatal("Please specify a threshold > 0, and < 1 when scores are normalized.")
	}
	sc, err := sim.NewScoring(sim.Params{
		Threshold:  p.Threshold,
		Normalized: p.Normalized,
		Distance:   p.Distance,
		Reference:  p.Reference,
		Ratio:      p.Ratio,
	}, entries)
	if err != nil {
		log.Fatalf("Please specify a threshold > 0, and < 1 when scores are normalized: %v", err)
	}
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

	entries := sim.BuildEntries(len(groups), func(g int) ([]byte, int, bool) {
		head := &recs[groups[g].start]
		return head.clean, groups[g].count, head.ambiguous
	}, p.Workers)
	for i := range entries {
		if entries[i].Length > obialign.MaxBandedLCSLength {
			log.Fatalf("The sequence %s is too long (%d bp, maximum %d bp)",
				recs[groups[i].start].seq.Id(), entries[i].Length, obialign.MaxBandedLCSLength)
		}
	}

	sc := newScoring(p, entries)
	if sc.SimilarityLike() {
		log.Infof("Clustering sequences when similarity >= %f", sc.Threshold())
	} else {
		log.Infof("Clustering sequences when distance <= %f", sc.Threshold())
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
	barChecked := false // the bar is created once, even when it is disabled
	last := 0
	res := Cluster(db, ParamsFromCLI(), func(done, total int) {
		if !barChecked {
			bar, barChecked = progressBar(total), true
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
