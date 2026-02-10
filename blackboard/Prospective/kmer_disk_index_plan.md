# Plan de refonte du package obikmer : index disk-based par partitions minimizer

## Constat

Les roaring64 bitmaps ne sont pas adaptés au stockage de 10^10 k-mers
(k=31) dispersés sur un espace de 2^62. L'overhead structurel (containers
roaring par high key 32 bits) dépasse la taille des données elles-mêmes,
et les opérations `Or()` entre bitmaps fragmentés ne terminent pas en
temps raisonnable.

## Principe de la nouvelle architecture

Un `KmerSet` est un ensemble trié de k-mers canoniques (uint64) stocké
sur disque, partitionné par minimizer. Chaque partition est un fichier
binaire contenant des uint64 triés, compressés par delta-varint.

Un `KmerSetGroup` est un répertoire contenant N ensembles partitionnés
de la même façon (même k, même m, même P).

Un `KmerSet` est un `KmerSetGroup` de taille 1 (singleton).

Les opérations ensemblistes se font partition par partition, en merge
streaming, sans charger l'index complet en mémoire.

## Cycle de vie d'un index

L'index a deux phases distinctes :

1. **Phase de construction (mutable)** : on ouvre un index, on y ajoute
   des séquences. Pour chaque séquence, les super-kmers sont extraits
   et écrits de manière compacte (2 bits/base) dans le fichier
   temporaire de partition correspondant (`minimizer % P`). Les
   super-kmers sont une représentation compressée naturelle des k-mers
   chevauchants : un super-kmer de longueur L encode L-k+1 k-mers en
   ne stockant que ~L/4 bytes au lieu de (L-k+1) × 8 bytes.

2. **Phase de clôture (optimisation)** : on ferme l'index, ce qui
   déclenche le traitement **partition par partition** (indépendant,
   parallélisable) :
   - Charger les super-kmers de la partition
   - En extraire tous les k-mers canoniques
   - Trier le tableau de k-mers
   - Dédupliquer (et compter si FrequencyFilter)
   - Delta-encoder et écrire le fichier .kdi final
   Après clôture, l'index est statique et immuable.

3. **Phase de lecture (immutable)** : opérations ensemblistes,
   Jaccard, Quorum, Contains, itération. Toutes en streaming.

---

## Format sur disque

### Index finalisé

```
index_dir/
  metadata.toml
  set_0/
    part_0000.kdi
    part_0001.kdi
    ...
    part_{P-1}.kdi
  set_1/
    part_0000.kdi
    ...
  ...
  set_{N-1}/
    ...
```

### Fichiers temporaires pendant la construction

```
index_dir/
  .build/
    set_0/
      part_0000.skm          # super-kmers encodés 2 bits/base
      part_0001.skm
      ...
    set_1/
      ...
```

Le répertoire `.build/` est supprimé après Close().

### metadata.toml

```toml
id = "mon_index"
k = 31
m = 13
partitions = 1024
type = "KmerSetGroup"       # ou "KmerSet" (N=1)
size = 3                    # nombre de sets (N)
sets_ids = ["genome_A", "genome_B", "genome_C"]

[user_metadata]
organism = "Triticum aestivum"

[sets_metadata]
# métadonnées individuelles par set si nécessaire
```

### Fichier .kdi (Kmer Delta Index)

Format binaire :

```
[magic: 4 bytes "KDI\x01"]
[count: uint64 little-endian]       # nombre de k-mers dans cette partition
[first: uint64 little-endian]       # premier k-mer (valeur absolue)
[delta_1: varint]                   # arr[1] - arr[0]
[delta_2: varint]                   # arr[2] - arr[1]
...
[delta_{count-1}: varint]           # arr[count-1] - arr[count-2]
```

Varint : encoding unsigned, 7 bits utiles par byte, bit de poids fort
= continuation (identique au varint protobuf).

Fichier vide (partition sans k-mer) : magic + count=0.

