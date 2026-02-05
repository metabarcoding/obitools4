package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/RoaringBitmap/roaring/roaring64"
)

// KmerSet encapsule un ensemble de k-mers stockés dans un Roaring Bitmap
// Fournit des méthodes utilitaires pour manipuler des ensembles de k-mers
type KmerSet struct {
	K        int                    // Taille des k-mers
	bitmap   *roaring64.Bitmap      // Bitmap contenant les k-mers
	Metadata map[string]interface{} // Métadonnées utilisateur (clé=valeur atomique)
}

// NewKmerSet crée un nouveau KmerSet vide
func NewKmerSet(k int) *KmerSet {
	return &KmerSet{
		K:        k,
		bitmap:   roaring64.New(),
		Metadata: make(map[string]interface{}),
	}
}

// NewKmerSetFromBitmap crée un KmerSet à partir d'un bitmap existant
func NewKmerSetFromBitmap(k int, bitmap *roaring64.Bitmap) *KmerSet {
	return &KmerSet{
		K:        k,
		bitmap:   bitmap,
		Metadata: make(map[string]interface{}),
	}
}

// Add ajoute un k-mer à l'ensemble
func (ks *KmerSet) Add(kmer uint64) {
	ks.bitmap.Add(kmer)
}

// AddSequence ajoute tous les k-mers d'une séquence à l'ensemble
// Utilise un itérateur pour éviter l'allocation d'un vecteur intermédiaire
func (ks *KmerSet) AddSequence(seq *obiseq.BioSequence) {
	rawSeq := seq.Sequence()
	for canonical := range IterNormalizedKmers(rawSeq, ks.K) {
		ks.bitmap.Add(canonical)
	}
}

// AddSequences ajoute tous les k-mers de plusieurs séquences en batch
func (ks *KmerSet) AddSequences(sequences *obiseq.BioSequenceSlice) {
	for _, seq := range *sequences {
		ks.AddSequence(seq)
	}
}

// Contains vérifie si un k-mer est dans l'ensemble
func (ks *KmerSet) Contains(kmer uint64) bool {
	return ks.bitmap.Contains(kmer)
}

// Len retourne le nombre de k-mers dans l'ensemble
func (ks *KmerSet) Len() uint64 {
	return ks.bitmap.GetCardinality()
}

// MemoryUsage retourne l'utilisation mémoire en bytes
func (ks *KmerSet) MemoryUsage() uint64 {
	return ks.bitmap.GetSizeInBytes()
}

// Clear vide l'ensemble
func (ks *KmerSet) Clear() {
	ks.bitmap.Clear()
}

// Clone crée une copie de l'ensemble
func (ks *KmerSet) Clone() *KmerSet {
	// Copier les métadonnées
	metadata := make(map[string]interface{}, len(ks.Metadata))
	for k, v := range ks.Metadata {
		metadata[k] = v
	}

	return &KmerSet{
		K:        ks.K,
		bitmap:   ks.bitmap.Clone(),
		Metadata: metadata,
	}
}

// Union retourne l'union de cet ensemble avec un autre
func (ks *KmerSet) Union(other *KmerSet) *KmerSet {
	if ks.K != other.K {
		panic(fmt.Sprintf("Cannot union KmerSets with different k values: %d vs %d", ks.K, other.K))
	}
	result := ks.bitmap.Clone()
	result.Or(other.bitmap)
	return NewKmerSetFromBitmap(ks.K, result)
}

// Intersect retourne l'intersection de cet ensemble avec un autre
func (ks *KmerSet) Intersect(other *KmerSet) *KmerSet {
	if ks.K != other.K {
		panic(fmt.Sprintf("Cannot intersect KmerSets with different k values: %d vs %d", ks.K, other.K))
	}
	result := ks.bitmap.Clone()
	result.And(other.bitmap)
	return NewKmerSetFromBitmap(ks.K, result)
}

// Difference retourne la différence de cet ensemble avec un autre (this - other)
func (ks *KmerSet) Difference(other *KmerSet) *KmerSet {
	if ks.K != other.K {
		panic(fmt.Sprintf("Cannot subtract KmerSets with different k values: %d vs %d", ks.K, other.K))
	}
	result := ks.bitmap.Clone()
	result.AndNot(other.bitmap)
	return NewKmerSetFromBitmap(ks.K, result)
}

// Iterator retourne un itérateur sur tous les k-mers de l'ensemble
func (ks *KmerSet) Iterator() roaring64.IntIterable64 {
	return ks.bitmap.Iterator()
}

// Bitmap retourne le bitmap sous-jacent (pour compatibilité)
func (ks *KmerSet) Bitmap() *roaring64.Bitmap {
	return ks.bitmap
}
