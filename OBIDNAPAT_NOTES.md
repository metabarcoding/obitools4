# obidnapat — remplaçant Go pur d'obiapat, spécialisé ADN

Paquet `pkg/obidnapat`. Dernière mise à jour : 2026-11-05. Contexte et semantique d'`obiapat` : annexe D de
`OBIPRIMERS_NOTES.md`. **Rien n'est commité** ; aucun appelant n'a été migré.

## Décisions de l'utilisateur

- Limite de **63 positions** pour le **cœur** du motif (jokers `N` de tête et de queue mis à part) ; au-delà, une erreur.
- Moteur spécialisé ADN, Go pur, sans cgo. SIMD réservé au multi-motifs, après le noyau scalaire.
- Une base ambiguë de la séquence correspond si ses bases possibles **recoupent** l'ensemble accepté (`n` : tout, `r` : A ou G,
  `u` = `t`). L'ancienne règle (ambiguïté = erreur) est considérée comme un **défaut**, pas une propriété : il n'y a pas de mode ancien.
- `ApatSequence` disparaît : les méthodes prennent `*obiseq.BioSequence` ; variantes `...Circular` ; formes `Append...` pour réutiliser un tampon.
- Les motifs `NNNNmotifNNNN` sont recodés (cœur + longueurs de flancs).
- Ordre de travail : Hamming (substitutions, seul mode d'`obipcr`) comme laboratoire, puis `ManberIndel`. **Garder une voie stricte**
  (`e = 0`) : un algorithme strict est plus efficace que Manber, notamment en multi-motifs.
- Convention circulaire : une occurrence par site, `start ∈ [0, n)`, `end` pouvant dépasser `n` ; un motif plus long que la séquence circulaire ne donne rien.

## État (substitutions seules)

Fichiers : `syntax.go` (IUPAC, classes `[AG]`, négation `!`, `#`), `pattern.go` (`MakeApatPattern`, `ReverseComplement`, `String`, `Len`),
`scan.go` (noyaux), `search.go` (`AppendAllIndex[Circular]`, `FindAllIndex[Circular]`, `IsMatching[Circular]`, `BestMatch`, `FilterBestMatch`, `AllMatches`).
`MakeApatPattern(..., allowsIndel=true)` renvoie une erreur (non disponible).

Sémantique : un octet qui n'est pas un code IUPAC ne représente aucune base (il ne correspond qu'à une position joker) ; une position qui accepte
les quatre bases est un joker (`N`, `X`, `[ACGT]`) et accepte **tout octet** ; la région `[begin, begin+length)` est stricte pour une séquence linéaire
(obiapat dépassait de 64 octets, sans que `pcr.go` en dépende).

