# Filtre de Fréquence avec v Niveaux de Roaring Bitmaps

## Algorithme

```go
Pour chaque k-mer rencontré dans les données:
    c = 0
    tant que (k-mer ∈ index[c] ET c < v):
        c++
    
    si c < v:
        index[c].insert(k-mer)
```

**Résultat** : `index[v-1]` contient les k-mers vus **≥ v fois**

---

## Exemple d'exécution (v=3)

```
Données:
  Read1: kmer X
  Read2: kmer X
  Read3: kmer X  (X vu 3 fois)
  Read4: kmer Y
  Read5: kmer Y  (Y vu 2 fois)
  Read6: kmer Z  (Z vu 1 fois)

Exécution:

Read1 (X):
  c=0: X ∉ index[0] → index[0].add(X)
  État: index[0]={X}, index[1]={}, index[2]={}

Read2 (X):
  c=0: X ∈ index[0] → c=1
  c=1: X ∉ index[1] → index[1].add(X)
  État: index[0]={X}, index[1]={X}, index[2]={}

Read3 (X):
  c=0: X ∈ index[0] → c=1
  c=1: X ∈ index[1] → c=2
  c=2: X ∉ index[2] → index[2].add(X)
  État: index[0]={X}, index[1]={X}, index[2]={X}

Read4 (Y):
  c=0: Y ∉ index[0] → index[0].add(Y)
  État: index[0]={X,Y}, index[1]={X}, index[2]={X}

Read5 (Y):
  c=0: Y ∈ index[0] → c=1
  c=1: Y ∉ index[1] → index[1].add(Y)
  État: index[0]={X,Y}, index[1]={X,Y}, index[2]={X}

Read6 (Z):
  c=0: Z ∉ index[0] → index[0].add(Z)
  État: index[0]={X,Y,Z}, index[1]={X,Y}, index[2]={X}

Résultat final:
  index[0] (freq≥1): {X, Y, Z}
  index[1] (freq≥2): {X, Y}
  index[2] (freq≥3): {X}  ← K-mers filtrés ✓
```

---

## Utilisation

```go
// Créer le filtre
filter := obikmer.NewFrequencyFilter(31, 3) // k=31, minFreq=3

// Ajouter les séquences
for _, read := range reads {
    filter.AddSequence(read)
}

// Récupérer les k-mers filtrés (freq ≥ 3)
filtered := filter.GetFilteredSet("filtered")
fmt.Printf("K-mers de qualité: %d\n", filtered.Cardinality())

// Statistiques
stats := filter.Stats()
fmt.Println(stats.String())
```

---

## Performance

### Complexité

**Par k-mer** :
- Lookups : Moyenne ~v/2, pire cas v
- Insertions : 1 Add
- **Pas de Remove** ✅

**Total pour n k-mers** :
- Temps : O(n × v/2)
- Mémoire : O(unique_kmers × v × 2 bytes)

### Early exit pour distribution skewed

Avec distribution typique (séquençage) :
```
80% singletons → 1 lookup (early exit)
15% freq 2-3   → 2-3 lookups
5% freq ≥4     → jusqu'à v lookups

Moyenne réelle : ~2 lookups/kmer (au lieu de v/2)
```

---

## Mémoire

### Pour 10^8 k-mers uniques

| v (minFreq) | Nombre bitmaps | Mémoire | vs map simple |
|-------------|----------------|---------|---------------|
| v=2 | 2 | ~400 MB | 6x moins |
| v=3 | 3 | ~600 MB | 4x moins |
| v=5 | 5 | ~1 GB | 2.4x moins |
| v=10 | 10 | ~2 GB | 1.2x moins |
| v=20 | 20 | ~4 GB | ~égal |

**Note** : Avec distribution skewed (beaucoup de singletons), la mémoire réelle est bien plus faible car les niveaux hauts ont peu d'éléments.

### Exemple réaliste (séquençage)

Pour 10^8 k-mers totaux, v=3 :
```
Distribution:
  80% singletons  → 80M dans index[0]
  15% freq 2-3    → 15M dans index[1]
  5% freq ≥3      → 5M dans index[2]

Mémoire:
  index[0]: 80M × 2 bytes = 160 MB
  index[1]: 15M × 2 bytes = 30 MB
  index[2]: 5M × 2 bytes = 10 MB
  Total: ~200 MB ✅

vs map simple: 80M × 24 bytes = ~2 GB
Réduction: 10x
```

