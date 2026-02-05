package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// FrequencyFilter filters k-mers by minimum frequency
// Specialization of KmerSetGroup where index[i] contains k-mers seen at least i+1 times
type FrequencyFilter struct {
	*KmerSetGroup          // Group of KmerSet (one per frequency level)
	MinFreq       int      // v - minimum required frequency
}

// NewFrequencyFilter creates a new frequency filter
// minFreq: minimum number d'occurrences required (v)
func NewFrequencyFilter(k, minFreq int) *FrequencyFilter {
	ff := &FrequencyFilter{
		KmerSetGroup: NewKmerSetGroup(k, minFreq),
		MinFreq:      minFreq,
	}

	// Initialize group metadata
	ff.SetAttribute("type", "FrequencyFilter")
	ff.SetAttribute("min_freq", minFreq)

	// Initialize metadata for each level
	for i := 0; i < minFreq; i++ {
		level := ff.Get(i)
		level.SetAttribute("level", i)
		level.SetAttribute("min_occurrences", i+1)
		level.SetId(fmt.Sprintf("level_%d", i))
	}

	return ff
}

// AddSequence adds all k-mers from a sequence to the filter
// Uses an iterator to avoid allocating an intermediate vector
func (ff *FrequencyFilter) AddSequence(seq *obiseq.BioSequence) {
	rawSeq := seq.Sequence()
	for canonical := range IterCanonicalKmers(rawSeq, ff.K()) {
		ff.AddKmerCode(canonical)
	}
}

// AddKmerCode adds an encoded k-mer to the filter (main algorithm)
func (ff *FrequencyFilter) AddKmerCode(kmer uint64) {
	// Find the current level of the k-mer
	c := 0
	for c < ff.MinFreq && ff.Get(c).Contains(kmer) {
		c++
	}

	// Add to next level (if not yet at maximum)
	if c < ff.MinFreq {
		ff.Get(c).AddKmerCode(kmer)
	}
}

// AddCanonicalKmerCode adds an encoded canonical k-mer to the filter
func (ff *FrequencyFilter) AddCanonicalKmerCode(kmer uint64) {
	canonical := CanonicalKmer(kmer, ff.K())
	ff.AddKmerCode(canonical)
}

// AddKmer adds a k-mer to the filter by encoding the sequence
// The sequence must have exactly k nucleotides
// Zero-allocation: encodes directly without creating an intermediate slice
func (ff *FrequencyFilter) AddKmer(seq []byte) {
	kmer := EncodeKmer(seq, ff.K())
	ff.AddKmerCode(kmer)
}

// AddCanonicalKmer adds a canonical k-mer to the filter by encoding the sequence
// The sequence must have exactly k nucleotides
// Zero-allocation: encodes directly in canonical form without creating an intermediate slice
func (ff *FrequencyFilter) AddCanonicalKmer(seq []byte) {
	canonical := EncodeCanonicalKmer(seq, ff.K())
	ff.AddKmerCode(canonical)
}

// GetFilteredSet returns a KmerSet of k-mers with frequency ≥ minFreq
func (ff *FrequencyFilter) GetFilteredSet() *KmerSet {
	// Filtered k-mers are in the last level
	return ff.Get(ff.MinFreq - 1).Copy()
}

// GetKmersAtLevel returns a KmerSet of k-mers seen at least (level+1) times
// level doit être dans [0, minFreq-1]
func (ff *FrequencyFilter) GetKmersAtLevel(level int) *KmerSet {
	ks := ff.Get(level)
	if ks == nil {
		return NewKmerSet(ff.K())
	}
	return ks.Copy()
}

// Stats returns statistics on frequency levels
func (ff *FrequencyFilter) Stats() FrequencyFilterStats {
	stats := FrequencyFilterStats{
		MinFreq: ff.MinFreq,
		Levels:  make([]LevelStats, ff.MinFreq),
	}

	for i := 0; i < ff.MinFreq; i++ {
		ks := ff.Get(i)
		card := ks.Len()
		sizeBytes := ks.MemoryUsage()

		stats.Levels[i] = LevelStats{
			Level:       i + 1, // Level 1 = freq ≥ 1
			Cardinality: card,
			SizeBytes:   sizeBytes,
		}

		stats.TotalBytes += sizeBytes
	}

	// The last level contains the result
	stats.FilteredKmers = stats.Levels[ff.MinFreq-1].Cardinality

	return stats
}

// FrequencyFilterStats contains the filter statistics
type FrequencyFilterStats struct {
	MinFreq       int
	FilteredKmers uint64      // K-mers with freq ≥ minFreq
	TotalBytes    uint64      // Total memory used
	Levels        []LevelStats
}

// LevelStats contains the stats of a level
type LevelStats struct {
	Level       int    // freq ≥ Level
	Cardinality uint64 // Number of k-mers
	SizeBytes   uint64 // Size in bytes
}

func (ffs FrequencyFilterStats) String() string {
	result := fmt.Sprintf(`Frequency Filter Statistics (minFreq=%d):
  Filtered k-mers (freq≥%d): %d
  Total memory: %.2f MB

Level breakdown:
`, ffs.MinFreq, ffs.MinFreq, ffs.FilteredKmers, float64(ffs.TotalBytes)/1024/1024)

	for _, level := range ffs.Levels {
		result += fmt.Sprintf("  freq≥%d: %d k-mers (%.2f MB)\n",
			level.Level,
			level.Cardinality,
			float64(level.SizeBytes)/1024/1024)
	}

	return result
}

