# Prospective : Index k-mer v3 — Super-kmers canoniques, unitigs, et Aho-Corasick

## 1. Constat sur l'index v1

L'index actuel (`.kdi` delta-varint) stocke 18.6 milliards de k-mers (k=31, m=13, P=4096, 2 sets) en 85 Go, soit 4.8-5.6 bytes/k-mer. Les causes :

- Le canonical standard `min(fwd, rc)` disperse les k-mers sur 62 bits → deltas ~2^40 → 5-6 bytes varint
- Les k-mers partagés entre sets sont stockés N fois (une fois par set)
- Le matching nécessite N×P ouvertures de fichier (N passes)

## 2. Observations expérimentales

### 2.1 Déréplication brute

Sur un génome de *Betula exilis* 15× couvert, le pipeline `obik lowmask | obik super | obiuniq` réduit **80 Go de fastq.gz en 5.6 Go de fasta.gz** — un facteur 14×. Cela montre que la déréplication au niveau super-kmer est extrêmement efficace et que les super-kmers forment une représentation naturellement compacte.

### 2.2 Après filtre de fréquence (count > 1)

En éliminant les super-kmers observés une seule fois (erreurs de séquençage), le fichier passe de 5.6 Go à **2.7 Go de fasta.gz**. Les statistiques détaillées (obicount) :

| Métrique | Valeur |
|----------|--------|
| Variants (super-kmers uniques) | 37,294,271 |
| Reads (somme des counts) | 148,828,167 |
| Symboles (bases totales variants) | 1,415,018,593 |
| Longueur moyenne super-kmer | **37.9 bases** |
| K-mers/super-kmer moyen (k=31) | **7.9** |
| K-mers totaux estimés | **~295M** |
| Count moyen par super-kmer | **4.0×** |

### 2.3 Comparaison avec l'index v1

| Format | Taille | K-mers | Bytes/k-mer |
|--------|--------|--------|-------------|
| Index .kdi v1 (set Human dans Contaminent_idx) | 12.8 Go | ~3B | 4.3 |
| Delta-varint hypothétique (295M k-mers) | ~1.5 Go | 295M | 5.0 |
| Super-kmers 2-bit packed (*Betula* count>1) | ~354 Mo | 295M | **1.2** |
| Super-kmers fasta.gz (*Betula* count>1) | 2.7 Go | 295M | 9.2* |

\* Le fasta.gz inclut les headers, les counts, et la compression gzip — pas directement comparable au format binaire.

**Le format super-kmer 2-bit est ~4× plus compact que le delta-varint** à nombre égal de k-mers. Cette efficacité vient du fait qu'un super-kmer de 38 bases encode 8 k-mers en ~10 bytes au lieu de 8 × 5 = 40 bytes en delta-varint.

Note : la comparaison n'est pas directe (Contaminent_idx = génomes assemblés, *Betula* = reads bruts filtrés), mais le ratio bytes/k-mer est comparable car il dépend de la longueur des super-kmers, pas de la source des données.

## 3. Stratégie proposée : pipeline de construction v3

### 3.1 Définition du k-mer minimizer-canonique

On redéfinit la forme canonique d'un k-mer en fonction de son minimiseur :

```
CanonicalKmer(kmer, k, m) :
  minimizer = plus petit m-mer canonique dans le k-mer
  si minimizer == forward_mmer(minimizer_pos)
    → garder le k-mer tel quel
  sinon
    → prendre le reverse-complement du k-mer
```

Propriétés :
- **m impair** → aucun m-mer ne peut être palindromique (`m_mer != RC(m_mer)` toujours) → la canonisation par le minimiseur est toujours non-ambiguë. C'est m, pas k, qui doit être impair : l'ambiguïté viendrait d'un minimiseur palindrome (`min == RC(min)`), auquel cas on ne saurait pas dans quel sens orienter le k-mer/super-kmer.
- Tous les k-mers d'un super-kmer partagent le même minimiseur
- **La canonisation peut se faire au niveau du super-kmer entier** : si `minimizer != canonical(minimizer)`, on RC le super-kmer complet. Tous les k-mers qu'il contient deviennent automatiquement minimizer-canoniques.

### 3.2 Pipeline de construction

```
Séquences brutes ([]byte, 1 byte/base)
    │
    ▼
[0] Encodage 2-bit + nettoyage
    │  - Encoder chaque séquence en 2 bits/base ([]byte packed)
    │  - Couper aux bases ambiguës (N, R, Y, W, S, K, M, B, D, H, V)
    │  - Retirer les fragments de longueur < k
    │  - Résultat : fragments 2-bit clean, prêts pour toutes les opérations
    ▼
[1] Filtre de complexité (lowmask sur vecteurs 2-bit)
    │  Supprime/masque les régions de faible entropie
    ▼
[2] Extraction des super-kmers (sur vecteurs 2-bit, non canonisé)
    │  Chaque super-kmer a un minimiseur et une séquence 2-bit packed
    ▼
[3] Canonisation au niveau super-kmer
    │  Si minimizer != CanonicalKmer(minimizer) → RC le super-kmer (op bit)
    │  Résultat : super-kmers canoniques 2-bit packed
    ▼
[4] Écriture dans les partitions .skm (partition = minimizer % P)
    │  Format natif 2-bit → écriture directe, pas de conversion
    ▼
[5] Déréplication des super-kmers par partition
    │  Trier les super-kmers (comparaison uint64 sur données packed → très rapide)
    │  Compter les occurrences identiques
    │  Résultat : super-kmers uniques avec count
    ▼
[6] Construction des unitigs canoniques par partition
    │  Assembler les super-kmers qui se chevauchent de (k-1) bases
    │  en chaînes linéaires non-branchantes (tout en 2-bit)
    │  Propager les counts : vecteur de poids par unitig
    ▼
[7] Filtre de fréquence sur le graphe pondéré (voir section 4)
    │  Supprimer les k-mers (positions) avec poids < seuil
    │  Re-calculer les unitigs après filtrage
    ▼
[8] Stockage des unitigs avec bitmask multiset
    │  Format compact sur disque (déjà en 2-bit, écriture directe)
    ▼
Index v3
```

