package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// FrequencyFilter filtre les k-mers par fréquence minimale
// Spécialisation de KmerSetGroup où index[i] contient les k-mers vus au moins i+1 fois
type FrequencyFilter struct {
	*KmerSetGroup          // Groupe de KmerSet (un par niveau de fréquence)
	MinFreq       int      // v - fréquence minimale requise
}

// NewFrequencyFilter crée un nouveau filtre par fréquence
// minFreq: nombre minimum d'occurrences requises (v)
func NewFrequencyFilter(k, minFreq int) *FrequencyFilter {
	return &FrequencyFilter{
		KmerSetGroup: NewKmerSetGroup(k, minFreq),
		MinFreq:      minFreq,
	}
}

// AddSequence ajoute tous les k-mers d'une séquence au filtre
// Utilise un itérateur pour éviter l'allocation d'un vecteur intermédiaire
func (ff *FrequencyFilter) AddSequence(seq *obiseq.BioSequence) {
	rawSeq := seq.Sequence()
	for canonical := range IterCanonicalKmers(rawSeq, ff.K()) {
		ff.AddKmerCode(canonical)
	}
}

// AddKmerCode ajoute un k-mer encodé au filtre (algorithme principal)
func (ff *FrequencyFilter) AddKmerCode(kmer uint64) {
	// Trouver le niveau actuel du k-mer
	c := 0
	for c < ff.MinFreq && ff.Get(c).Contains(kmer) {
		c++
	}

	// Ajouter au niveau suivant (si pas encore au maximum)
	if c < ff.MinFreq {
		ff.Get(c).AddKmerCode(kmer)
	}
}

// AddCanonicalKmerCode ajoute un k-mer encodé canonique au filtre
func (ff *FrequencyFilter) AddCanonicalKmerCode(kmer uint64) {
	canonical := CanonicalKmer(kmer, ff.K())
	ff.AddKmerCode(canonical)
}

// AddKmer ajoute un k-mer au filtre en encodant la séquence
// La séquence doit avoir exactement k nucléotides
// Zero-allocation: encode directement sans créer de slice intermédiaire
func (ff *FrequencyFilter) AddKmer(seq []byte) {
	kmer := EncodeKmer(seq, ff.K())
	ff.AddKmerCode(kmer)
}

// AddCanonicalKmer ajoute un k-mer canonique au filtre en encodant la séquence
// La séquence doit avoir exactement k nucléotides
// Zero-allocation: encode directement en forme canonique sans créer de slice intermédiaire
func (ff *FrequencyFilter) AddCanonicalKmer(seq []byte) {
	canonical := EncodeCanonicalKmer(seq, ff.K())
	ff.AddKmerCode(canonical)
}

// GetFilteredSet retourne un KmerSet des k-mers avec fréquence ≥ minFreq
func (ff *FrequencyFilter) GetFilteredSet() *KmerSet {
	// Les k-mers filtrés sont dans le dernier niveau
	return ff.Get(ff.MinFreq - 1).Copy()
}

// GetKmersAtLevel retourne un KmerSet des k-mers vus au moins (level+1) fois
// level doit être dans [0, minFreq-1]
func (ff *FrequencyFilter) GetKmersAtLevel(level int) *KmerSet {
	ks := ff.Get(level)
	if ks == nil {
		return NewKmerSet(ff.K())
	}
	return ks.Copy()
}

// Stats retourne des statistiques sur les niveaux de fréquence
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
// (héritée de KmerSetGroup mais redéfinie pour clarté)
func (ff *FrequencyFilter) Clear() {
	ff.KmerSetGroup.Clear()
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
	canonical := CanonicalKmer(kmer, ff.K())
	return ff.Get(ff.MinFreq - 1).Contains(canonical)
}

// GetFrequency retourne la fréquence approximative d'un k-mer
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

// Len retourne le nombre de k-mers filtrés ou à un niveau spécifique
// Sans argument: retourne le nombre de k-mers avec freq ≥ minFreq (dernier niveau)
// Avec argument level: retourne le nombre de k-mers avec freq ≥ (level+1)
// Exemple: Len() pour les k-mers filtrés, Len(2) pour freq ≥ 3
// (héritée de KmerSetGroup mais redéfinie pour la documentation)
func (ff *FrequencyFilter) Len(level ...int) uint64 {
	return ff.KmerSetGroup.Len(level...)
}

// MemoryUsage retourne l'utilisation mémoire en bytes
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