---

## Comparaison des approches

| Approche | Mémoire (10^8 kmers) | Passes | Lookups/kmer | Quand utiliser |
|----------|----------------------|--------|--------------|----------------|
| **v-Bitmaps** | **200-600 MB** | **1** | **~2 (avg)** | **Standard** ✅ |
| Map simple | 2.4 GB | 1 | 1 | Si RAM illimitée |
| Multi-pass | 400 MB | v | v | Si I/O pas cher |

---

## Avantages de v-Bitmaps

✅ **Une seule passe** sur les données  
✅ **Mémoire optimale** avec Roaring bitmaps  
✅ **Pas de Remove** (seulement Contains + Add)  
✅ **Early exit** efficace sur singletons  
✅ **Scalable** jusqu'à v~10-20  
✅ **Simple** à implémenter et comprendre  

---

## Cas d'usage typiques

### 1. Éliminer erreurs de séquençage

```go
filter := obikmer.NewFrequencyFilter(31, 3)

// Traiter FASTQ
for read := range StreamFastq("sample.fastq") {
    filter.AddSequence(read)
}

// K-mers de qualité (pas d'erreurs)
cleaned := filter.GetFilteredSet("cleaned")
```

**Résultat** : Élimine 70-80% des k-mers (erreurs)

### 2. Assemblage de génome

```go
filter := obikmer.NewFrequencyFilter(31, 2)

// Filtrer avant l'assemblage
for read := range reads {
    filter.AddSequence(read)
}

solidKmers := filter.GetFilteredSet("solid")
// Utiliser solidKmers pour le graphe de Bruijn
```

### 3. Comparaison de génomes

```go
collection := obikmer.NewKmerSetCollection(31)

for _, genome := range genomes {
    filter := obikmer.NewFrequencyFilter(31, 3)
    filter.AddSequences(genome.Reads)
    
    cleaned := filter.GetFilteredSet(genome.ID)
    collection.Add(cleaned)
}

// Analyses comparatives sur k-mers de qualité
matrix := collection.ParallelPairwiseJaccard(8)
```

---

## Limites

**Pour v > 20** :
- Trop de lookups (v lookups/kmer)
- Mémoire importante (v × 200MB pour 10^8 kmers)

**Solutions alternatives pour v > 20** :
- Utiliser map simple (9 bytes/kmer) si RAM disponible
- Algorithme différent (sketch, probabiliste)

---

## Optimisations possibles

### 1. Parallélisation

```go
// Traiter plusieurs fichiers en parallèle
filters := make([]*FrequencyFilter, numFiles)

var wg sync.WaitGroup
for i, file := range files {
    wg.Add(1)
    go func(idx int, f string) {
        defer wg.Done()
        filters[idx] = ProcessFile(f, k, minFreq)
    }(i, file)
}
wg.Wait()

// Merger les résultats
merged := MergeFilters(filters)
```

### 2. Streaming avec seuil adaptatif

```go
// Commencer avec v=5, réduire progressivement
filter := obikmer.NewFrequencyFilter(31, 5)

// ... traitement ...

// Si trop de mémoire, réduire à v=3
if filter.MemoryUsage() > threshold {
    filter = ConvertToLowerThreshold(filter, 3)
}
```

---

## Récapitulatif final

**Pour filtrer les k-mers par fréquence ≥ v :**

1. **Créer** : `filter := NewFrequencyFilter(k, v)`
2. **Traiter** : `filter.AddSequence(read)` pour chaque read
3. **Résultat** : `filtered := filter.GetFilteredSet(id)`

**Mémoire** : ~2v MB par million de k-mers uniques  
**Temps** : Une seule passe, ~2 lookups/kmer en moyenne  
**Optimal pour** : v ≤ 20, distribution skewed (séquençage)  

---

## Code fourni

1. **frequency_filter.go** - Implémentation complète
2. **examples_frequency_filter_final.go** - Exemples d'utilisation

**Tout est prêt à utiliser !** 🚀