### 3.2bis Pourquoi encoder en 2-bit dès le début ?

**Alternative rejetée** : travailler en `[]byte` (1 byte/base) puis encoder en 2-bit seulement pour le stockage final.

| Aspect | `[]byte` (1 byte/base) | 2-bit packed |
|--------|----------------------|--------------|
| Programmation | Simple (slicing natif, pas de bit-shift) | Plus complexe (masques, shifts) |
| Mémoire par super-kmer (38 bases) | 38 bytes | 10 bytes (**3.8×** moins) |
| 37M super-kmers en RAM | ~1.4 Go | ~370 Mo |
| Tri (comparaison) | `bytes.Compare` sur slices | Comparaison uint64 (**beaucoup** plus rapide) |
| Format .skm | Conversion encode/decode à chaque I/O | Écriture/lecture directe |
| RC d'un super-kmer | Boucle sur bytes + lookup | Opérations bit (une instruction pour complement) |

L'opération la plus coûteuse du pipeline est le **tri des super-kmers** pour la déréplication (étape 5). En 2-bit packed, un super-kmer de ≤32 bases tient dans un `uint64` → tri par comparaison entière (une instruction CPU). Un super-kmer de 33-64 bases tient dans deux `uint64` → tri en deux comparaisons.

Le code de manipulation 2-bit est plus complexe à écrire mais **s'écrit une seule fois** (bibliothèque de primitives) et bénéficie à toute la chaîne. Le gain en mémoire (4×) et en temps de tri est significatif sur des dizaines de millions de super-kmers.

### 3.3 Canonisation des super-kmers : pourquoi ça marche

