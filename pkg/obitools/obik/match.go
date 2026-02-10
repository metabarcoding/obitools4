package obik

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/DavidGamba/go-getoptions"
)

// defaultMatchQueryThreshold is the minimum number of k-mer entries to
// accumulate before launching a MatchBatch. Larger values amortize the
// cost of opening .kdi files across more query k-mers.
const defaultMatchQueryThreshold = 10_000_000

// preparedBatch pairs a batch with its pre-computed queries.
type preparedBatch struct {
	batch   obiiter.BioSequenceBatch
	seqs    []*obiseq.BioSequence
	queries *obikmer.PreparedQueries
}

// accumulatedWork holds multiple prepared batches whose queries have been
// merged into a single PreparedQueries. The flat seqs slice allows
// MatchBatch results (indexed by merged SeqIdx) to be mapped back to
// the original sequences.
type accumulatedWork struct {
	batches []obiiter.BioSequenceBatch // original batches in order
	seqs    []*obiseq.BioSequence      // flat: seqs from all batches concatenated
	queries *obikmer.PreparedQueries   // merged queries with rebased SeqIdx
}

// runMatch implements the "obik match" subcommand.
//
// Pipeline architecture (no shared mutable state between stages):
//
//	[input batches]
//	     │  Split across nCPU goroutines
//	     ▼
//	PrepareQueries (CPU, parallel)
//	     │  preparedCh
//	     ▼
//	Accumulate & MergeQueries (1 goroutine)
//	     │  matchCh — fires when totalKmers >= threshold
//	     ▼
//	MatchBatch + annotate (1 goroutine, internal parallelism per partition)
//	     │
//	     ▼
//	[output batches]
func runMatch(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	indexDir := CLIIndexDirectory()

	// Open the k-mer index
	ksg, err := obikmer.OpenKmerSetGroup(indexDir)
	if err != nil {
		return fmt.Errorf("failed to open kmer index: %w", err)
	}

	log.Infof("Opened index: k=%d, m=%d, %d partitions, %d set(s)",
		ksg.K(), ksg.M(), ksg.Partitions(), ksg.Size())

	// Resolve which sets to match against
	patterns := CLISetPatterns()
	var setIndices []int
	if len(patterns) > 0 {
		setIndices, err = ksg.MatchSetIDs(patterns)
		if err != nil {
			return fmt.Errorf("failed to match set patterns: %w", err)
		}
		if len(setIndices) == 0 {
			return fmt.Errorf("no sets match the given patterns")
		}
	} else {
		setIndices = make([]int, ksg.Size())
		for i := range setIndices {
			setIndices[i] = i
		}
	}

	for _, idx := range setIndices {
		id := ksg.SetIDOf(idx)
		if id == "" {
			id = fmt.Sprintf("set_%d", idx)
		}
		log.Infof("Matching against set %d (%s): %d k-mers", idx, id, ksg.Len(idx))
	}

	// Read input sequences
	sequences, err := obiconvert.CLIReadBioSequences(args...)
	if err != nil {
		return fmt.Errorf("failed to open sequence files: %w", err)
	}

	nworkers := obidefault.ParallelWorkers()

	// --- Stage 1: Prepare queries in parallel ---
	preparedCh := make(chan preparedBatch, nworkers)

	var prepWg sync.WaitGroup
	preparer := func(iter obiiter.IBioSequence) {
		defer prepWg.Done()
		for iter.Next() {
			batch := iter.Get()
			slice := batch.Slice()

			seqs := make([]*obiseq.BioSequence, len(slice))
			for i, s := range slice {
				seqs[i] = s
			}

			pq := ksg.PrepareQueries(seqs)

			preparedCh <- preparedBatch{
				batch:   batch,
				seqs:    seqs,
				queries: pq,
			}
		}
	}

	for i := 1; i < nworkers; i++ {
		prepWg.Add(1)
		go preparer(sequences.Split())
	}
	prepWg.Add(1)
	go preparer(sequences)

	go func() {
		prepWg.Wait()
		close(preparedCh)
	}()

	// --- Stage 2: Accumulate & merge queries ---
	matchCh := make(chan *accumulatedWork, 2)

	go func() {
		defer close(matchCh)

		var acc *accumulatedWork

		for pb := range preparedCh {
			if acc == nil {
				acc = &accumulatedWork{
					batches: []obiiter.BioSequenceBatch{pb.batch},
					seqs:    pb.seqs,
					queries: pb.queries,
				}
			} else {
				// Merge this batch's queries into the accumulator
				obikmer.MergeQueries(acc.queries, pb.queries)
				acc.batches = append(acc.batches, pb.batch)
				acc.seqs = append(acc.seqs, pb.seqs...)
			}

			// Flush when we exceed the threshold
			if acc.queries.NKmers >= defaultMatchQueryThreshold {
				matchCh <- acc
				acc = nil
			}
		}

		// Flush remaining
		if acc != nil {
			matchCh <- acc
		}
	}()

	// --- Stage 3: Match & annotate ---
	output := obiiter.MakeIBioSequence()
	if sequences.IsPaired() {
		output.MarkAsPaired()
	}

	output.Add(1)
	go func() {
		defer output.Done()

		for work := range matchCh {
			// Match against each selected set
			for _, setIdx := range setIndices {
				result := ksg.MatchBatch(setIdx, work.queries)

				setID := ksg.SetIDOf(setIdx)
				if setID == "" {
					setID = fmt.Sprintf("set_%d", setIdx)
				}
				attrName := "kmer_matched_" + setID

				for seqIdx, positions := range result {
					if len(positions) > 0 {
						work.seqs[seqIdx].SetAttribute(attrName, positions)
					}
				}
			}

			// Push annotated batches to output
			for _, b := range work.batches {
				output.Push(b)
			}

			// Help GC
			work.seqs = nil
			work.queries = nil
		}
	}()

	go output.WaitAndClose()

	obiconvert.CLIWriteBioSequences(output, true)
	obiutils.WaitForLastPipe()

	return nil
}