### Fichier .skm (Super-Kmer temporaire)

Format binaire, séquence de super-kmers encodés :

```
[len: uint16 little-endian]         # longueur du super-kmer en bases
[sequence: ceil(len/4) bytes]       # séquence encodée 2 bits/base, packed
...
```

**Compression par rapport au stockage de k-mers bruts** :

Un super-kmer de longueur L contient L-k+1 k-mers.
- Stockage super-kmer : 2 + ceil(L/4) bytes
- Stockage k-mers bruts : (L-k+1) × 8 bytes

Exemple avec k=31, super-kmer typique L=50 :
- Super-kmer : 2 + 13 = 15 bytes → encode 20 k-mers
- K-mers bruts : 20 × 8 = 160 bytes
- **Facteur de compression : ~10×**

Pour un génome de 10 Gbases (~10^10 k-mers bruts) :
- K-mers bruts : ~80 Go par set temporaire
- Super-kmers : **~8 Go** par set temporaire

Avec FrequencyFilter et couverture 30× :
- K-mers bruts : ~2.4 To
- Super-kmers : **~240 Go**

---

## FrequencyFilter

Le FrequencyFilter n'est plus un type de données séparé. C'est un
**mode de construction** du builder. Le résultat est un KmerSetGroup
standard.

### Principe

Pendant la construction, tous les super-kmers sont écrits dans les
fichiers temporaires .skm, y compris les doublons (chaque occurrence
de chaque séquence est écrite).

Pendant Close(), pour chaque partition :
1. Charger tous les super-kmers de la partition
2. Extraire tous les k-mers canoniques dans un tableau []uint64
3. Trier le tableau
4. Parcourir linéairement : les k-mers identiques sont consécutifs
5. Compter les occurrences de chaque k-mer
6. Si count >= minFreq → écrire dans le .kdi final (une seule fois)
7. Sinon → ignorer

### Dimensionnement

Pour un génome de 10 Gbases avec couverture 30× :
- N_brut ≈ 3×10^11 k-mers bruts
- Espace temporaire .skm ≈ 240 Go (compressé super-kmer)
- RAM par partition pendant Close() :
  Avec P=1024 : ~3×10^8 k-mers/partition × 8 = **~2.4 Go**
  Avec P=4096 : ~7.3×10^7 k-mers/partition × 8 = **~600 Mo**

Le choix de P détermine le compromis nombre de fichiers vs RAM par
partition.

### Sans FrequencyFilter (déduplication simple)

Pour de la déduplication simple (chaque k-mer écrit une fois), le
builder peut dédupliquer au niveau des buffers en RAM avant flush.
Cela réduit significativement l'espace temporaire car les doublons
au sein d'un même buffer (provenant de séquences proches) sont
éliminés immédiatement.

---

## API publique visée

### Structures

```go
// KmerSetGroup est l'entité de base.
// Un KmerSet est un KmerSetGroup avec Size() == 1.
type KmerSetGroup struct {
    // champs internes : path, k, m, P, N, metadata, état
}

// KmerSetGroupBuilder construit un KmerSetGroup mutable.
type KmerSetGroupBuilder struct {
    // champs internes : buffers I/O par partition et par set,
    // fichiers temporaires .skm, paramètres (minFreq, etc.)
}
```

### Construction

