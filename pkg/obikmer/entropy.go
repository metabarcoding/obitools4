package obikmer

import "math"

// KmerEntropy computes the entropy of a single encoded k-mer.
//
// The algorithm mirrors the lowmask entropy calculation: it decodes the k-mer
// to a DNA sequence, extracts all sub-words of each size from 1 to levelMax,
// normalizes them by circular canonical form, counts their frequencies, and
// computes Shannon entropy normalized by the maximum possible entropy.
// The returned value is the minimum entropy across all word sizes.
//
// A value close to 0 indicates very low complexity (e.g. "AAAA..."),
// while a value close to 1 indicates high complexity.
//
// Parameters:
//   - kmer: the encoded k-mer (2 bits per base)
//   - k: the k-mer size
//   - levelMax: maximum sub-word size for entropy (typically 6)
//
// Returns:
//   - minimum normalized entropy across all word sizes 1..levelMax
func KmerEntropy(kmer uint64, k int, levelMax int) float64 {
	if k < 1 || levelMax < 1 {
		return 1.0
	}
	if levelMax >= k {
		levelMax = k - 1
	}
	if levelMax < 1 {
		return 1.0
	}

	// Decode k-mer to DNA sequence
	var seqBuf [32]byte
	seq := DecodeKmer(kmer, k, seqBuf[:])

	// Pre-compute nLogN lookup (same as lowmask)
	nLogN := make([]float64, k+1)
	for i := 1; i <= k; i++ {
		nLogN[i] = float64(i) * math.Log(float64(i))
	}

	// Build circular-canonical normalization tables per word size
	normTables := make([][]int, levelMax+1)
	for ws := 1; ws <= levelMax; ws++ {
		size := 1 << (ws * 2)
		normTables[ws] = make([]int, size)
		for code := 0; code < size; code++ {
			normTables[ws][code] = int(NormalizeCircular(uint64(code), ws))
		}
	}

	minEntropy := math.MaxFloat64

	for ws := 1; ws <= levelMax; ws++ {
		nwords := k - ws + 1
		if nwords < 1 {
			continue
		}

		// Count circular-canonical sub-word frequencies
		tableSize := 1 << (ws * 2)
		table := make([]int, tableSize)
		mask := (1 << (ws * 2)) - 1

		wordIndex := 0
		for i := 0; i < ws-1; i++ {
			wordIndex = (wordIndex << 2) + int(EncodeNucleotide(seq[i]))
		}

		for i, j := 0, ws-1; j < k; i, j = i+1, j+1 {
			wordIndex = ((wordIndex << 2) & mask) + int(EncodeNucleotide(seq[j]))
			normWord := normTables[ws][wordIndex]
			table[normWord]++
		}

		// Compute Shannon entropy
		floatNwords := float64(nwords)
		logNwords := math.Log(floatNwords)

		var sumNLogN float64
		for j := 0; j < tableSize; j++ {
			n := table[j]
			if n > 0 {
				sumNLogN += nLogN[n]
			}
		}

		// Compute emax (maximum possible entropy for this word size)
		na := CanonicalCircularKmerCount(ws)
		var emax float64
		if nwords < na {
			emax = math.Log(float64(nwords))
		} else {
			cov := nwords / na
			remains := nwords - (na * cov)
			f1 := float64(cov) / floatNwords
			f2 := float64(cov+1) / floatNwords
			emax = -(float64(na-remains)*f1*math.Log(f1) +
				float64(remains)*f2*math.Log(f2))
		}

		if emax <= 0 {
			continue
		}

		entropy := (logNwords - sumNLogN/floatNwords) / emax
		if entropy < 0 {
			entropy = 0
		}

		if entropy < minEntropy {
			minEntropy = entropy
		}
	}

	if minEntropy == math.MaxFloat64 {
		return 1.0
	}

	return math.Round(minEntropy*10000) / 10000
}

// KmerEntropyFilter is a reusable entropy filter for batch processing.
// It pre-computes normalization tables and lookup values to avoid repeated
// allocation across millions of k-mers.
//
// IMPORTANT: a KmerEntropyFilter is NOT safe for concurrent use.
// Each goroutine must create its own instance via NewKmerEntropyFilter.
type KmerEntropyFilter struct {
	k          int
	levelMax   int
	threshold  float64
	nLogN      []float64
	normTables [][]int
	emaxValues []float64
	logNwords  []float64
	// Pre-allocated frequency tables reused across Entropy() calls.
	// One per word size (index 0 unused). Reset to zero before each use.
	freqTables [][]int
}

