package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// KmerSetGroup représente un vecteur de KmerSet
// Utilisé pour gérer plusieurs ensembles de k-mers (par exemple, par niveau de fréquence)
type KmerSetGroup struct {
	id       string                 // Identifiant unique du KmerSetGroup
	k        int                    // Taille des k-mers (immutable)
	sets     []*KmerSet             // Vecteur de KmerSet
	Metadata map[string]interface{} // Métadonnées du groupe (pas des sets individuels)
}

// NewKmerSetGroup crée un nouveau groupe de n KmerSets
func NewKmerSetGroup(k int, n int) *KmerSetGroup {
	if n < 1 {
		panic("KmerSetGroup size must be >= 1")
	}

	sets := make([]*KmerSet, n)
	for i := range sets {
		sets[i] = NewKmerSet(k)
	}

	return &KmerSetGroup{
		k:        k,
		sets:     sets,
		Metadata: make(map[string]interface{}),
	}
}

// K retourne la taille des k-mers (immutable)
func (ksg *KmerSetGroup) K() int {
	return ksg.k
}

// Size retourne le nombre de KmerSet dans le groupe
func (ksg *KmerSetGroup) Size() int {
	return len(ksg.sets)
}

// Get retourne le KmerSet à l'index donné
// Retourne nil si l'index est invalide
func (ksg *KmerSetGroup) Get(index int) *KmerSet {
	if index < 0 || index >= len(ksg.sets) {
		return nil
	}
	return ksg.sets[index]
}

// Set remplace le KmerSet à l'index donné
// Panique si l'index est invalide ou si le k ne correspond pas
func (ksg *KmerSetGroup) Set(index int, ks *KmerSet) {
	if index < 0 || index >= len(ksg.sets) {
		panic(fmt.Sprintf("Index out of bounds: %d (size: %d)", index, len(ksg.sets)))
	}
	if ks.k != ksg.k {
		panic(fmt.Sprintf("KmerSet k mismatch: expected %d, got %d", ksg.k, ks.k))
	}
	ksg.sets[index] = ks
}

// Len retourne le nombre de k-mers dans un KmerSet spécifique
// Sans argument: retourne le nombre de k-mers dans le dernier KmerSet
// Avec argument index: retourne le nombre de k-mers dans le KmerSet à cet index
func (ksg *KmerSetGroup) Len(index ...int) uint64 {
	if len(index) == 0 {
		// Sans argument: dernier KmerSet
		return ksg.sets[len(ksg.sets)-1].Len()
	}

	// Avec argument: KmerSet spécifique
	idx := index[0]
	if idx < 0 || idx >= len(ksg.sets) {
		return 0
	}
	return ksg.sets[idx].Len()
}

// MemoryUsage retourne l'utilisation mémoire totale en bytes
func (ksg *KmerSetGroup) MemoryUsage() uint64 {
	total := uint64(0)
	for _, ks := range ksg.sets {
		total += ks.MemoryUsage()
	}
	return total
}

// Clear vide tous les KmerSet du groupe
func (ksg *KmerSetGroup) Clear() {
	for _, ks := range ksg.sets {
		ks.Clear()
	}
}

// Copy crée une copie complète du groupe (cohérent avec BioSequence.Copy)
func (ksg *KmerSetGroup) Copy() *KmerSetGroup {
	copiedSets := make([]*KmerSet, len(ksg.sets))
	for i, ks := range ksg.sets {
		copiedSets[i] = ks.Copy() // Copy chaque KmerSet avec ses métadonnées
	}

	// Copier les métadonnées du groupe
	groupMetadata := make(map[string]interface{}, len(ksg.Metadata))
	for k, v := range ksg.Metadata {
		groupMetadata[k] = v
	}

	return &KmerSetGroup{
		id:       ksg.id,
		k:        ksg.k,
		sets:     copiedSets,
		Metadata: groupMetadata,
	}
}

// Id retourne l'identifiant du KmerSetGroup (cohérent avec BioSequence.Id)
func (ksg *KmerSetGroup) Id() string {
	return ksg.id
}

// SetId définit l'identifiant du KmerSetGroup (cohérent avec BioSequence.SetId)
func (ksg *KmerSetGroup) SetId(id string) {
	ksg.id = id
}