```go
// NewKmerSetGroupBuilder crée un builder pour un nouveau KmerSetGroup.
//   directory : répertoire de destination
//   k : taille des k-mers (1-31)
//   m : taille des minimizers (-1 pour auto = ceil(k/2.5))
//   n : nombre de sets dans le groupe
//   P : nombre de partitions (-1 pour auto)
//   options : options de construction (FrequencyFilter, etc.)
func NewKmerSetGroupBuilder(directory string, k, m, n, P int,
    options ...BuilderOption) (*KmerSetGroupBuilder, error)

// WithMinFrequency active le mode FrequencyFilter.
// Seuls les k-mers vus >= minFreq fois sont conservés dans l'index
// final. Les super-kmers sont écrits avec leurs doublons pendant
// la construction ; le comptage exact se fait au Close().
func WithMinFrequency(minFreq int) BuilderOption

// AddSequence extrait les super-kmers d'une séquence et les écrit
// dans les fichiers temporaires de partition du set i.
func (b *KmerSetGroupBuilder) AddSequence(setIndex int, seq *obiseq.BioSequence)

// AddSuperKmer écrit un super-kmer dans le fichier temporaire de
// sa partition pour le set i.
func (b *KmerSetGroupBuilder) AddSuperKmer(setIndex int, sk SuperKmer)

// Close finalise la construction :
//   - flush des buffers d'écriture
//   - pour chaque partition de chaque set (parallélisable) :
//     - charger les super-kmers depuis le .skm
//     - extraire les k-mers canoniques
//     - trier, dédupliquer (compter si freq filter)
//     - delta-encoder et écrire le .kdi
//   - écrire metadata.toml
//   - supprimer le répertoire .build/
// Retourne le KmerSetGroup en lecture seule.
func (b *KmerSetGroupBuilder) Close() (*KmerSetGroup, error)
```

### Lecture et opérations

```go
// OpenKmerSetGroup ouvre un index finalisé en lecture seule.
func OpenKmerSetGroup(directory string) (*KmerSetGroup, error)

// --- Métadonnées (API inchangée) ---
func (ksg *KmerSetGroup) K() int
func (ksg *KmerSetGroup) M() int          // nouveau : taille du minimizer
func (ksg *KmerSetGroup) Partitions() int  // nouveau : nombre de partitions
func (ksg *KmerSetGroup) Size() int
func (ksg *KmerSetGroup) Id() string
func (ksg *KmerSetGroup) SetId(id string)
func (ksg *KmerSetGroup) HasAttribute(key string) bool
func (ksg *KmerSetGroup) GetAttribute(key string) (interface{}, bool)
func (ksg *KmerSetGroup) SetAttribute(key string, value interface{})
// ... etc (toute l'API attributs actuelle est conservée)

// --- Opérations ensemblistes ---
// Toutes produisent un nouveau KmerSetGroup singleton sur disque.
// Opèrent partition par partition en streaming.

func (ksg *KmerSetGroup) Union(outputDir string) (*KmerSetGroup, error)
func (ksg *KmerSetGroup) Intersect(outputDir string) (*KmerSetGroup, error)
func (ksg *KmerSetGroup) Difference(outputDir string) (*KmerSetGroup, error)
func (ksg *KmerSetGroup) QuorumAtLeast(q int, outputDir string) (*KmerSetGroup, error)
func (ksg *KmerSetGroup) QuorumExactly(q int, outputDir string) (*KmerSetGroup, error)
func (ksg *KmerSetGroup) QuorumAtMost(q int, outputDir string) (*KmerSetGroup, error)

// --- Opérations entre deux KmerSetGroups ---
// Les deux groupes doivent avoir les mêmes k, m, P.

func (ksg *KmerSetGroup) UnionWith(other *KmerSetGroup, outputDir string) (*KmerSetGroup, error)
func (ksg *KmerSetGroup) IntersectWith(other *KmerSetGroup, outputDir string) (*KmerSetGroup, error)

// --- Métriques (résultat en mémoire, pas de sortie disque) ---

func (ksg *KmerSetGroup) JaccardDistanceMatrix() *obidist.DistMatrix
func (ksg *KmerSetGroup) JaccardSimilarityMatrix() *obidist.DistMatrix

// --- Accès individuel ---

func (ksg *KmerSetGroup) Len(setIndex ...int) uint64
func (ksg *KmerSetGroup) Contains(setIndex int, kmer uint64) bool
func (ksg *KmerSetGroup) Iterator(setIndex int) iter.Seq[uint64]
```