// Clear libère la mémoire de tous les niveaux
// (héritée de KmerSetGroup mais redéfinie pour clarté)
func (ff *FrequencyFilter) Clear() {
	ff.KmerSetGroup.Clear()
}

// ==================================
// BATCH PROCESSING
// ==================================

// AddSequences adds multiple sequences in batch
func (ff *FrequencyFilter) AddSequences(sequences *obiseq.BioSequenceSlice) {
	for _, seq := range *sequences {
		ff.AddSequence(seq)
	}
}

// ==================================
// PERSISTANCE
// ==================================

// Save sauvegarde le FrequencyFilter dans un répertoire
// Utilise le format de sérialisation du KmerSetGroup sous-jacent
// Les métadonnées incluent le type "FrequencyFilter" et min_freq
//
// Format:
//   - directory/metadata.{toml,yaml,json} - métadonnées du filtre
//   - directory/set_0.roaring - k-mers vus ≥1 fois
//   - directory/set_1.roaring - k-mers vus ≥2 fois
//   - ...
//   - directory/set_{minFreq-1}.roaring - k-mers vus ≥minFreq fois
//
// Parameters:
//   - directory: répertoire de destination
//   - format: format des métadonnées (FormatTOML, FormatYAML, FormatJSON)
//
// Example:
//
//	err := ff.Save("./my_filter", obikmer.FormatTOML)
func (ff *FrequencyFilter) Save(directory string, format MetadataFormat) error {
	// Déléguer à KmerSetGroup qui gère déjà tout
	return ff.KmerSetGroup.Save(directory, format)
}

// LoadFrequencyFilter charge un FrequencyFilter depuis un répertoire
// Vérifie que les métadonnées correspondent à un FrequencyFilter
//
// Parameters:
//   - directory: répertoire source
//
// Returns:
//   - *FrequencyFilter: le filtre chargé
//   - error: erreur si le chargement échoue ou si ce n'est pas un FrequencyFilter
//
// Example:
//
//	ff, err := obikmer.LoadFrequencyFilter("./my_filter")
func LoadFrequencyFilter(directory string) (*FrequencyFilter, error) {
	// Charger le KmerSetGroup
	ksg, err := LoadKmerSetGroup(directory)
	if err != nil {
		return nil, err
	}

	// Vérifier que c'est bien un FrequencyFilter
	if typeAttr, ok := ksg.GetAttribute("type"); !ok || typeAttr != "FrequencyFilter" {
		return nil, fmt.Errorf("loaded data is not a FrequencyFilter (type=%v)", typeAttr)
	}

	// Récupérer min_freq
	minFreqAttr, ok := ksg.GetIntAttribute("min_freq")
	if !ok {
		return nil, fmt.Errorf("FrequencyFilter missing min_freq attribute")
	}

	// Créer le FrequencyFilter
	ff := &FrequencyFilter{
		KmerSetGroup: ksg,
		MinFreq:      minFreqAttr,
	}

	return ff, nil
}

// ==================================
// UTILITAIRES
// ==================================

// Contains vérifie si un k-mer a atteint la fréquence minimale
func (ff *FrequencyFilter) Contains(kmer uint64) bool {
	canonical := CanonicalKmer(kmer, ff.K())
	return ff.Get(ff.MinFreq - 1).Contains(canonical)
}

// GetFrequency returns the approximate frequency of a k-mer
// Retourne le niveau maximum atteint (freq ≥ niveau)
func (ff *FrequencyFilter) GetFrequency(kmer uint64) int {
	canonical := CanonicalKmer(kmer, ff.K())

	freq := 0
	for i := 0; i < ff.MinFreq; i++ {
		if ff.Get(i).Contains(canonical) {
			freq = i + 1
		} else {
			break
		}
	}

	return freq
}

// Len returns the number of filtered k-mers or at a specific level
// Without argument: returns the number of k-mers with freq ≥ minFreq (last level)
// With argument level: returns the number of k-mers with freq ≥ (level+1)
// Exemple: Len() pour les k-mers filtrés, Len(2) pour freq ≥ 3
// (héritée de KmerSetGroup mais redéfinie pour la documentation)
func (ff *FrequencyFilter) Len(level ...int) uint64 {
	return ff.KmerSetGroup.Len(level...)
}

// MemoryUsage returns memory usage in bytes
// (héritée de KmerSetGroup mais redéfinie pour clarté)
func (ff *FrequencyFilter) MemoryUsage() uint64 {
	return ff.KmerSetGroup.MemoryUsage()
}

// ==================================
// COMPARAISON AVEC D'AUTRES APPROCHES
// ==================================

// CompareWithSimpleMap compare la mémoire avec une simple map
func (ff *FrequencyFilter) CompareWithSimpleMap() string {
	totalKmers := ff.Get(0).Len()

	simpleMapBytes := totalKmers * 24 // ~24 bytes par entrée
	roaringBytes := ff.MemoryUsage()

	reduction := float64(simpleMapBytes) / float64(roaringBytes)

	return fmt.Sprintf(`Memory Comparison for %d k-mers:
  Simple map[uint64]uint32: %.2f MB
  Roaring filter (v=%d):    %.2f MB
  Reduction:                %.1fx
`,
		totalKmers,
		float64(simpleMapBytes)/1024/1024,
		ff.MinFreq,
		float64(roaringBytes)/1024/1024,
		reduction,
	)
}
