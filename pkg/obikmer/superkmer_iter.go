package obikmer

import (
	"fmt"
	"iter"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// IterSuperKmers returns an iterator over super k-mers extracted from a DNA sequence.
// It uses the same algorithm as ExtractSuperKmers but yields super k-mers one at a time.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between m+1 and 31)
//   - m: minimizer size (must be between 1 and k-1)
//
// Returns:
//   - An iterator that yields SuperKmer structs
//
// Example:
//
//	for sk := range IterSuperKmers(sequence, 21, 11) {
//	    fmt.Printf("SuperKmer at %d-%d with minimizer %d\n", sk.Start, sk.End, sk.Minimizer)
//	}
func IterSuperKmers(seq []byte, k int, m int) iter.Seq[SuperKmer] {
	return func(yield func(SuperKmer) bool) {
		if m < 1 || m >= k || k < 2 || k > 31 || len(seq) < k {
			return
		}

		deque := make([]dequeItem, 0, k-m+1)

		mMask := uint64(1)<<(m*2) - 1
		rcShift := uint((m - 1) * 2)

		var fwdMmer, rvcMmer uint64
		for i := 0; i < m-1 && i < len(seq); i++ {
			code := uint64(__single_base_code__[seq[i]&31])
			fwdMmer = (fwdMmer << 2) | code
			rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)
		}

		superKmerStart := 0
		var currentMinimizer uint64
		firstKmer := true

		for pos := m - 1; pos < len(seq); pos++ {
			code := uint64(__single_base_code__[seq[pos]&31])
			fwdMmer = ((fwdMmer << 2) | code) & mMask
			rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)

			canonical := fwdMmer
			if rvcMmer < fwdMmer {
				canonical = rvcMmer
			}

			mmerPos := pos - m + 1

			if pos >= k-1 {
				windowStart := pos - k + 1
				for len(deque) > 0 && deque[0].position < windowStart {
					deque = deque[1:]
				}
			}

			for len(deque) > 0 && deque[len(deque)-1].canonical >= canonical {
				deque = deque[:len(deque)-1]
			}

			deque = append(deque, dequeItem{position: mmerPos, canonical: canonical})

			if pos >= k-1 {
				newMinimizer := deque[0].canonical
				kmerStart := pos - k + 1

				if firstKmer {
					currentMinimizer = newMinimizer
					firstKmer = false
				} else if newMinimizer != currentMinimizer {
					endPos := kmerStart + k - 1
					superKmer := SuperKmer{
						Minimizer: currentMinimizer,
						Start:     superKmerStart,
						End:       endPos,
						Sequence:  seq[superKmerStart:endPos],
					}
					if !yield(superKmer) {
						return
					}

					superKmerStart = kmerStart
					currentMinimizer = newMinimizer
				}
			}
		}

		if !firstKmer && len(seq[superKmerStart:]) >= k {
			superKmer := SuperKmer{
				Minimizer: currentMinimizer,
				Start:     superKmerStart,
				End:       len(seq),
				Sequence:  seq[superKmerStart:],
			}
			yield(superKmer)
		}
	}
}

// ToBioSequence converts a SuperKmer to a BioSequence with metadata.
//
// The resulting BioSequence contains:
//   - ID: "{parentID}_superkmer_{start}_{end}"
//   - Sequence: the actual DNA subsequence
//   - Attributes:
//   - "minimizer_value" (uint64): the canonical minimizer value
//   - "minimizer_seq" (string): the DNA sequence of the minimizer
//   - "k" (int): the k-mer size
//   - "m" (int): the minimizer size
//   - "start" (int): starting position in original sequence
//   - "end" (int): ending position in original sequence
//   - "parent_id" (string): ID of the parent sequence
//
// Parameters:
//   - k: k-mer size used for extraction
//   - m: minimizer size used for extraction
//   - parentID: ID of the parent sequence
//   - parentSource: source field from the parent sequence
//
// Returns:
//   - *obiseq.BioSequence: A new BioSequence representing this super k-mer
func (sk *SuperKmer) ToBioSequence(k int, m int, parentID string, parentSource string) *obiseq.BioSequence {
	// Create ID for the super-kmer
	var id string
	if parentID != "" {
		id = fmt.Sprintf("%s_superkmer_%d_%d", parentID, sk.Start, sk.End)
	} else {
		id = fmt.Sprintf("superkmer_%d_%d", sk.Start, sk.End)
	}

	// Create the BioSequence
	seq := obiseq.NewBioSequence(id, sk.Sequence, "")

	// Copy source from parent
	if parentSource != "" {
		seq.SetSource(parentSource)
	}

	// Set attributes
	seq.SetAttribute("minimizer_value", sk.Minimizer)

	// Decode the minimizer to get its DNA sequence
	minimizerSeq := DecodeKmer(sk.Minimizer, m, nil)
	seq.SetAttribute("minimizer_seq", string(minimizerSeq))

	seq.SetAttribute("k", k)
	seq.SetAttribute("m", m)
	seq.SetAttribute("start", sk.Start)
	seq.SetAttribute("end", sk.End)

	if parentID != "" {
		seq.SetAttribute("parent_id", parentID)
	}

	return seq
}

// SuperKmerWorker creates a SeqWorker that extracts super k-mers from a BioSequence
// and returns them as a slice of BioSequence objects.
//
// The worker copies the source field from the parent sequence to all extracted super k-mers.
//
// Parameters:
//   - k: k-mer size (must be between m+1 and 31)
//   - m: minimizer size (must be between 1 and k-1)
//
// Returns:
//   - SeqWorker: A worker function that can be used in obiiter pipelines
//
// Example:
//
//	worker := SuperKmerWorker(21, 11)
//	iterator := iterator.MakeIWorker(worker, false)
func SuperKmerWorker(k int, m int) obiseq.SeqWorker {
	return func(seq *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		if seq == nil {
			return obiseq.BioSequenceSlice{}, nil
		}

		// Validate parameters
		if m < 1 || m >= k || k < 2 || k > 31 {
			return obiseq.BioSequenceSlice{}, fmt.Errorf(
				"invalid parameters: k=%d, m=%d (need 1 <= m < k <= 31)",
				k, m)
		}

		sequence := seq.Sequence()
		if len(sequence) < k {
			return obiseq.BioSequenceSlice{}, nil
		}

		parentID := seq.Id()
		parentSource := seq.Source()

		// Extract super k-mers and convert to BioSequences
		result := make(obiseq.BioSequenceSlice, 0)

		for sk := range IterSuperKmers(sequence, k, m) {
			bioSeq := sk.ToBioSequence(k, m, parentID, parentSource)
			result = append(result, bioSeq)
		}

		return result, nil
	}
}
