package obik

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/DavidGamba/go-getoptions"
)

// matchSliceWorker creates a SeqSliceWorker that annotates each sequence
// in a batch with k-mer match positions from the index.
// For each set, an attribute "kmer_matched_<setID>" is added containing
// a sorted []int of 0-based positions where matched k-mers start.
func matchSliceWorker(ksg *obikmer.KmerSetGroup, setIndices []int) obiseq.SeqSliceWorker {
	return func(batch obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {
		if len(batch) == 0 {
			return batch, nil
		}

		// Build slice of *BioSequence for PrepareQueries
		seqs := make([]*obiseq.BioSequence, len(batch))
		for i := range batch {
			seqs[i] = batch[i]
		}

		// Prepare queries once (shared across sets)
		queries := ksg.PrepareQueries(seqs)

		// Match against each selected set
		for _, setIdx := range setIndices {
			result := ksg.MatchBatch(setIdx, queries)

			setID := ksg.SetIDOf(setIdx)
			if setID == "" {
				setID = fmt.Sprintf("set_%d", setIdx)
			}
			attrName := "kmer_matched_" + setID

			for seqIdx, positions := range result {
				if len(positions) > 0 {
					batch[seqIdx].SetAttribute(attrName, positions)
				}
			}
		}

		return batch, nil
	}
}

// runMatch implements the "obik match" subcommand.
// It reads sequences, looks up their k-mers in a disk-based index,
// and annotates each sequence with match positions.
func runMatch(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: obik match [options] <index_directory> [sequence_files...]")
	}

	indexDir := args[0]
	seqArgs := args[1:]

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
		// All sets
		setIndices = make([]int, ksg.Size())
		for i := range setIndices {
			setIndices[i] = i
		}
	}

	// Log which sets we'll match
	for _, idx := range setIndices {
		id := ksg.SetIDOf(idx)
		if id == "" {
			id = fmt.Sprintf("set_%d", idx)
		}
		log.Infof("Matching against set %d (%s): %d k-mers", idx, id, ksg.Len(idx))
	}

	// Read sequences
	sequences, err := obiconvert.CLIReadBioSequences(seqArgs...)
	if err != nil {
		return fmt.Errorf("failed to open sequence files: %w", err)
	}

	// Apply the batch worker
	worker := matchSliceWorker(ksg, setIndices)
	matched := sequences.MakeISliceWorker(
		worker,
		false,
		obidefault.ParallelWorkers(),
	)

	obiconvert.CLIWriteBioSequences(matched, true)
	obiutils.WaitForLastPipe()

	return nil
}