---

## Implémentation interne

### Primitives bas niveau

**`varint.go`** : encode/decode varint uint64

```go
func EncodeVarint(w io.Writer, v uint64) (int, error)
func DecodeVarint(r io.Reader) (uint64, error)
```

### Format .kdi

**`kdi_writer.go`** : écriture d'un fichier .kdi à partir d'un flux
trié de uint64 (delta-encode au vol).

```go
type KdiWriter struct { ... }
func NewKdiWriter(path string) (*KdiWriter, error)
func (w *KdiWriter) Write(kmer uint64) error
func (w *KdiWriter) Close() error
```

**`kdi_reader.go`** : lecture streaming d'un fichier .kdi (décode
les deltas au vol).

```go
type KdiReader struct { ... }
func NewKdiReader(path string) (*KdiReader, error)
func (r *KdiReader) Next() (uint64, bool)
func (r *KdiReader) Count() uint64
func (r *KdiReader) Close() error
```

### Format .skm

**`skm_writer.go`** : écriture de super-kmers encodés 2 bits/base.

```go
type SkmWriter struct { ... }
func NewSkmWriter(path string) (*SkmWriter, error)
func (w *SkmWriter) Write(sk SuperKmer) error
func (w *SkmWriter) Close() error
```

**`skm_reader.go`** : lecture de super-kmers depuis un fichier .skm.

```go
type SkmReader struct { ... }
func NewSkmReader(path string) (*SkmReader, error)
func (r *SkmReader) Next() (SuperKmer, bool)
func (r *SkmReader) Close() error
```

### Merge streaming

**`kdi_merge.go`** : k-way merge de plusieurs flux triés.

```go
type KWayMerge struct { ... }
func NewKWayMerge(readers []*KdiReader) *KWayMerge
func (m *KWayMerge) Next() (kmer uint64, count int, ok bool)
func (m *KWayMerge) Close() error
```

### Builder

**`kmer_set_builder.go`** : construction d'un KmerSetGroup.

Le builder gère :
- P × N écrivains .skm bufferisés (un par partition × set)
- À la clôture : traitement partition par partition
  (parallélisable sur plusieurs cores)

Gestion mémoire des buffers d'écriture :
- Chaque SkmWriter a un buffer I/O de taille raisonnable (~64 Ko)
- Avec P=1024 et N=1 : 1024 × 64 Ko = 64 Mo de buffers
- Avec P=1024 et N=10 : 640 Mo de buffers
- Pas de buffer de k-mers en RAM : tout est écrit sur disque
  immédiatement via les super-kmers

