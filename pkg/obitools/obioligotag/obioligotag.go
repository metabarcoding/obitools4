package obioligotag

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obigraph"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// buildConstraints translates the command line options into constraints on
// oligonucleotides of the given size.
func buildConstraints(size int) (obikmer.KmerConstraints, error) {
	c := obikmer.NoKmerConstraints()

	if err := c.RestrictAlphabet(CLIAlphabet()); err != nil {
		return c, fmt.Errorf("invalid --alphabet: %w", err)
	}

	c.MaxGC = CLIGCMax()
	if CLIHomopolymerMax() > 0 {
		c.MaxRun = CLIHomopolymerMax()
	}
	c.MinRun = CLIHomopolymerMin()

	if pattern := CLIAccepted(); pattern != "" {
		if len(pattern) != size {
			return c, fmt.Errorf("--accepted pattern length (%d) differs from the oligonucleotide size (%d)", len(pattern), size)
		}
		masks, err := obikmer.ParseIUPACMasks(pattern)
		if err != nil {
			return c, fmt.Errorf("invalid --accepted: %w", err)
		}
		c.Accepted = masks
	}

	if pattern := CLIRejected(); pattern != "" {
		if len(pattern) != size {
			return c, fmt.Errorf("--rejected pattern length (%d) differs from the oligonucleotide size (%d)", len(pattern), size)
		}
		masks, err := obikmer.ParseIUPACMasks(pattern)
		if err != nil {
			return c, fmt.Errorf("invalid --rejected: %w", err)
		}
		c.Rejected = masks
	}

	return c, nil
}

// candidateOligos returns the sorted candidate oligonucleotides and their size:
// either every oligonucleotide satisfying the constraints, or those of the
// --oligo-list file satisfying them.
func candidateOligos() ([]uint64, int, error) {
	size := CLIOligoSize()
	list := CLIOligoList()

	switch {
	case list != "" && size != 0:
		return nil, 0, fmt.Errorf("--oligo-size and --oligo-list cannot be used together")
	case list == "" && size == 0:
		return nil, 0, fmt.Errorf("either --oligo-size or --oligo-list is required")
	case list == "" && (size < 1 || size > maxOligoSize):
		return nil, 0, fmt.Errorf("--oligo-size must be between 1 and %d", maxOligoSize)
	}

	var kmers []uint64
	if list != "" {
		var err error
		kmers, size, err = readOligoList(list)
		if err != nil {
			return nil, 0, fmt.Errorf("cannot read %s: %w", list, err)
		}
		log.Infof("%d oligonucleotides of size %d read from %s", len(kmers), size, list)
	}

	constraints, err := buildConstraints(size)
	if err != nil {
		return nil, 0, err
	}

	if list == "" {
		kmers = slices.Collect(obikmer.IterConstrainedKmers(size, constraints))
	} else {
		kmers = slices.DeleteFunc(kmers, func(km uint64) bool { return !constraints.Accepts(km, size) })
	}

	if len(kmers) == 0 {
		return nil, 0, fmt.Errorf("no oligonucleotide satisfies the constraints")
	}
	log.Infof("%d candidate oligonucleotides of size %d", len(kmers), size)
	return kmers, size, nil
}

// checkFeasibility stops the program early when the family size is beyond what
// the Hamming bound allows.
func checkFeasibility(size int, metric obikmer.KmerMetric) error {
	if metric != obikmer.HammingMetric {
		return nil
	}
	alphabet := obikmer.NoKmerConstraints()
	if err := alphabet.RestrictAlphabet(CLIAlphabet()); err != nil {
		return err
	}
	q := 0
	for _, m := range alphabet.MaxCount {
		if m != 0 {
			q++
		}
	}
	if CLIOligoList() != "" {
		q = 4 // the candidates of a list are not restricted by the alphabet of the generator
	}
	bound := obikmer.HammingBound(q, size, CLIDistance())
	if float64(CLIFamilySize()) > math.Floor(bound+1e-9) {
		return fmt.Errorf("no family of %d oligonucleotides of size %d at Hamming distance >= %d exists "+
			"(sphere-packing bound: %.0f)", CLIFamilySize(), size, CLIDistance(), math.Floor(bound+1e-9))
	}
	return nil
}

// CLIOligoFamily builds the family of oligonucleotides described by the
// command line options: it returns oligonucleotides, sorted, at a distance of at
// least --distance from each other. If no family of the required size is
// found before the time limit, the largest family found is returned together
// with a warning.
func CLIOligoFamily() ([]string, error) {
	target := CLIFamilySize()
	dmin := CLIDistance()

	if target < 1 {
		return nil, fmt.Errorf("--family-size is required and must be at least 1")
	}
	if dmin < 1 {
		return nil, fmt.Errorf("--distance must be at least 1")
	}
	metric, err := obikmer.ParseKmerMetric(CLIDistanceType())
	if err != nil {
		return nil, err
	}

	kmers, size, err := candidateOligos()
	if err != nil {
		return nil, err
	}
	if err := checkFeasibility(size, metric); err != nil {
		return nil, err
	}

	ctx := context.Background()
	if CLITimeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(CLITimeout())*time.Second)
		defer cancel()
	}

	filter := obikmer.NewKmerDissimilarityFilter(kmers, size, metric, dmin)
	progress := startCliqueProgress(target)
	result := obigraph.FindClique(ctx, len(kmers), filter, target,
		obigraph.CliqueOptions{
			Workers:  obidefault.ParallelWorkers(),
			Seed:     1,
			Progress: progress.Update,
		})
	progress.Stop(result.Reached)

	switch {
	case result.Reached:
		log.Infof("family of %d oligonucleotides found", len(result.Clique))
	case result.Exhausted:
		log.Warnf("no family of %d oligonucleotides exists with these constraints: "+
			"the best family found has %d oligonucleotides", target, len(result.Clique))
	default:
		log.Warnf("time limit reached before finding a family of %d oligonucleotides: "+
			"the best family found has %d oligonucleotides", target, len(result.Clique))
	}

	family := make([]uint64, len(result.Clique))
	for i, v := range result.Clique {
		family[i] = kmers[v]
	}
	slices.Sort(family)

	oligos := make([]string, len(family))
	for i, km := range family {
		oligos[i] = string(obikmer.DecodeKmer(km, size, nil))
	}
	return oligos, nil
}

// CLIOligosToIterator wraps oligonucleotides in a sequence iterator, with the
// identifiers oligo_001, oligo_002...
func CLIOligosToIterator(oligos []string) obiiter.IBioSequence {
	width := len(fmt.Sprint(len(oligos)))
	batchSize := max(obidefault.BatchSize(), 1)

	iterator := obiiter.MakeIBioSequence()
	iterator.Add(1)

	go func() {
		defer iterator.Done()
		for order, start := 0, 0; start < len(oligos); order, start = order+1, start+batchSize {
			end := min(start+batchSize, len(oligos))
			slice := make(obiseq.BioSequenceSlice, 0, end-start)
			for i := start; i < end; i++ {
				slice = append(slice,
					obiseq.NewBioSequence(fmt.Sprintf("oligo_%0*d", width, i+1), []byte(oligos[i]), ""))
			}
			iterator.Push(obiiter.MakeBioSequenceBatch("obioligotag", order, slice))
		}
	}()

	go iterator.WaitAndClose()
	return iterator
}