**Point crucial** : les super-kmers doivent être construits en utilisant le minimiseur **non-canonique** (le m-mer brut tel qu'il apparaît dans la séquence), et non le minimiseur canonique `min(fwd, rc)`.

**Pourquoi ?** Si on utilise le minimiseur canonique comme critère de regroupement, un même super-kmer pourrait contenir le minimiseur dans ses **deux orientations** à des positions différentes (le m-mer forward à une position, et sa forme RC à une autre position, ayant la même valeur canonique). Dans ce cas, le RC du super-kmer ne résoudrait pas l'ambiguïté.

**Algorithme correct** :

1. **Extraction** : construire les super-kmers en regroupant les k-mers consécutifs qui partagent le même m-mer minimal **non-canonique** (le m-mer brut). Au sein d'un tel super-kmer, le minimiseur apparaît toujours dans **une seule orientation**.

2. **Canonisation** : pour chaque super-kmer, comparer son minimiseur brut à `canonical(minimizer) = min(minimizer, RC(minimizer))` :
   - Si `minimizer == canonical(minimizer)` → le minimiseur est déjà en forward → garder le super-kmer tel quel
   - Si `minimizer != canonical(minimizer)` → le minimiseur est en RC → RC le super-kmer entier → le minimiseur apparaît maintenant en forward

Après cette étape, **chaque k-mer du super-kmer** contient le minimiseur canonique en position forward, ce qui correspond exactement à notre définition de k-mer minimizer-canonique.

**Note** : cela signifie que l'algorithme `IterSuperKmers` actuel (qui utilise le minimiseur canonique pour le regroupement) doit être modifié pour utiliser le minimiseur brut. C'est un changement dans le critère de rupture des super-kmers : on casse quand le **m-mer minimal brut** change, pas quand le **m-mer minimal canonique** change. Les super-kmers résultants seront potentiellement plus courts (un changement d'orientation du minimiseur force une coupure), mais c'est le prix de la canonicité absolue.

### 3.4 Déréplication des super-kmers

Deux super-kmers identiques (même séquence, même minimiseur) correspondent aux mêmes k-mers. On peut les dérépliquer en triant :

1. Par minimiseur (déjà partitionné)
2. Par séquence (tri lexicographique des séquences 2-bit packed)

Les super-kmers identiques deviennent consécutifs dans le tri → comptage linéaire.

Le tri peut se faire sur les fichiers .skm d'une partition, en mémoire si la partition tient en RAM, ou par merge-sort externe sinon.

## 4. Filtre de fréquence

### 4.1 Problème

Le filtre de fréquence (`--min-occurrence N`) élimine les k-mers vus moins de N fois. Avec la déréplication des super-kmers, on a un count par super-kmer, pas par k-mer. Un k-mer peut apparaître dans plusieurs super-kmers différents (aux jonctions, ou quand le minimiseur change), donc le count exact d'un k-mer n'est connu qu'après fusion.

### 4.2 Solution : filtrage sur le graphe de De Bruijn pondéré

Le filtre de fréquence doit être appliqué **après** la construction des unitigs canoniques (section 5), et non avant. Le pipeline devient :

```
Super-kmers canoniques dérepliqués (avec counts)
    │
    ▼
Construction des unitigs canoniques (section 5)
    │  Chaque position dans un unitig porte un poids
    │  = somme des counts des super-kmers couvrant ce k-mer
    ▼
Graphe de De Bruijn pondéré (implicite dans les unitigs)
    │
    ▼
Filtrage : supprimer les k-mers (positions) avec poids < seuil
    │  Cela casse certains unitigs en fragments
    ▼
Recalcul des unitigs sur le graphe filtré
    │
    ▼
Unitigs filtrés finaux
```

**Avantages** :
- Le filtre opère sur les **k-mers exacts** avec leurs **counts exacts** (pas une approximation par super-kmer)
- Le graphe de De Bruijn est implicitement contenu dans les unitigs — pas besoin de le construire explicitement avec une `map[uint64]uint`
- Les k-mers aux jonctions de super-kmers ont leurs counts correctement agrégés

### 4.3 Calcul du poids de chaque position dans un unitig

Un unitig est construit par chaînage de super-kmers. Chaque super-kmer S de longueur L et count C contribue (L-k+1) k-mers, chacun avec poids C. Quand deux super-kmers se chevauchent de (k-1) bases dans l'unitig, les k-mers de la zone de chevauchement reçoivent la **somme** des counts des deux super-kmers.

En pratique, lors de la construction de l'unitig par chaînage, on construit un vecteur de poids `weights[0..nkmers-1]` :
```
Pour chaque super-kmer S (count=C) ajouté à l'unitig:
  Pour chaque position i couverte par S dans l'unitig:
    weights[i] += C
```

### 4.4 Filtrage et re-construction

Après filtrage (`weights[i] < seuil` → supprimer position i), l'unitig est potentiellement coupé en fragments. Chaque fragment continu de positions conservées forme un nouvel unitig (ou super-kmer si court).

Le recalcul des unitigs après filtrage est trivial : les fragments sont déjà des chemins linéaires, il suffit de vérifier les conditions de non-branchement aux nouvelles extrémités.

### 4.5 Spectre de fréquence

Le spectre de fréquence exact peut être calculé directement depuis les vecteurs de poids des unitigs : `weights[i]` donne le count exact du k-mer à la position i. C'est un histogramme sur toutes les positions de tous les unitigs.

### 4.6 Faisabilité mémoire : graphe pondéré par partition

Données mesurées sur un index *Betula* (k=31, P=4096 partitions, 1 set, génome assemblé) — distribution des tailles de fichiers .kdi :

| Métrique | Taille fichier .kdi | K-mers estimés (~5 B/kmer) | Super-kmers (~8 kmer/skm) |
|----------|--------------------|-----------------------------|---------------------------|
| Mode | 100-200 Ko | 20 000 – 40 000 | 2 500 – 5 000 |
| Médiane | ~350-400 Ko | ~70 000 – 80 000 | ~9 000 – 10 000 |
| Max | ~2.3 Mo | ~460 000 | ~57 000 |

Le graphe de De Bruijn pondéré pour une partition nécessite d'extraire tous les k-mers (arêtes) et (k-1)-mers (nœuds) des super-kmers :

| Partition | K-mers (arêtes) | RAM arêtes (~20 B) | RAM nœuds (~16 B) | Total |
|-----------|-----------------|--------------------|--------------------|-------|
| Typique (~10K skm, 38 bases avg) | ~80K | ~1.6 Mo | ~1.3 Mo | **~3 Mo** |
| Maximale (~57K skm) | ~460K | ~9.2 Mo | ~7.4 Mo | **~17 Mo** |

C'est **largement en mémoire**. Les partitions étant indépendantes, elles peuvent être traitées en parallèle par un pool de goroutines. Avec 8 goroutines : **~136 Mo** au pic — négligeable. Les tableaux sont réutilisables entre partitions (allocation unique).

**Conclusion** : la construction du graphe de De Bruijn pondéré partition par partition est non seulement faisable mais triviale en termes de mémoire. C'est un argument fort en faveur de l'approche « filtre après unitigs » plutôt que « filtre sur super-kmers ».

### 4.7 Invariance de la distribution par rapport à la canonisation

La redéfinition du k-mer canonique (par le minimiseur au lieu de `min(fwd, rc)`) ne change **rien** à l'ensemble des k-mers ni à leur répartition par partition :

- C'est une **bijection** : chaque k-mer a toujours exactement un représentant canonique, on change juste lequel des deux brins on choisit
- Le partitionnement se fait sur `canonical(minimizer) % P` — la valeur du minimiseur canonique est la même dans les deux conventions
- **Même nombre de k-mers par partition, même distribution de tailles**
- **Même topologie du graphe de De Bruijn** (mêmes nœuds, mêmes arêtes)

Ce qui change, c'est **l'orientation** : avec la canonicité par minimiseur, les unitigs canoniques ne suivent que les arêtes « forward » (`suffix(S1) == prefix(S2)`, identité exacte). Certaines arêtes traversables en RC dans BCALM2 deviennent des points de cassure. Le graphe n'est pas plus gros — il suffit de ne construire que des unitigs canoniques, ce qui **simplifie** l'algorithme (pas de gestion des traversées de brin).

## 5. Construction des unitigs canoniques

### 5.1 Définition : unitig canonique absolu

Un **unitig canonique** est un chemin linéaire non-branchant dans le graphe de De Bruijn où :
1. Chaque k-mer est **minimizer-canonique** (le minimiseur y apparaît en forward)
2. Chaque super-kmer constituant est **canonique** (même convention)
3. Le chaînage se fait **sans traversée de brin** : `suffix(k-1, S1) == prefix(k-1, S2)` dans le même sens (pas en RC)

C'est plus restrictif que les unitigs BCALM2 (qui autorisent `suffix(S1) == RC(prefix(S2))`), mais cela garantit que **tout k-mer extrait par fenêtre glissante est directement dans sa forme canonique**, sans re-canonisation.

### 5.2 Pourquoi la canonicité absolue est essentielle

**Matching** : les k-mers requête sont canonisés une fois (par le minimiseur), puis comparés directement aux k-mers de l'unitig par scan. Pas de re-canonisation à la volée → plus rapide, plus simple.

**Opérations ensemblistes** : deux index utilisant la même convention produisent les mêmes unitigs canoniques pour les mêmes k-mers. L'intersect/union peut opérer par comparaison directe de séquences triées.

**Bitmask multiset** : la fusion de N sets est triviale — merger des listes de super-kmers/unitigs canoniques triés par séquence.

**Déterminisme** : un ensemble de k-mers produit toujours les mêmes unitigs canoniques, quel que soit l'ordre d'insertion ou la source des données.

### 5.3 Impact sur la compaction

La contrainte canonique interdit les traversées de brin aux jonctions → les unitigs canoniques sont **plus courts** que les unitigs BCALM2. Estimation :
- BCALM2 (unitigs libres) : 63 bases moyennes (mesuré sur *Betula*)
- Unitigs canoniques : probablement ~45-55 bases moyennes
- Super-kmers dérepliqués : 38 bases moyennes

Le facteur de compaction est légèrement réduit mais le gain en simplicité opérationnelle compense largement.

### 5.4 Construction par partition — le vrai graphe de De Bruijn

Les super-kmers canoniques dérepliqués sont des **chemins** dans le graphe de De Bruijn, pas des nœuds. On ne peut pas les chaîner directement comme des nœuds car :
- Deux super-kmers peuvent **se chevaucher** (partager des k-mers aux jonctions)
- Un super-kmer court peut avoir ses k-mers **inclus** dans un super-kmer plus long

Un super-kmer de longueur L contient (L-k+1) k-mers, soit (L-k+1) arêtes et (L-k+2) nœuds ((k-1)-mers) dans le graphe de De Bruijn.

#### 5.4.1 Nœuds = (k-1)-mers, Arêtes = k-mers

Le graphe de De Bruijn par partition a :
- **Nœuds** : les (k-1)-mers uniques (extraits de toutes les positions dans les super-kmers)
- **Arêtes** : les k-mers (chaque position dans un super-kmer = une arête entre deux (k-1)-mers consécutifs)
- **Poids** : chaque arête (k-mer) porte le count du super-kmer qui la contient

Les branchements (nœud avec degré entrant > 1 ou degré sortant > 1) peuvent être :
- Aux **bords** des super-kmers (jonctions entre super-kmers)
- Aux **positions internes** si un k-mer d'un autre super-kmer rejoint un (k-1)-mer interne

#### 5.4.2 Graphe complet nécessaire

Construire le graphe avec **tous** les (k-1)-mers (internes et bords) est nécessaire pour détecter correctement les branchements. Se limiter aux seuls bords de super-kmers serait incorrect car un (k-1)-mer de bord d'un super-kmer peut correspondre à un nœud interne d'un autre super-kmer.

Pour une partition typique de 10K super-kmers de longueur moyenne 38 bases → ~80K k-mers → ~80K arêtes et ~80K nœuds. Voir section 4.6 pour la faisabilité mémoire (~3 Mo par partition typique, ~17 Mo max).

#### 5.4.3 Structure de données : tableau trié d'arêtes

Plutôt que des hash maps, on utilise un **tableau trié** pour le graphe :

```go
type Edge struct {
    srcKmer  uint64    // (k-1)-mer source (prefix du k-mer)
    dstKmer  uint64    // (k-1)-mer destination (suffix du k-mer)
    weight   int32     // count du super-kmer contenant ce k-mer
}
```

Pour chaque super-kmer S de longueur L et count C, on émet (L-k+1) arêtes. Tableau total pour une partition typique : ~80K × 20 bytes = **~1.6 Mo**.

On trie par `srcKmer` pour obtenir la liste d'adjacence sortante, ou on construit deux vues triées (par src et par dst) pour avoir adjacence entrante et sortante.

#### 5.4.4 Détection des unitigs canoniques

Un unitig canonique est un chemin maximal non-branchant. L'algorithme :

```
1. Extraire toutes les arêtes des super-kmers → tableau edges[]
2. Trier edges[] par srcKmer → vue sortante
   Trier une copie par dstKmer → vue entrante
3. Pour chaque (k-1)-mer unique :
   - degré_sortant = nombre d'arêtes avec ce srcKmer
   - degré_entrant = nombre d'arêtes avec ce dstKmer
   - Si degré_sortant == 1 ET degré_entrant == 1 → nœud interne d'unitig
   - Sinon → nœud de branchement (début ou fin d'unitig)
4. Parcourir les chemins non-branchants pour construire les unitigs
   - Chaque unitig est une séquence de (k-1)-mers chaînés
   - Le vecteur de poids est la séquence des weight des arêtes traversées
```

Les (k-1)-mers ne sont **pas canonisés** — on respecte l'orientation des super-kmers canoniques. Le chaînage est strictement orienté.

#### 5.4.5 Estimation mémoire

| Partition | K-mers (arêtes) | RAM arêtes | RAM nœuds | Total |
|-----------|-----------------|------------|-----------|-------|
| Typique (~10K skm, 38 bases avg) | ~80K | ~1.6 Mo | ~1.3 Mo | **~3 Mo** |
| Maximale (~57K skm) | ~460K | ~9.2 Mo | ~7.4 Mo | **~17 Mo** |

Avec traitement parallèle par un pool de G goroutines : RAM max = G × 17 Mo. Avec G=8 : **~136 Mo** au pic. Le tableau d'arêtes est réutilisable entre partitions (allocation unique, remise à zéro).

Complexité : O(E log E) par partition, avec E = nombre total de k-mers. Dominé par les deux tris.

### 5.5 Graphe par minimiseur, pas par partition

Les k-mers (arêtes) sont partitionnés par minimiseur. Deux k-mers adjacents dans le graphe de De Bruijn peuvent avoir des minimiseurs différents — c'est exactement ce qui définit les frontières de super-kmers. Si on construit le graphe par partition (qui regroupe plusieurs minimiseurs), des (k-1)-mers de jonction entre minimiseurs différents apparaîtraient comme nœuds partagés entre partitions → le graphe par partition n'est pas autonome.

**Solution : construire un graphe par minimiseur.**

Un super-kmer est par définition un chemin dont **tous les k-mers partagent le même minimiseur**. Donc :
- Toutes les arêtes d'un super-kmer appartiennent à un seul minimiseur
- Chaque graphe par minimiseur est **100% autonome** : toutes ses arêtes et nœuds internes sont auto-contenus
- Les (k-1)-mers aux bords des super-kmers qui touchent un autre minimiseur sont des extrémités (degré 0 dans ce graphe) → bouts d'unitig naturels
- Aucune jonction inter-graphe → pas de cassure artificielle d'unitig

**Taille des graphes** : avec ~16K minimiseurs théoriques par partition (P=4096, m=13), le calcul naïf donne ~5 arêtes/minimiseur. Mais en pratique, beaucoup de minimiseurs ne sont pas représentés (séquences biologiques, pas aléatoires) et la distribution est très inégale. Si seuls ~500-1000 minimiseurs sont effectivement présents dans une partition typique de 80K arêtes, on a plutôt **80-160 arêtes en moyenne** par minimiseur, avec une queue de distribution vers les centaines ou milliers pour les minimiseurs les plus fréquents. Même dans ce cas, les graphes restent petits (quelques Ko à quelques dizaines de Ko).

*À mesurer* : nombre de minimiseurs distincts par partition et distribution du nombre de k-mers par minimiseur sur un index existant.

**Algorithme** : les super-kmers étant déjà triés par minimiseur dans la partition, on itère séquentiellement et on construit/détruit un petit graphe à chaque changement de minimiseur. C'est plus simple que le graphe par partition — pas de tri global de toutes les arêtes, juste un buffer local réutilisé.

**Les unitigs résultants** sont les chemins maximaux non-branchants au sein d'un minimiseur. Un unitig ne traverse jamais une frontière de minimiseur, ce qui est correct : tous les k-mers d'un unitig partagent le même minimiseur canonique, ce qui renforce la propriété de canonicité absolue.

### 5.6 Quand les unitigs canoniques n'aident pas

- Si les super-kmers sont courts (peu de chevauchement entre super-kmers adjacents)
- Si le graphe est très branché (zones de divergence entre génomes)
- Si beaucoup de jonctions se font par traversée de brin (la contrainte canonique empêche la fusion)
- Données metabarcoding avec grande diversité taxonomique → courts unitigs

Dans ces cas, stocker les super-kmers dérepliqués directement est suffisant — ils sont déjà canoniques par construction.

## 6. Construction multiset : super-kmers par set, graphe commun

### 6.1 Pipeline en deux phases

La construction d'un index multiset (N sets) se fait en deux phases distinctes :

**Phase 1 — Par set (indépendant, parallélisable)** :

Chaque set i (i = 0..N-1) produit indépendamment ses super-kmers canoniques :
```
Set i : séquences → [0] 2-bit → [1] lowmask → [2] super-kmers → [3] canonisation → [4] partition .skm_i
```
Puis déréplication par partition : super-kmers triés avec counts, écrits dans des fichiers `.skm` distincts par set.

**Phase 2 — Par partition, tous sets confondus** :

Pour chaque partition P (parallélisable par goroutine) :
```
.skm_0[P], .skm_1[P], ..., .skm_{N-1}[P]   (super-kmers triés de chaque set)
    │
    ▼
[a] N-way merge des super-kmers triés
    │  Même super-kmer dans sets i et j → fusionner en vecteur de counts [c_0, ..., c_{N-1}]
    │  Super-kmer uniquement dans set i → counts = [0, ..., c_i, ..., 0]
    ▼
[b] Extraction des arêtes du graphe de De Bruijn
    │  Chaque k-mer (arête) porte un vecteur de poids [w_0, ..., w_{N-1}]
    │  w_i = count du super-kmer contenant ce k-mer dans le set i
    ▼
[c] Construction d'un SEUL graphe de De Bruijn par partition
    │  Les branchements sont définis par l'UNION de tous les sets :
    │  si un (k-1)-mer a degré > 1 dans n'importe quel set, c'est un branchement
    ▼
[d] Extraction des unitigs canoniques communs
    │  Même séquence pour tous les sets
    │  Chaque position porte un vecteur de poids (un count par set)
    ▼
[e] Filtre de fréquence (optionnel, par set ou global)
    ▼
[f] Encodage du bitmask par runs le long de chaque unitig
    ▼
Écriture dans le .sku de la partition
```

### 6.2 Le graphe est défini par l'union

Point crucial : les unitigs sont déterminés par la **topologie de l'union** de tous les sets. Un branchement dans un seul set force une coupure d'unitig pour tous les sets. Cela garantit que :
- Les unitigs sont les mêmes quelle que soit l'ordre des sets
- Un k-mer donné se trouve toujours au même endroit (même unitig, même position)
- Les opérations ensemblistes (intersect, union, difference) opèrent sur les mêmes unitigs

### 6.3 Arêtes à vecteur de poids

La structure Edge (section 5.4.3) est étendue pour le multiset :

```go
type Edge struct {
    srcKmer  uint64      // (k-1)-mer source
    dstKmer  uint64      // (k-1)-mer destination
    weights  []int32     // weights[i] = count dans le set i (0 si absent)
}
```

Pour la détection des branchements, le degré d'un nœud est le nombre d'arêtes distinctes (par dstKmer pour le degré sortant), **indépendamment** des sets. Une arête présente dans le set 0 mais pas le set 1 compte quand même.

### 6.4 Bitmask par runs le long des unitigs

Le long d'un unitig, le bitmask (quels sets contiennent ce k-mer) change rarement — les régions conservées entre génomes sont longues. On encode :

```
unitig_bitmask = [(bitmask_1, run_length_1), (bitmask_2, run_length_2), ...]
```

Où `bitmask_i` a un bit par set (bit j = 1 si `weights[j] > 0` à cette position).

Pour un unitig de 70 k-mers avec 2 sets :
- Si complètement partagé : 1 run `(0b11, 70)` → 2 bytes
- Si divergent au milieu : 2-3 runs → 4-6 bytes
- Pire cas : 70 runs → 140 bytes (très rare)

### 6.5 Impact mémoire du multiset

Le vecteur de poids par arête augmente la taille du graphe :
- 1 set : `weight int32` → 4 bytes/arête
- N sets : `weights [N]int32` → 4N bytes/arête

Pour la partition typique (~80K arêtes) avec N=2 sets : overhead = 80K × 4 = **320 Ko** supplémentaires. Négligeable.

Pour N=64 sets (cas extrême) : 80K × 256 = **~20 Mo** par partition. Reste faisable mais les sets très nombreux pourraient nécessiter un encodage plus compact (sparse vector si beaucoup de zéros).

### 6.6 Merge des super-kmers : N-way sur séquences triées

Le merge des N listes de super-kmers triés (par séquence 2-bit) est un N-way merge classique avec min-heap :
- Chaque .skm est déjà trié par séquence (étape de déréplication)
- On compare les séquences 2-bit packed (comparaison uint64, très rapide)
- Quand le même super-kmer apparaît dans plusieurs sets, on fusionne les counts
- Quand un super-kmer est unique à un set, les autres counts sont 0

C'est analogue au `KWayMerge` existant sur les k-mers triés, étendu aux super-kmers.

## 7. Format de stockage v3 : fichiers parallèles

### 7.1 Architecture : 3 fichiers par partition

Pour chaque partition, trois fichiers alignés :

```
index_v3/
  metadata.toml
  parts/
    part_PPPP.sku         # séquences 2-bit des unitigs concaténés
    part_PPPP.skx         # index par minimiseur (offsets dans .sku)
    part_PPPP.skb         # bitmask multiset (1 entrée par k-mer)
    ...
  set_N/spectrum.bin      # spectre de fréquence par set
```

### 7.2 Fichier .sku — séquences d'unitigs concaténées

Tous les unitigs d'une partition sont concaténés bout à bout en 2-bit packed, **ordonnés par minimiseur**. Entre deux unitigs, pas de séparateur dans le flux 2-bit.

Un **tableau de longueurs** stocké en en-tête ou dans le .skx donne la longueur (en bases) de chaque unitig dans l'ordre. Ce tableau permet :
- De retrouver les frontières d'unitigs
- De savoir si un match AC chevauche une jonction (à filtrer)
- D'indexer directement un unitig par son numéro

```
Format .sku :
  Magic: "SKU\x01"           (4 bytes)
  TotalBases: uint64 LE       (nombre total de bases dans la partition)
  NUnitigs: uint64 LE         (nombre d'unitigs)
  Lengths: [NUnitigs]varint   (longueur en bases de chaque unitig)
  Sequence: ceil(TotalBases/4) bytes  (flux 2-bit continu)
```

### 7.3 Fichier .skx — index par minimiseur

Pour chaque minimiseur présent dans la partition, l'index donne l'offset (en bases) dans le flux .sku et le nombre d'unitigs :

```
Format .skx :
  Magic: "SKX\x01"           (4 bytes)
  NMinimizers: uint32 LE      (nombre de minimiseurs présents)
  Entries: [NMinimizers] {
    Minimizer: uint64 LE      (valeur du minimiseur canonique)
    BaseOffset: uint64 LE     (offset en bases dans le flux .sku)
    UnitigOffset: uint32 LE   (index du premier unitig de ce minimiseur dans le tableau de longueurs)
    NUnitigs: uint32 LE       (nombre d'unitigs pour ce minimiseur)
  }
```

Les entrées sont triées par `Minimizer` → recherche binaire en O(log N).

Pour accéder aux unitigs d'un minimiseur donné :
1. Recherche binaire dans le .skx → `BaseOffset`, `UnitigOffset`, `NUnitigs`
2. Seek dans le .sku au bit `BaseOffset × 2`
3. Lecture de `NUnitigs` unitigs (longueurs dans le tableau à partir de `UnitigOffset`)

### 7.4 Fichier .skb — bitmask multiset parallèle

Le fichier bitmask est **aligné position par position** avec le flux de k-mers des unitigs. Chaque k-mer (position dans un unitig) a exactement une entrée dans le .skb, dans le même ordre que les k-mers apparaissent en lisant les unitigs séquentiellement.

```
Format .skb :
  Magic: "SKB\x01"           (4 bytes)
  TotalKmers: uint64 LE       (nombre total de k-mers)
  NSets: uint8                 (nombre de sets)
  BitmaskSize: uint8           (ceil(NSets/8) bytes par entrée)
  Bitmasks: [TotalKmers × BitmaskSize] bytes
```

**Accès direct** : la position absolue d'un k-mer dans le flux d'unitigs (offset en k-mers depuis le début de la partition) donne directement l'index dans le fichier .skb :
```
bitmask_offset = header_size + kmer_position × BitmaskSize
```

Pour 2 sets : 1 byte par k-mer (6 bits inutilisés).
Pour ≤8 sets : 1 byte par k-mer.
Pour ≤16 sets : 2 bytes par k-mer.

**Coût** : pour 295M k-mers (*Betula*, 2 sets) : 295 Mo. Pour l'index Contaminent_idx (18.6B k-mers, 2 sets) : ~18.6 Go. C'est significatif — voir section 7.5 pour la compression.

### 7.5 Compression du bitmask : RLE ou non ?

| Approche | Taille (2 sets, 295M k-mers) | Accès |
|----------|------------------------------|-------|
| Non compressé (1 byte/k-mer) | 295 Mo | O(1) direct |
| RLE par unitig | ~10-50 Mo (estimé) | O(decode) par unitig |
| Bitset par set (1 bit/k-mer/set) | 74 Mo | O(1) direct |

L'approche **bitset par set** (1 bit par k-mer par set, packed en bytes) est un bon compromis :
- 2 sets : 2 bits/k-mer → ~74 Mo (vs 295 Mo non compressé)
- Accès O(1) : `bit = (data[kmer_pos / 4] >> ((kmer_pos % 4) × 2)) & 0x3`
- Pas besoin de décompression séquentielle

Pour les très grands index (18.6B k-mers), même le bitset fait ~4.6 Go. Le RLE par minimiseur (ou par unitig) pourrait réduire à ~1-2 Go mais perd l'accès O(1).

**Recommandation** : bitset packed pour ≤8 sets (accès O(1)), RLE pour >8 sets ou très grands index.

## 8. Matching avec Aho-Corasick sur le flux d'unitigs

### 8.1 Principe

Pour chaque partition dont les k-mers requête partagent le minimiseur :
1. Seek dans le .sku au bloc du minimiseur (via .skx)
2. Construire un automate AC avec les k-mers requête canoniques de ce minimiseur
3. Scanner le flux 2-bit des unitigs de ce minimiseur
4. Pour chaque match : vérifier qu'il ne chevauche pas une frontière d'unitig
5. Pour chaque match valide : lookup dans le .skb à la position correspondante → bitmask

### 8.2 Le problème des faux matches aux jonctions

En 2-bit, pas de 5e lettre pour séparer les unitigs. Le scan AC sur le flux continu peut produire des matches à cheval sur deux unitigs adjacents.

**Solution : post-filtrage par le tableau de longueurs.**

Pendant le scan, on maintient un compteur de position et un index dans le tableau de longueurs (préfixe cumulé). Quand un match est trouvé à la position `p` :
- Le match couvre les bases `[p, p+k-1]`
- Si ces bases chevauchent une frontière d'unitig → faux positif, ignorer
- Sinon → match valide

Le coût du post-filtrage est O(1) par match (le compteur de frontière avance séquentiellement).

**Estimation du taux de faux positifs** : avec des unitigs de ~50 bases en moyenne, une jonction tous les ~50 bases, et k=31 : ~31/50 = ~62% des positions de jonction peuvent produire un faux match. Mais seule une infime fraction de ces positions correspond à un pattern dans l'automate AC. En pratique, le nombre de faux positifs est négligeable.

### 8.3 Du match à la position absolue dans le .skb

Un match AC à la position `p` dans le flux du minimiseur se traduit en position k-mer dans le .skb :

```
kmer_position_in_partition = base_offset_of_minimizer_in_partition
                           + p
                           - (nombre de bases de padding/frontières avant p)
```

En fait, si le tableau de longueurs donne les longueurs d'unitigs en bases, la position k-mer cumulative est :
```
Pour l'unitig i contenant le match :
  kmer_base = somme des longueurs des unitigs 0..i-1
  kmer_offset_in_unitig = p - kmer_base
  kmer_index = somme des (len_j - k + 1) pour j=0..i-1  +  kmer_offset_in_unitig
```

Ce `kmer_index` est l'index direct dans le fichier .skb.

### 8.4 Comparaison avec le merge-scan v1

| Aspect | Merge-scan (v1) | AC sur unitigs (v3) |
|--------|----------------|---------------------|
| Pré-requis | Tri des requêtes O(Q log Q) | Construction automate AC O(Q×k) |
| Seek | .kdx sparse index | .skx index par minimiseur |
| Scan | O(Q + K) merge linéaire par set | O(bases_du_minimiseur + matches) |
| Multi-set | **N passes** (une par set) | **1 seule passe** (bitmask .skb) |
| I/O | N×P ouvertures de fichier | 1 seek + lecture séquentielle + lookup .skb |
| Accès bitmask | implicite (chaque .kdi = 1 set) | O(1) dans .skb |

Le gain principal du v3 est l'**élimination des N passes** : au lieu de scanner N fois (une par set), on scanne une seule fois et on consulte le bitmask. Pour N=2 sets et P=4096 partitions, cela réduit les ouvertures de fichier de 2×4096 = 8192 à 4096.

## 9. Estimations de taille et validation expérimentale

### 9.1 Cas mesuré : *Betula exilis* 15× (reads bruts, count > 1)

| Métrique | Valeur |
|----------|--------|
| Super-kmers uniques (count > 1) | 37.3M |
| Longueur moyenne | 37.9 bases |
| Bases totales | 1.415G |

**Stockage binaire 2-bit packed** :
- Séquences : 1.415G / 4 = **354 Mo**
- Headers (longueur varint + minimiseur) : 37.3M × ~4 bytes = **150 Mo**
- Bitmask (1 set → 0 bytes, ou 2 sets → 1 byte/entrée = 37 Mo)
- **Total estimé : ~500-550 Mo** pour un set

### 9.2 Extrapolation pour l'index Plants+Human (2 sets)

L'index v1 actuel contient 18.6B k-mers en 85 Go. Avec le pipeline v3 :

**Scénario reads bruts 15× par génome** (extrapolé depuis *Betula exilis*) :
- *Betula exilis* mesuré : ~37M super-kmers, ~1.4G bases → ~500 Mo
- Proportionnellement pour l'index Contaminent_idx (18.6B k-mers) : **~2-5 Go**

**Scénario génome assemblé (pas de filtre de fréquence)** :
- Un génome assemblé de 3 Gbases → estimation ~80M super-kmers × 38 bases → **760 Mo**
- Un génome assemblé de 10 Gbases → estimation ~350M super-kmers × 38 bases → **3.3 Go**
- Avec overlap multiset : super-kmers partagés fusionnés (bitmask) → **~4 Go**

**Le gain est spectaculaire dans les deux scénarios** :
- Reads bruts : facteur **~30-40×** grâce à la déréplication + filtre de fréquence
- Génomes assemblés : facteur **~20×** grâce au format super-kmer seul

Le format super-kmer est intrinsèquement plus efficace que le delta-varint car il exploite la structure locale du graphe de De Bruijn : des k-mers consécutifs partagent (k-1) bases, encodées une seule fois dans le super-kmer.

### 9.3 Validation expérimentale : unitigs BCALM2

*Betula exilis* 15×, après lowmask + super-kmers canoniques + déréplication + filtre count>1, passé dans BCALM2 (`-kmer-size 31 -abundance-min 1`) :

| Métrique | Super-kmers (count>1) | Unitigs (BCALM2) | Ratio |
|----------|----------------------|-------------------|-------|
| Variants | 37,294,271 | 6,473,171 | **5.8×** |
| Bases totales | 1,415,018,593 | 408,070,894 | **3.5×** |
| Longueur moyenne | 37.9 bases | 63.0 bases | 1.7× |
| K-mers estimés | ~295M | ~213M | — |

### Stockage estimé

| Format | Taille estimée | Bytes/k-mer | Facteur vs v1 |
|--------|---------------|-------------|---------------|
| .kdi v1 (delta-varint, assemblé) | 12.8 Go | 4.3 | 1× |
| Super-kmers 2-bit (count>1) | ~500 Mo | 1.7 | 25× |
| **Unitigs 2-bit (BCALM2)** | **~130 Mo** | **0.6** | **98×** |

### Extrapolation pour l'index Contaminent_idx (Plants+Human, 2 sets)

Le facteur ~100× mesuré sur *Betula exilis* 15× se décompose :
- Déréplication des reads redondants : facteur ~15× (couverture 15×)
- Compaction super-kmer/unitig vs delta-varint : facteur ~100/15 ≈ **6.7×**

L'index Contaminent_idx est construit à partir de **génomes assemblés** (sans redondance de séquençage). Seul le facteur de compaction unitig s'applique :
- Index v1 actuel : 85 Go (Plants 72 Go + Human 12.8 Go)
- **Estimation unitigs : ~85 / 6.7 ≈ 12-13 Go** (facteur **~6.7×**)

C'est un gain significatif mais bien moins spectaculaire que sur des reads bruts. Le facteur pourrait être meilleur si les unitigs des génomes assemblés sont plus longs que ceux des reads (moins de fragmentation par les erreurs de séquençage).

### Observation sur le nombre de k-mers

Les unitigs contiennent ~213M k-mers vs ~295M estimés dans les super-kmers. La différence (~80M) provient probablement de k-mers qui étaient comptés dans plusieurs super-kmers (aux jonctions) et qui ne sont comptés qu'une fois dans les unitigs (déduplication exacte par le graphe de De Bruijn).

### Conclusion

L'approche unitig est massivement plus compacte que toutes les alternatives. Le format de stockage final devrait être basé sur les unitigs (ou au minimum sur les super-kmers dérepliqués) plutôt que sur des k-mers individuels en delta-varint.

## 10. Questions ouvertes

### 10.1 Le format super-kmer est-il toujours meilleur que delta-varint ?

D'après les estimations révisées (section 8.3), le format super-kmer 2-bit est **toujours plus compact** que le delta-varint, même pour des génomes assemblés :
- Reads bruts 15× : ~500 Mo vs ~1.5 Go (facteur 3×, à k-mers égaux) + déréplication massive
- Génomes assemblés : ~1.2 bytes/k-mer vs ~5 bytes/k-mer (facteur 4×)

La raison fondamentale : le delta-varint encode chaque k-mer indépendamment (même avec deltas), tandis que le super-kmer exploite le chevauchement de (k-1) bases entre k-mers consécutifs. C'est un avantage structurel irrattrapable par le delta-varint.

**Le format super-kmer semble donc préférable dans tous les cas.**

### 10.2 L'index doit-il stocker les super-kmers ou les k-mers ?

Stocker les super-kmers/unitigs comme format d'index final a des avantages (compacité, scan naturel) mais des inconvénients :
- Pas de seek rapide vers un k-mer spécifique (vs .kdx sparse index)
- Le matching par scan complet est O(total_bases) vs O(Q + K) pour le merge-scan
- Les opérations ensemblistes (Union, Intersect) deviennent plus complexes

**Approche hybride possible** :
1. Phase de construction : lowmask → super-kmers canoniques → déréplication → filtre de fréquence
2. Phase de finalisation : extraire les k-mers uniques des super-kmers filtrés → delta-varint .kdi (v1 ou v2)
3. Les super-kmers servent de **format intermédiaire efficace**, pas de format d'index final

Cela combine le meilleur des deux mondes :
- Déréplication ultra-efficace au niveau super-kmer (facteur 16× sur reads bruts)
- Index final compact et query-efficient en delta-varint

### 10.3 Le filtre de fréquence simple (niveau super-kmer) est-il suffisant ?

À valider expérimentalement :
- Comparer le nombre de k-mers retenus par filtre super-kmer vs filtre k-mer exact
- Mesurer l'impact sur les métriques biologiques (Jaccard, match positions)
- Si la différence est <1%, le filtre simple suffit

### 10.4 Aho-Corasick vs merge-scan pour le matching final ?

Si le format d'index final reste delta-varint (question 9.2), le merge-scan reste la méthode naturelle de matching. L'AC/hash-set n'a d'intérêt que si le format de stockage est basé sur des séquences (unitigs/super-kmers).

## 11. Prochaine étape : validation expérimentale

Avant de modifier l'architecture, valider sur des données réelles :

1. **Taux de compaction super-kmer** : sur un génome assemblé vs reads bruts, mesurer le nombre de super-kmers uniques et leur longueur moyenne
2. **Impact du filtre super-kmer** : comparer filtre au niveau super-kmer vs filtre au niveau k-mer exact sur un jeu de données de référence
3. **Taux d'assembly en unitigs** : mesurer la longueur des unitigs obtenus à partir des super-kmers dérepliqués
4. **Benchmark stockage** : comparer taille index super-kmer vs delta-varint vs unitig sur les mêmes données
5. **Benchmark matching** : comparer temps de matching AC/hash vs merge-scan sur différentes densités de requêtes