// AddSequence ajoute tous les k-mers d'une séquence à un KmerSet spécifique
func (ksg *KmerSetGroup) AddSequence(seq *obiseq.BioSequence, index int) {
	if index < 0 || index >= len(ksg.sets) {
		panic(fmt.Sprintf("Index out of bounds: %d (size: %d)", index, len(ksg.sets)))
	}
	ksg.sets[index].AddSequence(seq)
}

// AddSequences ajoute tous les k-mers de plusieurs séquences à un KmerSet spécifique
func (ksg *KmerSetGroup) AddSequences(sequences *obiseq.BioSequenceSlice, index int) {
	if index < 0 || index >= len(ksg.sets) {
		panic(fmt.Sprintf("Index out of bounds: %d (size: %d)", index, len(ksg.sets)))
	}
	ksg.sets[index].AddSequences(sequences)
}

// Union retourne l'union de tous les KmerSet du groupe
// Optimisation: part du plus grand ensemble pour minimiser les opérations
func (ksg *KmerSetGroup) Union() *KmerSet {
	if len(ksg.sets) == 0 {
		return NewKmerSet(ksg.k)
	}

	if len(ksg.sets) == 1 {
		return ksg.sets[0].Copy()
	}

	// Trouver l'index du plus grand ensemble (celui avec le plus de k-mers)
	maxIdx := 0
	maxCard := ksg.sets[0].Len()
	for i := 1; i < len(ksg.sets); i++ {
		card := ksg.sets[i].Len()
		if card > maxCard {
			maxCard = card
			maxIdx = i
		}
	}

	// Copier le plus grand ensemble et faire les unions in-place
	result := ksg.sets[maxIdx].bitmap.Clone()
	for i := 0; i < len(ksg.sets); i++ {
		if i != maxIdx {
			result.Or(ksg.sets[i].bitmap)
		}
	}

	return NewKmerSetFromBitmap(ksg.k, result)
}

// Intersect retourne l'intersection de tous les KmerSet du groupe
// Optimisation: part du plus petit ensemble pour minimiser les opérations
func (ksg *KmerSetGroup) Intersect() *KmerSet {
	if len(ksg.sets) == 0 {
		return NewKmerSet(ksg.k)
	}

	if len(ksg.sets) == 1 {
		return ksg.sets[0].Copy()
	}

	// Trouver l'index du plus petit ensemble (celui avec le moins de k-mers)
	minIdx := 0
	minCard := ksg.sets[0].Len()
	for i := 1; i < len(ksg.sets); i++ {
		card := ksg.sets[i].Len()
		if card < minCard {
			minCard = card
			minIdx = i
		}
	}

	// Copier le plus petit ensemble et faire les intersections in-place
	result := ksg.sets[minIdx].bitmap.Clone()
	for i := 0; i < len(ksg.sets); i++ {
		if i != minIdx {
			result.And(ksg.sets[i].bitmap)
		}
	}

	return NewKmerSetFromBitmap(ksg.k, result)
}

// Stats retourne des statistiques pour chaque KmerSet du groupe
type KmerSetGroupStats struct {
	K          int
	Size       int              // Nombre de KmerSet
	TotalBytes uint64           // Mémoire totale utilisée
	Sets       []KmerSetStats   // Stats de chaque KmerSet
}

type KmerSetStats struct {
	Index     int    // Index du KmerSet dans le groupe
	Len       uint64 // Nombre de k-mers
	SizeBytes uint64 // Taille en bytes
}

func (ksg *KmerSetGroup) Stats() KmerSetGroupStats {
	stats := KmerSetGroupStats{
		K:    ksg.k,
		Size: len(ksg.sets),
		Sets: make([]KmerSetStats, len(ksg.sets)),
	}

	for i, ks := range ksg.sets {
		sizeBytes := ks.MemoryUsage()
		stats.Sets[i] = KmerSetStats{
			Index:     i,
			Len:       ks.Len(),
			SizeBytes: sizeBytes,
		}
		stats.TotalBytes += sizeBytes
	}

	return stats
}

func (ksgs KmerSetGroupStats) String() string {
	result := fmt.Sprintf(`KmerSetGroup Statistics (k=%d, size=%d):
  Total memory: %.2f MB

Set breakdown:
`, ksgs.K, ksgs.Size, float64(ksgs.TotalBytes)/1024/1024)

	for _, set := range ksgs.Sets {
		result += fmt.Sprintf("  Set[%d]: %d k-mers (%.2f MB)\n",
			set.Index,
			set.Len,
			float64(set.SizeBytes)/1024/1024)
	}

	return result
}
