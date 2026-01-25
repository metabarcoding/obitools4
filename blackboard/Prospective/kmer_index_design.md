# Index de k-mers pour génomes de grande taille

## Contexte et objectifs

### Cas d'usage

- Indexation de k-mers longs (k=31) pour des génomes de grande taille (< 10 Go par génome)
- Nombre de génomes : plusieurs dizaines à quelques centaines
- Indexation en parallèle
- Stockage sur disque
- Possibilité d'ajouter des génomes, mais pas de modifier un génome existant

### Requêtes cibles

- **Présence/absence** d'un k-mer dans un génome
- **Intersection** entre génomes
- **Distances** : Jaccard (présence/absence) et potentiellement Bray-Curtis (comptage)

### Ressources disponibles

- 128 Go de RAM
- Stockage disque

---

## Estimation des volumes

### Par génome

- **10 Go de séquence** → ~10¹⁰ k-mers bruts (chevauchants)
- **Après déduplication** : typiquement 10-50% de k-mers uniques → **~1-5 × 10⁹ k-mers distincts**

### Espace théorique

- **k=31** → 62 bits → ~4.6 × 10¹⁸ k-mers possibles
- Table d'indexation directe impossible

---

## Métriques de distance

### Présence/absence (binaire)

- **Jaccard** : |A ∩ B| / |A ∪ B|
- **Sørensen-Dice** : 2|A ∩ B| / (|A| + |B|)

### Comptage (abondance)

- **Bray-Curtis** : 1 - (2 × Σ min(aᵢ, bᵢ)) / (Σ aᵢ + Σ bᵢ)

Note : Pour Bray-Curtis, le stockage des comptages est nécessaire, ce qui augmente significativement la taille de l'index.

---

## Options d'indexation

### Option 1 : Bloom Filter par génome

**Principe** : Structure probabiliste pour test d'appartenance.

**Avantages :**
- Très compact : ~10 bits/élément pour FPR ~1%
- Construction rapide, streaming
- Facile à sérialiser/désérialiser
- Intersection et Jaccard estimables via formules analytiques

**Inconvénients :**
- Faux positifs (pas de faux négatifs)
- Distances approximatives

**Taille estimée** : 1-6 Go par génome (selon FPR cible)

#### Dimensionnement des Bloom filters

```
\mathrm{FPR} ;=; \left(1 - e^{-h n / m}\right)^h
```


| Bits/élément | FPR optimal | k (hash functions) |
|--------------|-------------|---------------------|
| 8            | ~2%         | 5-6                 |
| 10           | ~1%         | 7                   |
| 12           | ~0.3%       | 8                   |
| 16           | ~0.01%      | 11                  |

Formule du taux de faux positifs :
```
FPR ≈ (1 - e^(-kn/m))^k
```
Où n = nombre d'éléments, m = nombre de bits, k = nombre de hash functions.

### Option 2 : Ensemble trié de k-mers

**Principe** : Stocker les k-mers (uint64) triés, avec compression possible.

**Avantages :**
- Exact (pas de faux positifs)
- Intersection/union par merge sort O(n+m)
- Compression efficace (delta encoding sur k-mers triés)

**Inconvénients :**
- Plus volumineux : 8 octets/k-mer
- Construction plus lente (tri nécessaire)

**Taille estimée** : 8-40 Go par génome (non compressé)

### Option 3 : MPHF (Minimal Perfect Hash Function)

**Principe** : Fonction de hash parfaite minimale pour les k-mers présents.

**Avantages :**
- Très compact : ~3-4 bits/élément
- Lookup O(1)
- Exact pour les k-mers présents

**Inconvénients :**
- Construction coûteuse (plusieurs passes)
- Statique (pas d'ajout de k-mers après construction)
- Ne distingue pas "absent" vs "jamais vu" sans structure auxiliaire

### Option 4 : Hybride MPHF + Bloom filter

- MPHF pour mapping compact des k-mers présents
- Bloom filter pour pré-filtrage des absents

---

## Optimisation : Indexation de (k-2)-mers pour requêtes k-mers

### Principe

Au lieu d'indexer directement les 31-mers dans un Bloom filter, on indexe les 29-mers. Pour tester la présence d'un 31-mer, on vérifie que les **trois 29-mers** qu'il contient sont présents :

- positions 0-28
- positions 1-29
- positions 2-30

### Analyse probabiliste

Si le Bloom filter a un FPR de p pour un 29-mer individuel, le FPR effectif pour un 31-mer devient **p³** (les trois requêtes doivent toutes être des faux positifs).

| FPR 29-mer | FPR 31-mer effectif |
|------------|---------------------|
| 10%        | 0.1%                |
| 5%         | 0.0125%             |
| 1%         | 0.0001%             |

### Avantages

1. **Moins d'éléments à stocker** : il y a moins de 29-mers distincts que de 31-mers distincts dans un génome (deux 31-mers différents peuvent partager un même 29-mer)

2. **FPR drastiquement réduit** : FPR³ avec seulement 3 requêtes

3. **Index plus compact** : on peut utiliser moins de bits par élément (FPR plus élevé acceptable sur le 29-mer) tout en obtenant un FPR très bas sur le 31-mer

### Trade-off

Un Bloom filter à **5-6 bits/élément** pour les 29-mers donnerait un FPR effectif < 0.01% pour les 31-mers, soit environ **2× plus compact** que l'approche directe à qualité égale.

**Coût** : 3× plus de requêtes par lookup (mais les requêtes Bloom sont très rapides).

---

## Accélération des calculs de distance : MinHash

### Principe

Pré-calculer une "signature" compacte (sketch) de chaque génome permettant d'estimer rapidement Jaccard sans charger les index complets.

### Avantages

- Matrice de distances entre 100+ génomes en quelques secondes
- Signature de taille fixe (ex: 1000-10000 hash values) quel que soit le génome
- Stockage minimal

### Utilisation

1. Construction : une passe sur les k-mers de chaque génome
2. Distance : comparaison des sketches en O(taille du sketch)

---

## Architecture recommandée

### Pour présence/absence + Jaccard

1. **Index principal** : Bloom filter de (k-2)-mers avec l'optimisation décrite
   - Compact (~3-5 Go par génome)
   - FPR très bas pour les k-mers grâce aux requêtes triples

2. **Sketches MinHash** : pour calcul rapide des distances entre génomes
   - Quelques Ko par génome
   - Permet exploration rapide de la matrice de distances

### Pour comptage + Bray-Curtis

1. **Index principal** : k-mers triés + comptages
   - uint64 (k-mer) + uint8/uint16 (count)
   - Compression delta possible
   - Plus volumineux mais exact

2. **Sketches** : variantes de MinHash pour données pondérées (ex: HyperMinHash)

---

## Prochaines étapes

1. Implémenter un Bloom filter optimisé pour k-mers
2. Implémenter l'optimisation (k-2)-mer → k-mer
3. Implémenter MinHash pour les sketches
4. Définir le format de sérialisation sur disque
5. Benchmarker sur des génomes réels
