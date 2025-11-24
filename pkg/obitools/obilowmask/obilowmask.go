package obilowmask

import (
	"fmt"
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// MaskingMode defines how to handle low-complexity regions
type MaskingMode int

const (
	Mask  MaskingMode = iota // Mask mode: replace low-complexity regions with masked characters
	Split                    // Split mode: split sequence into high-complexity fragments
	Extract
)

// LowMaskWorker creates a worker to mask low-complexity regions in DNA sequences.
//
// Algorithm principle:
// Calculate the normalized entropy of each k-mer at different scales (wordSize = 1 to level_max).
// K-mers with entropy below the threshold are masked.
//
// Parameters:
//   - kmer_size: size of the sliding window for entropy calculation
//   - level_max: maximum word size used for entropy calculation (finest scale)
//   - threshold: normalized entropy threshold below which masking occurs (between 0 and 1)
//   - mode: Mask (masking) or Split (splitting)
//   - maskChar: character used for masking (typically 'n' or 'N')
//
// Returns: a SeqWorker function that can be applied to each sequence
func LowMaskWorker(kmer_size int, level_max int, threshold float64, mode MaskingMode, maskChar byte) obiseq.SeqWorker {

	// ========================================================================
	// FUNCTION 1: emax - Calculate theoretical maximum entropy
	// ========================================================================
	// Computes the maximum entropy of a k-mer of length lseq containing words of size word_size.
	//
	// Maximum entropy depends on the theoretical optimal word distribution:
	// - If we have more positions (nw) than possible canonical words (na),
	//   some words will appear multiple times
	// - We calculate the entropy of a distribution where all words appear
	//   cov or cov+1 times (most uniform distribution possible)
	//
	// IMPORTANT: Uses CanonicalKmerCount to get the actual number of canonical words
	// after circular normalization (e.g., "atg", "tga", "gat" → all "atg").
	// This is much smaller than 4^word_size (e.g., 10 instead of 16 for word_size=2).
	emax := func(lseq, word_size int) float64 {
		nw := lseq - word_size + 1                  // Number of words in a k-mer of length lseq
		na := obikmer.CanonicalKmerCount(word_size) // Number of canonical words after normalization

		// Case 1: Fewer positions than possible words
		// Maximum entropy is simply log(nw) since we can have at most nw different words
		if nw < na {
			return math.Log(float64(nw))
		}

		// Case 2: More positions than possible words
		// Some words must appear multiple times
		cov := nw / na             // Average coverage (average number of occurrences per word)
		remains := nw - (na * cov) // Number of words that will have one additional occurrence

		// Calculate frequencies in the optimal distribution:
		// - (na - remains) words appear cov times → frequency f1 = cov/nw
		// - remains words appear (cov+1) times → frequency f2 = (cov+1)/nw
		f1 := float64(cov) / float64(nw)
		f2 := float64(cov+1) / float64(nw)

		// Shannon entropy: H = -Σ p(i) * log(p(i))
		// where p(i) is the probability of observing word i
		return -(float64(na-remains)*f1*math.Log(f1) +
			float64(remains)*f2*math.Log(f2))
	}

	// ========================================================================
	// FUNCTION 2: maskAmbiguities - Mark positions containing ambiguities
	// ========================================================================
	// Identifies positions with ambiguous nucleotides (N, Y, R, etc.) and marks
	// all k-mers that contain them.
	//
	// Returns: a slice where maskPositions[i] = -1 if position i is part of a
	//          k-mer containing an ambiguity, 0 otherwise
	maskAmbiguities := func(sequence []byte) []int {
		maskPositions := make([]int, len(sequence))
		for i, nuc := range sequence {
			// If nucleotide is not a, c, g or t (lowercase), it's an ambiguity
			if nuc != 'a' && nuc != 'c' && nuc != 'g' && nuc != 't' {
				// Mark all positions of k-mers that contain this nucleotide
				// A k-mer starting at position (i - kmer_size + 1) will contain position i
				end := max(0, i-kmer_size+1)
				for j := i; j >= end; j-- {
					maskPositions[j] = -1
				}
			}
		}
		return maskPositions
	}

	// ========================================================================
	// FUNCTION 3: cleanTable - Reset a frequency table to zero
	// ========================================================================
	cleanTable := func(table []int, over int) {
		for i := 0; i < over; i++ {
			table[i] = 0
		}
	}

	// ========================================================================
	// FUNCTION 4: slidingMin - Calculate sliding minimum over a window
	// ========================================================================
	// Applies a sliding window of size window over data and replaces each
	// value with the minimum in the window centered on that position.
	//
	// Uses a MinMultiset to efficiently maintain the minimum in the window.
	slidingMin := func(data []float64, window int) {
		minimier := obiutils.NewMinMultiset(func(a, b float64) bool { return a < b })
		ldata := len(data)
		mem := make([]float64, window) // Circular buffer to store window values

		// Initialize buffer with sentinel value
		for i := range mem {
			mem[i] = 10000
		}

		for i, v := range data {
			// Get the old value leaving the window
			m := mem[i%window]
			mem[i%window] = v

			// Remove old value from multiset if it was valid
			if m < 10000 {
				minimier.RemoveOne(m)
			}

			// Add new value if full window is ahead of us
			if (ldata - i) >= window {
				minimier.Add(v)
			}

			// log.Warnf("taille du minimier %d @ %d", minimier.Len(), i)

			// Retrieve and store current minimum
			var ok bool
			if data[i], ok = minimier.Min(); !ok {
				log.Error("problem with minimum entropy")
				data[i] = 0.0
			}

			//xx, _ := minimier.Min()
			//log.Warnf("Pos: %d n: %d min: %.3f -> %.3f", i, minimier.Len(), v, xx)
		}
	}

	// ========================================================================
	// FUNCTION 5: computeEntropies - Calculate normalized entropy for each position
	// ========================================================================
	// This is the central function that calculates the entropy of each k-mer in the sequence
	// at a given scale (wordSize).
	//
	// Algorithm:
	// 1. Encode the sequence into words (subsequences of size wordSize)
	// 2. For each k-mer, count the frequencies of words it contains
	// 3. Calculate normalized entropy = observed_entropy / maximum_entropy
	// 4. Apply a sliding min filter to smooth results
	//
	// IMPORTANT: Line 147 uses NormalizeInt for circular normalization of words!
	// This means "atg", "tga", and "gat" are considered the same word.
	computeEntropies := func(sequence []byte,
		maskPositions []int, // Positions of ambiguities
		entropies []float64, // Output: normalized entropies for each position
		table []int, // Frequency table for words (reused between calls)
		words []int, // Buffer to store encoded words (reused)
		wordSize int) { // Word size (scale of analysis)

		lseq := len(sequence)              // Sequence length
		tableSize := 1 << (wordSize * 2)   // Actual table size (must fit all codes 0 to 4^wordSize-1)
		nwords := kmer_size - wordSize + 1 // Number of words in a k-mer
		float_nwords := float64(nwords)
		log_nwords := math.Log(float_nwords)    // log(nwords) used in entropy calculation
		entropyMax := emax(kmer_size, wordSize) // Theoretical maximum entropy (uses CanonicalKmerCount internally)

		// Reset frequency table (must clear entire table, not just nalpha entries)
		cleanTable(table, tableSize)

		for i := 1; i < lseq; i++ {
			entropies[i] = 6
		}
		end := lseq - wordSize + 1 // Last position where a word can start

		// ========================================================================
		// STEP 1: Encode all words in the sequence
		// ========================================================================
		// Uses left-shift encoding: each nucleotide is encoded on 2 bits
		// a=00, c=01, g=10, t=11

		mask := (1 << (wordSize * 2)) - 1 // Mask to keep only last wordSize*2 bits

		// Initialize first word (all nucleotides except the last one)
		word_index := 0
		for i := 0; i < wordSize-1; i++ {
			word_index = (word_index << 2) + int(obikmer.EncodeNucleotide(sequence[i]))
		}

		// Encode all words with sliding window
		for i, j := 0, wordSize-1; i < end; i, j = i+1, j+1 {
			// Shift left by 2 bits, mask, and add new nucleotide
			word_index = ((word_index << 2) & mask) + int(obikmer.EncodeNucleotide(sequence[j]))

			// *** CIRCULAR NORMALIZATION ***
			// Convert word to its canonical form (smallest by circular rotation)
			// This is where "atg", "tga", "gat" all become "atg"
			words[i] = obikmer.NormalizeInt(word_index, wordSize)
		}

		// ========================================================================
		// STEP 2: Calculate entropy for each k-mer with sliding window
		// ========================================================================
		s := 0            // Number of words processed in current k-mer
		sum_n_logn := 0.0 // Sum of n*log(n) for entropy calculation
		entropy := 1.0    // Current normalized entropy
		cleaned := true   // Flag indicating if table has been cleaned

		for i := range end {
			s++

			switch {
			// CASE 1: Filling phase (fewer than nwords words collected)
			case s < nwords:
				cleaned = false
				table[words[i]]++ // Increment word frequency

			// CASE 2: Position contains an ambiguity
			case i >= (nwords-1) && maskPositions[i-nwords+1] < 0:
				entropies[i-nwords+1] = 4.0 // Mark entropy as invalid
				if !cleaned {
					cleanTable(table, tableSize) // Reset table
				}
				cleaned = true
				s = 0
				sum_n_logn = 0.0

			// CASE 3: First complete k-mer (s == nwords)
			case s == nwords:
				cleaned = false
				table[words[i]]++

				// Calculate Shannon entropy: H = -Σ p(i)*log(p(i))
				// = log(N) - (1/N)*Σ n(i)*log(n(i))
				// where N = nwords, n(i) = frequency of word i
				//
				// NOTE: We iterate over entire table (tableSize = 4^wordSize) to count all frequencies.
				// Canonical codes are not contiguous (e.g., for k=2: {0,1,2,3,5,6,7,10,11,15})
				// so we must scan the full table even though only ~10 entries will be non-zero
				sum_n_logn = 0
				for j := range tableSize {
					n := float64(table[j])
					if n > 0 {
						sum_n_logn += n * math.Log(n)
					}
				}
				// Normalized entropy = observed entropy / maximum entropy
				entropy = (log_nwords - sum_n_logn/float_nwords) / entropyMax

			// CASE 4: Sliding window (s > nwords)
			// Incremental update of entropy by adding a new word
			// and removing the old one
			case s > nwords:
				cleaned = false

				new_word := words[i]
				old_word := words[i-nwords]

				// Optimization: only recalculate if word changes
				if old_word != new_word {
					table[new_word]++
					table[old_word]--

					n_old := float64(table[old_word])
					n_new := float64(table[new_word])

					// Incremental update of sum_n_logn
					// Remove contribution of old word (before decrement)
					sum_n_logn -= (n_old + 1) * math.Log(n_old+1)
					// Add contribution of old word (after decrement)
					if n_old > 0 {
						sum_n_logn += n_old * math.Log(n_old)
					}
					// Add contribution of new word (after increment)
					if n_new > 0 {
						sum_n_logn += n_new * math.Log(n_new)
					}
					// Remove contribution of new word (before increment)
					if n_new > 1 {
						sum_n_logn -= (n_new - 1) * math.Log(n_new-1)
					}
				}

				entropy = (log_nwords - sum_n_logn/float_nwords) / entropyMax
			}

			// Store entropy for position corresponding to start of k-mer
			if s >= nwords && maskPositions[i-nwords+1] >= 0 {
				if entropy < 0 {
					entropy = 0

				}
				entropy = math.Round(entropy*10000) / 10000
				entropies[i-nwords+1] = entropy
			}
		}

		// ========================================================================
		// STEP 3: Apply sliding min filter
		// ========================================================================
		// Replace each entropy with minimum in window of size kmer_size
		// This allows robust detection of low-complexity regions
		slidingMin(entropies, kmer_size)
		// log.Warnf("%v\n%v", e, entropies)
	}

	// ========================================================================
	// FUNCTION 6: applyMaskMode - Apply masking to sequence
	// ========================================================================
	applyMaskMode := func(sequence *obiseq.BioSequence, maskPositions []bool, mask byte) (obiseq.BioSequenceSlice, error) {
		// Create copy to avoid modifying original
		seqCopy := sequence.Copy()
		sequenceBytes := seqCopy.Sequence()

		// Mask identified positions
		for i := range sequenceBytes {
			if maskPositions[i] {
				// Operation &^ 32 converts to UPPERCASE (clears bit 5)
				// sequenceBytes[i] = sequenceBytes[i] &^ 32
				sequenceBytes[i] = mask
			}
		}

		return obiseq.BioSequenceSlice{seqCopy}, nil
	}

	selectMasked := func(sequence *obiseq.BioSequence, maskPosition []bool) (obiseq.BioSequenceSlice, error) {
		rep := obiseq.NewBioSequenceSlice()

		inlow := false
		fromlow := -1
		for i, masked := range maskPosition {
			if masked && !inlow {
				fromlow = i
				inlow = true
			}
			if inlow && !masked {
				if fromlow >= 0 {
					frg, err := sequence.Subsequence(fromlow, i, false)
					if err != nil {
						return nil, err
					}
					rep.Push(frg)
				}
				inlow = false
				fromlow = -1
			}
		}

		// Handle the case where we end in a masked region
		if inlow && fromlow >= 0 {
			frg, err := sequence.Subsequence(fromlow, len(maskPosition), false)
			if err != nil {
				return nil, err
			}
			rep.Push(frg)
		}

		return *rep, nil
	}

	selectunmasked := func(sequence *obiseq.BioSequence, maskPosition []bool) (obiseq.BioSequenceSlice, error) {
		rep := obiseq.NewBioSequenceSlice()

		inhigh := false
		fromhigh := -1
		for i, masked := range maskPosition {
			if !masked && !inhigh {
				fromhigh = i
				inhigh = true
			}
			if inhigh && masked {
				if fromhigh >= 0 {
					frg, err := sequence.Subsequence(fromhigh, i, false)
					if err != nil {
						return nil, err
					}
					rep.Push(frg)
				}
				inhigh = false
				fromhigh = -1
			}
		}

		// Handle the case where we end in an unmasked region
		if inhigh && fromhigh >= 0 {
			frg, err := sequence.Subsequence(fromhigh, len(maskPosition), false)
			if err != nil {
				return nil, err
			}
			rep.Push(frg)
		}

		return *rep, nil
	}

	// ========================================================================
	// FUNCTION 7: masking - Main masking function
	// ========================================================================
	// Calculates entropies at all scales and masks positions
	// whose minimum entropy is below the threshold.
	masking := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		if sequence.Len() < kmer_size {
			sequence.SetAttribute("obilowmask_error", "Sequence too short")
			remove := make([]bool, sequence.Len())
			for i := range remove {
				remove[i] = true
			}
			return applyMaskMode(sequence, remove, maskChar)
		}

		bseq := sequence.Sequence()

		// Identify ambiguities
		maskPositions := maskAmbiguities(bseq)

		// Initialize data structures
		mask := make([]int, len(bseq))          // Stores scale detecting minimum entropy
		entropies := make([]float64, len(bseq)) // Minimum entropy at each position
		for i := range entropies {
			entropies[i] = 4.0 // Very high initial value
		}

		freqs := make([]int, 1<<(2*level_max)) // Frequency table (max size)
		words := make([]int, len(bseq))        // Buffer for encoded words

		// ========================================================================
		// Calculate entropy at maximum scale (level_max)
		// ========================================================================
		computeEntropies(bseq, maskPositions, entropies, freqs, words, level_max)

		// Initialize mask with level_max everywhere (except ambiguities)
		for i := range bseq {
			v := level_max
			// if nuc != 'a' && nuc != 'c' && nuc != 'g' && nuc != 't' {
			// 	v = 0
			// }
			mask[i] = v
		}

		// ========================================================================
		// Calculate entropy at lower scales
		// ========================================================================
		entropies2 := make([]float64, len(bseq))

		for ws := level_max - 1; ws > 0; ws-- {
			// *** WARNING: POTENTIAL BUG ***
			// The parameter passed is level_max instead of ws!
			// This means we always recalculate with the same scale
			// Should be: computeEntropies(bseq, maskPositions, entropies2, freqs, words, ws)
			computeEntropies(bseq, maskPositions, entropies2, freqs, words, ws)
			// Keep minimum entropy and corresponding scale
			for i, e2 := range entropies2 {
				if e2 < entropies[i] {
					entropies[i] = e2
					mask[i] = ws
				}
			}
		}

		// Force entropy to 0 for ambiguous positions
		for i, nuc := range bseq {
			if nuc != 'a' && nuc != 'c' && nuc != 'g' && nuc != 't' {
				entropies[i] = 0
			}
		}

		// ========================================================================
		// Identify positions to mask
		// ========================================================================
		remove := make([]bool, len(entropies))
		for i, e := range entropies {
			remove[i] = e <= threshold
		}

		// Save metadata in sequence attributes
		sequence.SetAttribute("mask", mask)
		sequence.SetAttribute("Entropies", entropies)

		switch mode {
		case Mask:
			return applyMaskMode(sequence, remove, maskChar)
		case Split:
			return selectunmasked(sequence, remove)
		case Extract:
			return selectMasked(sequence, remove)
		}
		return nil, fmt.Errorf("Unknown mode %d", mode)
	}

	return masking
}

// CLISequenceEntropyMasker creates an iterator that applies entropy masking
// to all sequences in an input iterator.
//
// Uses command-line parameters to configure the worker.
func CLISequenceEntropyMasker(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	var newIter obiiter.IBioSequence

	worker := LowMaskWorker(
		CLIKmerSize(),
		CLILevelMax(),
		CLIThreshold(),
		CLIMaskingMode(),
		CLIMaskingChar(),
	)

	// Apply worker in parallel
	newIter = iterator.MakeIWorker(worker, false, obidefault.ParallelWorkers())

	// Filter resulting empty sequences
	return newIter.FilterEmpty()
}
