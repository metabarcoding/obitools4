package obilowmask

import (
	"fmt"
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
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

	nLogN := make([]float64, kmer_size+1)
	for i := 1; i <= kmer_size; i++ {
		nLogN[i] = float64(i) * math.Log(float64(i))
	}

	normTables := make([][]int, level_max+1)
	for ws := 1; ws <= level_max; ws++ {
		size := 1 << (ws * 2)
		normTables[ws] = make([]int, size)
		for code := 0; code < size; code++ {
			normTables[ws][code] = int(obikmer.NormalizeCircular(uint64(code), ws))
		}
	}

	type pair struct {
		index int
		value float64
	}

	slidingMin := func(data []float64, window int) {
		if len(data) == 0 || window <= 0 {
			return
		}
		if window >= len(data) {
			minVal := data[0]
			for i := 1; i < len(data); i++ {
				if data[i] < minVal {
					minVal = data[i]
				}
			}
			for i := range data {
				data[i] = minVal
			}
			return
		}

		deque := make([]pair, 0, window)

		for i, v := range data {
			for len(deque) > 0 && deque[0].index <= i-window {
				deque = deque[1:]
			}

			for len(deque) > 0 && deque[len(deque)-1].value >= v {
				deque = deque[:len(deque)-1]
			}

			deque = append(deque, pair{index: i, value: v})

			data[i] = deque[0].value
		}
	}
	emaxValues := make([]float64, level_max+1)
	logNwords := make([]float64, level_max+1)
	for ws := 1; ws <= level_max; ws++ {
		nw := kmer_size - ws + 1
		na := obikmer.CanonicalCircularKmerCount(ws)
		if nw < na {
			logNwords[ws] = math.Log(float64(nw))
			emaxValues[ws] = math.Log(float64(nw))
		} else {
			cov := nw / na
			remains := nw - (na * cov)
			f1 := float64(cov) / float64(nw)
			f2 := float64(cov+1) / float64(nw)
			logNwords[ws] = math.Log(float64(nw))
			emaxValues[ws] = -(float64(na-remains)*f1*math.Log(f1) +
				float64(remains)*f2*math.Log(f2))
		}
	}

	maskAmbiguities := func(sequence []byte) []int {
		maskPositions := make([]int, len(sequence))
		for i, nuc := range sequence {
			if nuc != 'a' && nuc != 'c' && nuc != 'g' && nuc != 't' {
				end := max(0, i-kmer_size+1)
				for j := i; j >= end; j-- {
					maskPositions[j] = -1
				}
			}
		}
		return maskPositions
	}

	cleanTable := func(table []int, over int) {
		for i := 0; i < over; i++ {
			table[i] = 0
		}
	}

	computeEntropies := func(sequence []byte,
		maskPositions []int,
		entropies []float64,
		table []int,
		words []int,
		wordSize int,
		normTable []int) {

		lseq := len(sequence)
		tableSize := 1 << (wordSize * 2)
		nwords := kmer_size - wordSize + 1
		float_nwords := float64(nwords)
		log_nwords := logNwords[wordSize]
		entropyMax := emaxValues[wordSize]

		cleanTable(table, tableSize)

		for i := 1; i < lseq; i++ {
			entropies[i] = 6
		}
		end := lseq - wordSize + 1

		mask := (1 << (wordSize * 2)) - 1

		word_index := 0
		for i := 0; i < wordSize-1; i++ {
			word_index = (word_index << 2) + int(obikmer.EncodeNucleotide(sequence[i]))
		}

		for i, j := 0, wordSize-1; i < end; i, j = i+1, j+1 {
			word_index = ((word_index << 2) & mask) + int(obikmer.EncodeNucleotide(sequence[j]))
			words[i] = normTable[word_index]
		}

		s := 0
		sum_n_logn := 0.0
		entropy := 1.0
		cleaned := true

		for i := range end {
			s++

			switch {
			case s < nwords:
				cleaned = false
				table[words[i]]++

			case i >= (nwords-1) && maskPositions[i-nwords+1] < 0:
				entropies[i-nwords+1] = 4.0
				if !cleaned {
					cleanTable(table, tableSize)
				}
				cleaned = true
				s = 0
				sum_n_logn = 0.0

			case s == nwords:
				cleaned = false
				table[words[i]]++

				sum_n_logn = 0
				for j := range tableSize {
					n := float64(table[j])
					if n > 0 {
						sum_n_logn += nLogN[int(n)]
					}
				}
				entropy = (log_nwords - sum_n_logn/float_nwords) / entropyMax

			case s > nwords:
				cleaned = false

				new_word := words[i]
				old_word := words[i-nwords]

				if old_word != new_word {
					table[new_word]++
					table[old_word]--

					n_old := float64(table[old_word])
					n_new := float64(table[new_word])

					sum_n_logn -= nLogN[int(n_old+1)]
					if n_old > 0 {
						sum_n_logn += nLogN[int(n_old)]
					}
					if n_new > 0 {
						sum_n_logn += nLogN[int(n_new)]
					}
					if n_new > 1 {
						sum_n_logn -= nLogN[int(n_new-1)]
					}
				}

				entropy = (log_nwords - sum_n_logn/float_nwords) / entropyMax
			}

			if s >= nwords && maskPositions[i-nwords+1] >= 0 {
				if entropy < 0 {
					entropy = 0
				}
				entropy = math.Round(entropy*10000) / 10000
				entropies[i-nwords+1] = entropy
			}
		}

		slidingMin(entropies, kmer_size)
	}

	applyMaskMode := func(sequence *obiseq.BioSequence, maskPositions []bool, mask byte) (obiseq.BioSequenceSlice, error) {
		seqCopy := sequence.Copy()
		sequenceBytes := seqCopy.Sequence()

		for i := range sequenceBytes {
			if maskPositions[i] {
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

		if inhigh && fromhigh >= 0 {
			frg, err := sequence.Subsequence(fromhigh, len(maskPosition), false)
			if err != nil {
				return nil, err
			}
			rep.Push(frg)
		}

		return *rep, nil
	}

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

		maskPositions := maskAmbiguities(bseq)

		mask := make([]int, len(bseq))
		entropies := make([]float64, len(bseq))
		for i := range entropies {
			entropies[i] = 4.0
		}

		freqs := make([]int, 1<<(2*level_max))
		words := make([]int, len(bseq))
		entropies2 := make([]float64, len(bseq))

		computeEntropies(bseq, maskPositions, entropies, freqs, words, level_max, normTables[level_max])

		for i := range bseq {
			v := level_max
			mask[i] = v
		}

		for ws := level_max - 1; ws > 0; ws-- {
			computeEntropies(bseq, maskPositions, entropies2, freqs, words, ws, normTables[ws])
			for i, e2 := range entropies2 {
				if e2 < entropies[i] {
					entropies[i] = e2
					mask[i] = ws
				}
			}
		}

		for i, nuc := range bseq {
			if nuc != 'a' && nuc != 'c' && nuc != 'g' && nuc != 't' {
				entropies[i] = 0
			}
		}

		remove := make([]bool, len(entropies))
		for i, e := range entropies {
			remove[i] = e <= threshold
		}

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

	newIter = iterator.MakeIWorker(worker, false, obidefault.ParallelWorkers())

	return newIter.FilterEmpty()
}