Voie stricte (`e = 0`, `strict.go`) : **BNDM** (recherche par facteur, lecture de chaque fenêtre à reculons, sauts jusqu'à la longueur du motif) pour une séquence
linéaire quand l'avance estimée `advance()` (longueur du cœur / bases lues par fenêtre, sur un texte uniforme) vaut au moins `minAdvance = 3,6` ; sinon, et pour
une séquence circulaire, Shift-And. Même table de masques : ensembles IUPAC, classes, jokers internes et règle de recoupement sont gérés sans rien de plus.

Noyaux avec erreurs : Shift-And de Baeza-Yates/Gonnet étendu par Wu-Manber ; `e = 0` écrit à la main (`run0`), `e = 1` à `8` **générés** (`gen_kernels.go`, `go generate`,
résultat `scan_unrolled.go`) avec un état par variable, donc en registres ; noyau générique à tableau pour `e ≥ 9`. `maxUnrolled` se règle dans le générateur.

### Voie stricte : BNDM contre Shift-And (Mo/s, 1 Mb, `BenchmarkStrict`)

`obiapat` et le Shift-And restent à ~1 220 Mo/s quelle que soit la longueur ; BNDM dépend de la longueur et de la dégénérescence :

| Motif | BNDM | rapport au Shift-And |
|---|---|---|
| 4 / 6 / 8 positions | 545 / 920 / 1 300 | 0,45 / 0,75 / 1,06 |
| 10 / 12 / 16 | 1 400-1 530 / 1 600 / 2 100-2 300 | 1,2 / 1,3 / 1,75 |
| 20 / 30 | 3 050 / 4 500 | 2,5 / 3,7 |
| 50 / 63 | 7 600-10 100 / 10 100-13 100 | 6-8 / 8-11 |
| 20 avec 3 / 6 / 10 codes à deux bases | 2 300-2 900 / 1 950 / 1 700 | 1,9-2,3 / 1,6 / 1,4 |
| 20 avec 16 / 20 codes à deux bases | 1 800 / 1 300 | 1,45 / 1,06 |
| 20 avec 2 `N` internes | 2 300 | 1,9 |
| **20 avec 12 `N` internes** | **1 050** | **0,85 (perte)** |
| 10 avec 3, 12 avec 6 codes à deux bases | 1 180, 1 080 | 0,96, 0,88 (pertes) |

Le seuil `minAdvance = 3,6` sépare ces mesures : tous les gains ≥ 1,2 ont une avance estimée ≥ 3,67, toutes les égalités ou pertes ≤ 3,37. Calibrage sur des motifs
choisis au hasard avec une seule graine ; à reprendre si l'on change de machine ou de modèle de texte. Non mesuré : séquences riches en `N` (un `n` correspond à toute
position, donc BNDM ne s'arrête plus), textes avec beaucoup d'occurrences, circulaire (toujours Shift-And).

## Tests

- `oracle_test.go` : recherche naïve fenêtre par fenêtre, sur séquences avec ambiguïtés et caractères hors alphabet, `e` de 0 à 6, linéaire et circulaire,
  régions, jokers, `#`.
- `diff_test.go` : comparaison avec `obiapat` sur séquences ACGT (là où les deux règles d'ambiguïté ne peuvent pas différer).
- `TestStrictAgainstOracle` : motifs de 1 à 63 positions, `e = 0`, BNDM forcé puis Shift-And forcé, chacun contre l'oracle.
- `pattern_test.go` : syntaxe, ensembles de lettres, limite 63 sur le cœur, complément inverse, cas à la main.
- `OBIDNAPAT_ITER=300000 go test ./pkg/obidnapat/ -run TestAgainst` : 300 000 tirages chacun (`TestAgainst*` et `TestStrict*`), `e` de 0 à 10, **tous verts**, `-race` vert.
- Seule différence connue avec `obiapat` rencontrée : motif plus long que la séquence circulaire (obiapat lit le cercle plusieurs fois).

## Mesures (M3 Max, un cœur, `go test -bench`, un motif de 20 positions, séquence ACGT aléatoire de 1 Mb ; bruit non évalué)

| Cas | obiapat | obidnapat |
|---|---|---|
| `e = 0`, motif de 20 (BNDM) | 1 190 Mo/s | 2 500 à 2 900 Mo/s (×2,3) |
| `e = 1` / `2` / `3` | 330 / 320 / 297 Mo/s | 915 / 838 / 766 Mo/s (×2,8 / ×2,6 / ×2,6) |
| `e = 4` / `5` / `6` | 275 / 238 / 199 Mo/s | 724 / 656 / 596 Mo/s (×2,6 / ×2,8 / ×3,0) |
| `e = 8` | 146 Mo/s | 515 Mo/s (×3,5) |
| 1 000 lectures de 150 bases, `e = 0` / `3` | 175 / 121 Mo/s | 645 / 400 Mo/s (×3,7 / ×3,3) |
| motif de 44 positions encadré de `N`, `e = 3` | 299 Mo/s | 766 Mo/s |

Le coût du motif encadré est celui du cœur seul. **Le noyau générique à tableau est lent** : avant génération des noyaux déroulés, `e = 4` / `5` / `6` / `8` donnaient ~390 / 355 / 260 / 210 Mo/s (et un premier essai du noyau générique était même plus lent qu'`obiapat` à `e = 5`). Les noyaux déroulés ne montrent pas de palier jusqu'à 8 niveaux (pas de débordement de registres observé) ; au-delà de 8 on retombe sur le générique, non mesuré après coup. Les tests tirent `e` de 0 à 10 pour couvrir les deux.

## Essais écartés (ne pas les refaire sans raison nouvelle)

| Essai | Résultat | Cause |
|---|---|---|
| noyaux sans le masque `allow` pour les motifs sans `#` (une opération de moins par niveau) | **plus lents** : e=1 ~600 contre ~915 Mo/s, e=3 565 contre 770, e=8 420 contre 509 ; même avec des temporaires explicites | vu à l'assembleur (`go tool objdump`, arm64, Go 1.27) : le compilateur ajoute 3 `MOVD` qui échangent deux registres à chaque tour et réassocie les `ORR`, ce qui allonge la chaîne critique. Défaut de génération de code, pas d'algorithme ; pourrait changer avec une autre version de Go |
| noyau générique à tableau pour 4 niveaux et plus | ~390 Mo/s à e=4, ~260 à e=6, ~210 à e=8 (un premier essai : plus lent qu'`obiapat`) | état en mémoire. Remplacé par les noyaux générés (e=1..8) ; le noyau à tableau ne sert plus qu'à e ≥ 9 |
| BNDM pour les motifs courts ou très dégénérés | perte jusqu'à −55 % (motif de 4), −15 % (20 positions dont 12 `N`) | sauts trop courts ; remplacé par le choix à la compilation (`advance()` ≥ 3,6) |

## Suite

1. ~~Voie stricte~~ faite (BNDM). Reste : strict multi-motifs (Wu-Manber avec décalage par q-gramme, ou Aho-Corasick), BNDM sur séquence circulaire, Shift-And à plusieurs flux pour les motifs courts.
2. `PatternSet` : empaquetage de plusieurs motifs (et des deux brins) dans les mots de 64 bits, occurrences étiquetées ; puis SIMD par voies de motifs.
3. Port de `ManberIndel` (avec la logique `obialign.LocatePattern` inchangée), pour `obiannotate` et `obingslibrary`.
4. Migration des appelants (`obipcr`, `obisplit`, `obigrep`, `obiannotate`, `obingslibrary`), un par un ; `obiapat` reste jusqu'à la fin.
5. Question ouverte : `X` dans une séquence (traité comme ni IUPAC ni joker). (La variante sans masque `allow` a été essayée et abandonnée, voir « Essais écartés ».)