RAM pendant Close() (tri d'une partition) :
- Charger les super-kmers → extraire les k-mers → tableau []uint64
- Avec P=1024 et 10^10 k-mers/set : ~10^7 k-mers/partition × 8 = ~80 Mo
- Avec FrequencyFilter (doublons) et couverture 30× :
  ~3×10^8/partition × 8 = ~2.4 Go (ajustable via P)

### Structure disk-based

**`kmer_set_disk.go`** : KmerSetGroup en lecture seule.

**`kmer_set_disk_ops.go`** : opérations ensemblistes par merge
streaming partition par partition.

---

## Ce qui change par rapport à l'API actuelle

### Changements de sémantique

| Aspect | Ancien (roaring) | Nouveau (disk-based) |
|---|---|---|
| Stockage | En mémoire (roaring64.Bitmap) | Sur disque (.kdi delta-encoded) |
| Temporaire construction | En mémoire | Super-kmers sur disque (.skm 2 bits/base) |
| Mutabilité | Mutable à tout moment | Builder → Close() → immutable |
| Opérations ensemblistes | Résultat en mémoire | Résultat sur disque (nouveau répertoire) |
| Contains | O(1) roaring lookup | O(log n) recherche binaire sur .kdi |
| Itération | Roaring iterator | Streaming décodage delta-varint |

### API conservée (signatures identiques ou quasi-identiques)

- `KmerSetGroup` : `K()`, `Size()`, `Id()`, `SetId()`
- Toute l'API attributs
- `JaccardDistanceMatrix()`, `JaccardSimilarityMatrix()`
- `Len()`, `Contains()`

### API modifiée

- `Union()`, `Intersect()`, etc. : ajout du paramètre `outputDir`
- `QuorumAtLeast()`, etc. : idem
- Construction : `NewKmerSetGroupBuilder()` + `AddSequence()` + `Close()`
  au lieu de manipulation directe

### API supprimée

- `KmerSet` comme type distinct (remplacé par KmerSetGroup singleton)
- `FrequencyFilter` comme type distinct (mode du Builder)
- Tout accès direct à `roaring64.Bitmap`
- `KmerSet.Copy()` (copie de répertoire à la place)
- `KmerSet.Union()`, `.Intersect()`, `.Difference()` (deviennent méthodes
  de KmerSetGroup avec outputDir)

---

## Fichiers à créer / modifier dans pkg/obikmer

### Nouveaux fichiers

| Fichier | Contenu |
|---|---|
| `varint.go` | Encode/Decode varint uint64 |
| `kdi_writer.go` | Écrivain de fichiers .kdi (delta-encoded) |
| `kdi_reader.go` | Lecteur streaming de fichiers .kdi |
| `skm_writer.go` | Écrivain de super-kmers encodés 2 bits/base |
| `skm_reader.go` | Lecteur de super-kmers depuis .skm |
| `kdi_merge.go` | K-way merge streaming de flux triés |
| `kmer_set_builder.go` | KmerSetGroupBuilder (construction) |
| `kmer_set_disk.go` | KmerSetGroup disk-based (lecture, métadonnées) |
| `kmer_set_disk_ops.go` | Opérations ensemblistes streaming |

### Fichiers à supprimer

| Fichier | Raison |
|---|---|
| `kmer_set.go` | Remplacé par kmer_set_disk.go |
| `kmer_set_group.go` | Idem |
| `kmer_set_attributes.go` | Intégré dans kmer_set_disk.go |
| `kmer_set_persistence.go` | L'index est nativement sur disque |
| `kmer_set_group_quorum.go` | Intégré dans kmer_set_disk_ops.go |
| `frequency_filter.go` | Mode du Builder, plus de type séparé |
| `kmer_index_builder.go` | Remplacé par kmer_set_builder.go |

### Fichiers conservés tels quels

| Fichier | Contenu |
|---|---|
| `encodekmer.go` | Encodage/décodage k-mers |
| `superkmer.go` | Structure SuperKmer |
| `superkmer_iter.go` | IterSuperKmers, IterCanonicalKmers |
| `encodefourmer.go` | Encode4mer |
| `counting.go` | Count4Mer |
| `kmermap.go` | KmerMap (usage indépendant) |
| `debruijn.go` | Graphe de de Bruijn |

---

## Ordre d'implémentation

1. `varint.go` + tests
2. `skm_writer.go` + `skm_reader.go` + tests
3. `kdi_writer.go` + `kdi_reader.go` + tests
4. `kdi_merge.go` + tests
5. `kmer_set_builder.go` + tests (construction + Close)
6. `kmer_set_disk.go` (structure, métadonnées, Open)
7. `kmer_set_disk_ops.go` + tests (Union, Intersect, Quorum, Jaccard)
8. Adaptation de `pkg/obitools/obikindex/`
9. Suppression des anciens fichiers roaring
10. Adaptation des tests existants

Chaque étape est testable indépendamment.

---

## Dépendances externes

### Supprimées

- `github.com/RoaringBitmap/roaring` : plus nécessaire pour les
  index k-mers (vérifier si d'autres packages l'utilisent encore)

### Ajoutées

- Aucune. Varint, delta-encoding, merge, encodage 2 bits/base :
  tout est implémentable en Go standard.