// NewKmerEntropyFilter creates an entropy filter with pre-computed tables.
//
// Parameters:
//   - k: the k-mer size
//   - levelMax: maximum sub-word size for entropy (typically 6)
//   - threshold: entropy threshold (k-mers with entropy <= threshold are rejected)
func NewKmerEntropyFilter(k, levelMax int, threshold float64) *KmerEntropyFilter {
	if levelMax >= k {
		levelMax = k - 1
	}
	if levelMax < 1 {
		levelMax = 1
	}

	nLogN := make([]float64, k+1)
	for i := 1; i <= k; i++ {
		nLogN[i] = float64(i) * math.Log(float64(i))
	}

	normTables := make([][]int, levelMax+1)
	for ws := 1; ws <= levelMax; ws++ {
		size := 1 << (ws * 2)
		normTables[ws] = make([]int, size)
		for code := 0; code < size; code++ {
			normTables[ws][code] = int(NormalizeCircular(uint64(code), ws))
		}
	}

	emaxValues := make([]float64, levelMax+1)
	logNwords := make([]float64, levelMax+1)
	for ws := 1; ws <= levelMax; ws++ {
		nw := k - ws + 1
		na := CanonicalCircularKmerCount(ws)
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

	// Pre-allocate frequency tables per word size
	freqTables := make([][]int, levelMax+1)
	for ws := 1; ws <= levelMax; ws++ {
		freqTables[ws] = make([]int, 1<<(ws*2))
	}

	return &KmerEntropyFilter{
		k:          k,
		levelMax:   levelMax,
		threshold:  threshold,
		nLogN:      nLogN,
		normTables: normTables,
		emaxValues: emaxValues,
		logNwords:  logNwords,
		freqTables: freqTables,
	}
}

// Accept returns true if the k-mer has entropy strictly above the threshold.
// Low-complexity k-mers (entropy <= threshold) are rejected.
func (ef *KmerEntropyFilter) Accept(kmer uint64) bool {
	return ef.Entropy(kmer) > ef.threshold
}

// Entropy computes the entropy for a single k-mer using pre-computed tables.
func (ef *KmerEntropyFilter) Entropy(kmer uint64) float64 {
	k := ef.k

	// Decode k-mer to DNA sequence
	var seqBuf [32]byte
	seq := DecodeKmer(kmer, k, seqBuf[:])

	minEntropy := math.MaxFloat64

	for ws := 1; ws <= ef.levelMax; ws++ {
		nwords := k - ws + 1
		if nwords < 1 {
			continue
		}

		emax := ef.emaxValues[ws]
		if emax <= 0 {
			continue
		}

		// Count circular-canonical sub-word frequencies
		tableSize := 1 << (ws * 2)
		table := ef.freqTables[ws]
		clear(table) // reset to zero
		mask := (1 << (ws * 2)) - 1
		normTable := ef.normTables[ws]

		wordIndex := 0
		for i := 0; i < ws-1; i++ {
			wordIndex = (wordIndex << 2) + int(EncodeNucleotide(seq[i]))
		}

		for i, j := 0, ws-1; j < k; i, j = i+1, j+1 {
			wordIndex = ((wordIndex << 2) & mask) + int(EncodeNucleotide(seq[j]))
			normWord := normTable[wordIndex]
			table[normWord]++
		}

		// Compute Shannon entropy
		floatNwords := float64(nwords)
		logNwords := ef.logNwords[ws]

		var sumNLogN float64
		for j := 0; j < tableSize; j++ {
			n := table[j]
			if n > 0 {
				sumNLogN += ef.nLogN[n]
			}
		}

		entropy := (logNwords - sumNLogN/floatNwords) / emax
		if entropy < 0 {
			entropy = 0
		}

		if entropy < minEntropy {
			minEntropy = entropy
		}
	}

	if minEntropy == math.MaxFloat64 {
		return 1.0
	}

	return math.Round(minEntropy*10000) / 10000
}
