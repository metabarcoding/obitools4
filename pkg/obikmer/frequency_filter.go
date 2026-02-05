package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/RoaringBitmap/roaring/roaring64"
)

// FrequencyFilter filtre les k-mers par fréquence minimale
// Utilise v bitmaps où index[i] contient les k-mers vus au moins i+1 fois
type FrequencyFilter struct {
	K           int
	MinFreq     int                    // v - fréquence minimale requise
	index       []*roaring64.Bitmap    // index[i] = k-mers vus ≥(i+1) fois
}

// NewFrequencyFilter crée un nouveau filtre par fréquence
// minFreq: nombre minimum d'occurrences requises (v)
func NewFrequencyFilter(k, minFreq int) *FrequencyFilter {
	if minFreq < 1 {
		panic("minFreq must be >= 1")
	}

	// Créer v bitmaps
	bitmaps := make([]*roaring64.Bitmap, minFreq)
	for i := range bitmaps {
		bitmaps[i] = roaring64.New()
	}

	return &FrequencyFilter{
		K:       k,
		MinFreq: minFreq,
		index:   bitmaps,
	}
}

// AddSequence ajoute tous les k-mers d'une séquence au filtre
// Utilise un itérateur pour éviter l'allocation d'un vecteur intermédiaire
func (ff *FrequencyFilter) AddSequence(seq *obiseq.BioSequence) {
	rawSeq := seq.Sequence()
	for canonical := range IterNormalizedKmers(rawSeq, ff.K) {
		ff.addKmer(canonical)
	}
}

// addKmer ajoute un k-mer au filtre (algorithme principal)
func (ff *FrequencyFilter) addKmer(kmer uint64) {
	// Trouver le niveau actuel du k-mer
	c := 0
	for c < ff.MinFreq && ff.index[c].Contains(kmer) {
		c++
	}

	// Ajouter au niveau suivant (si pas encore au maximum)
	if c < ff.MinFreq {
		ff.index[c].Add(kmer)
	}
}

// GetFilteredSet retourne un KmerSet des k-mers avec fréquence ≥ minFreq
func (ff *FrequencyFilter) GetFilteredSet() *KmerSet {
	// Les k-mers filtrés sont dans le dernier niveau
	return NewKmerSetFromBitmap(ff.K, ff.index[ff.MinFreq-1].Clone())
}

// GetKmersAtLevel retourne un KmerSet des k-mers vus au moins (level+1) fois
// level doit être dans [0, minFreq-1]
func (ff *FrequencyFilter) GetKmersAtLevel(level int) *KmerSet {
	if level < 0 || level >= ff.MinFreq {
		return NewKmerSet(ff.K)
	}

	return NewKmerSetFromBitmap(ff.K, ff.index[level].Clone())
}

// Stats retourne des statistiques sur les niveaux de fréquence
func (ff *FrequencyFilter) Stats() FrequencyFilterStats {
	stats := FrequencyFilterStats{
		MinFreq: ff.MinFreq,
		Levels:  make([]LevelStats, ff.MinFreq),
	}

	for i := 0; i < ff.MinFreq; i++ {
		card := ff.index[i].GetCardinality()
		sizeBytes := ff.index[i].GetSizeInBytes()

		stats.Levels[i] = LevelStats{
			Level:       i + 1, // Niveau 1 = freq ≥ 1
			Cardinality: card,
			SizeBytes:   sizeBytes,
		}

		stats.TotalBytes += sizeBytes
	}

	// Le dernier niveau contient le résultat
	stats.FilteredKmers = stats.Levels[ff.MinFreq-1].Cardinality

	return stats
}

// FrequencyFilterStats contient les statistiques du filtre
type FrequencyFilterStats struct {
	MinFreq       int
	FilteredKmers uint64      // K-mers avec freq ≥ minFreq
	TotalBytes    uint64      // Mémoire totale utilisée
	Levels        []LevelStats
}

// LevelStats contient les stats d'un niveau
type LevelStats struct {
	Level       int    // freq ≥ Level
	Cardinality uint64 // Nombre de k-mers
	SizeBytes   uint64 // Taille en bytes
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
func (ff *FrequencyFilter) Clear() {
	for _, bitmap := range ff.index {
		bitmap.Clear()
	}
}

// ==================================
// BATCH PROCESSING
// ==================================

// AddSequences ajoute plusieurs séquences en batch
func (ff *FrequencyFilter) AddSequences(sequences *obiseq.BioSequenceSlice) {
	for _, seq := range *sequences {
		ff.AddSequence(seq)
	}
}

// ==================================
// PERSISTANCE
// ==================================

// Save sauvegarde le filtre sur disque
func (ff *FrequencyFilter) Save(path string) error {
	// TODO: implémenter la sérialisation
	// Pour chaque bitmap: bitmap.WriteTo(writer)
	return nil
}

// Load charge le filtre depuis le disque
func (ff *FrequencyFilter) Load(path string) error {
	// TODO: implémenter la désérialisation
	return nil
}

// ==================================
// UTILITAIRES
// ==================================

// Contains vérifie si un k-mer a atteint la fréquence minimale
func (ff *FrequencyFilter) Contains(kmer uint64) bool {
	canonical := NormalizeKmer(kmer, ff.K)
	return ff.index[ff.MinFreq-1].Contains(canonical)
}

// GetFrequency retourne la fréquence approximative d'un k-mer
// Retourne le niveau maximum atteint (freq ≥ niveau)
func (ff *FrequencyFilter) GetFrequency(kmer uint64) int {
	canonical := NormalizeKmer(kmer, ff.K)

	freq := 0
	for i := 0; i < ff.MinFreq; i++ {
		if ff.index[i].Contains(canonical) {
			freq = i + 1
		} else {
			break
		}
	}

	return freq
}

// Len retourne le nombre de k-mers filtrés ou à un niveau spécifique
// Sans argument: retourne le nombre de k-mers avec freq ≥ minFreq (dernier niveau)
// Avec argument level: retourne le nombre de k-mers avec freq ≥ (level+1)
// Exemple: Len() pour les k-mers filtrés, Len(2) pour freq ≥ 3
func (ff *FrequencyFilter) Len(level ...int) uint64 {
	if len(level) == 0 {
		// Sans argument: dernier niveau (k-mers filtrés)
		return ff.index[ff.MinFreq-1].GetCardinality()
	}

	// Avec argument: niveau spécifique
	lvl := level[0]
	if lvl < 0 || lvl >= ff.MinFreq {
		return 0
	}
	return ff.index[lvl].GetCardinality()
}

// MemoryUsage retourne l'utilisation mémoire en bytes
func (ff *FrequencyFilter) MemoryUsage() uint64 {
	total := uint64(0)
	for _, bitmap := range ff.index {
		total += bitmap.GetSizeInBytes()
	}
	return total
}

// ==================================
// COMPARAISON AVEC D'AUTRES APPROCHES
// ==================================

// CompareWithSimpleMap compare la mémoire avec une simple map
func (ff *FrequencyFilter) CompareWithSimpleMap() string {
	totalKmers := ff.index[0].GetCardinality()

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
