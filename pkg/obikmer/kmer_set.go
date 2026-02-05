package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/RoaringBitmap/roaring/roaring64"
)

// KmerSet encapsule un ensemble de k-mers stockés dans un Roaring Bitmap
// Fournit des méthodes utilitaires pour manipuler des ensembles de k-mers
type KmerSet struct {
	id       string                 // Identifiant unique du KmerSet
	k        int                    // Taille des k-mers (immutable)
	bitmap   *roaring64.Bitmap      // Bitmap contenant les k-mers
	Metadata map[string]interface{} // Métadonnées utilisateur (clé=valeur atomique)
}

// NewKmerSet crée un nouveau KmerSet vide
func NewKmerSet(k int) *KmerSet {
	return &KmerSet{
		k:        k,
		bitmap:   roaring64.New(),
		Metadata: make(map[string]interface{}),
	}
}

// NewKmerSetFromBitmap crée un KmerSet à partir d'un bitmap existant
func NewKmerSetFromBitmap(k int, bitmap *roaring64.Bitmap) *KmerSet {
	return &KmerSet{
		k:        k,
		bitmap:   bitmap,
		Metadata: make(map[string]interface{}),
	}
}

// K retourne la taille des k-mers (immutable)
func (ks *KmerSet) K() int {
	return ks.k
}

// AddKmerCode ajoute un k-mer encodé à l'ensemble
func (ks *KmerSet) AddKmerCode(kmer uint64) {
	ks.bitmap.Add(kmer)
}

// AddCanonicalKmerCode ajoute un k-mer encodé canonique à l'ensemble
func (ks *KmerSet) AddCanonicalKmerCode(kmer uint64) {
	canonical := CanonicalKmer(kmer, ks.k)
	ks.bitmap.Add(canonical)
}

// AddKmer ajoute un k-mer à l'ensemble en encodant la séquence
// La séquence doit avoir exactement k nucléotides
// Zero-allocation: encode directement sans créer de slice intermédiaire
func (ks *KmerSet) AddKmer(seq []byte) {
	kmer := EncodeKmer(seq, ks.k)
	ks.bitmap.Add(kmer)
}

// AddCanonicalKmer ajoute un k-mer canonique à l'ensemble en encodant la séquence
// La séquence doit avoir exactement k nucléotides
// Zero-allocation: encode directement en forme canonique sans créer de slice intermédiaire
func (ks *KmerSet) AddCanonicalKmer(seq []byte) {
	canonical := EncodeCanonicalKmer(seq, ks.k)
	ks.bitmap.Add(canonical)
}

// AddSequence ajoute tous les k-mers d'une séquence à l'ensemble
// Utilise un itérateur pour éviter l'allocation d'un vecteur intermédiaire
func (ks *KmerSet) AddSequence(seq *obiseq.BioSequence) {
	rawSeq := seq.Sequence()
	for canonical := range IterCanonicalKmers(rawSeq, ks.k) {
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

// Copy crée une copie de l'ensemble (cohérent avec BioSequence.Copy)
func (ks *KmerSet) Copy() *KmerSet {
	// Copier les métadonnées
	metadata := make(map[string]interface{}, len(ks.Metadata))
	for k, v := range ks.Metadata {
		metadata[k] = v
	}

	return &KmerSet{
		id:       ks.id,
		k:        ks.k,
		bitmap:   ks.bitmap.Clone(),
		Metadata: metadata,
	}
}

// Id retourne l'identifiant du KmerSet (cohérent avec BioSequence.Id)
func (ks *KmerSet) Id() string {
	return ks.id
}

// SetId définit l'identifiant du KmerSet (cohérent avec BioSequence.SetId)
func (ks *KmerSet) SetId(id string) {
	ks.id = id
}

// Union retourne l'union de cet ensemble avec un autre
func (ks *KmerSet) Union(other *KmerSet) *KmerSet {
	if ks.k != other.k {
		panic(fmt.Sprintf("Cannot union KmerSets with different k values: %d vs %d", ks.k, other.k))
	}
	result := ks.bitmap.Clone()
	result.Or(other.bitmap)
	return NewKmerSetFromBitmap(ks.k, result)
}

// Intersect retourne l'intersection de cet ensemble avec un autre
func (ks *KmerSet) Intersect(other *KmerSet) *KmerSet {
	if ks.k != other.k {
		panic(fmt.Sprintf("Cannot intersect KmerSets with different k values: %d vs %d", ks.k, other.k))
	}
	result := ks.bitmap.Clone()
	result.And(other.bitmap)
	return NewKmerSetFromBitmap(ks.k, result)
}

// Difference retourne la différence de cet ensemble avec un autre (this - other)
func (ks *KmerSet) Difference(other *KmerSet) *KmerSet {
	if ks.k != other.k {
		panic(fmt.Sprintf("Cannot subtract KmerSets with different k values: %d vs %d", ks.k, other.k))
	}
	result := ks.bitmap.Clone()
	result.AndNot(other.bitmap)
	return NewKmerSetFromBitmap(ks.k, result)
}

// Iterator retourne un itérateur sur tous les k-mers de l'ensemble
func (ks *KmerSet) Iterator() roaring64.IntIterable64 {
	return ks.bitmap.Iterator()
}

// Bitmap retourne le bitmap sous-jacent (pour compatibilité)
func (ks *KmerSet) Bitmap() *roaring64.Bitmap {
	return ks.bitmap
}
