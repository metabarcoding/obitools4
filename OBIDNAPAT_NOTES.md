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

Noyaux : Shift-And de Baeza-Yates/Gonnet étendu par Wu-Manber ; `e = 0` écrit à la main (`run0`), `e = 1` à `8` **générés** (`gen_kernels.go`, `go generate`,
résultat `scan_unrolled.go`) avec un état par variable, donc en registres ; noyau générique à tableau pour `e ≥ 9`. `maxUnrolled` se règle dans le générateur.

## Tests

- `oracle_test.go` : recherche naïve fenêtre par fenêtre, sur séquences avec ambiguïtés et caractères hors alphabet, `e` de 0 à 6, linéaire et circulaire,
  régions, jokers, `#`.
- `diff_test.go` : comparaison avec `obiapat` sur séquences ACGT (là où les deux règles d'ambiguïté ne peuvent pas différer).
- `pattern_test.go` : syntaxe, ensembles de lettres, limite 63 sur le cœur, complément inverse, cas à la main.
- `OBIDNAPAT_ITER=300000 go test ./pkg/obidnapat/ -run TestAgainst` : 300 000 tirages chacun, `e` de 0 à 10, **tous verts**, `-race` vert.
- Seule différence connue avec `obiapat` rencontrée : motif plus long que la séquence circulaire (obiapat lit le cercle plusieurs fois).

## Mesures (M3 Max, un cœur, `go test -bench`, un motif de 20 positions, séquence ACGT aléatoire de 1 Mb ; bruit non évalué)

| Cas | obiapat | obidnapat |
|---|---|---|
| `e = 0` | 1 001 Mo/s | 1 204 Mo/s |
| `e = 1` / `2` / `3` | 330 / 320 / 297 Mo/s | 915 / 838 / 766 Mo/s (×2,8 / ×2,6 / ×2,6) |
| `e = 4` / `5` / `6` | 275 / 238 / 199 Mo/s | 724 / 656 / 596 Mo/s (×2,6 / ×2,8 / ×3,0) |
| `e = 8` | 146 Mo/s | 515 Mo/s (×3,5) |
| 1 000 lectures de 150 bases, `e = 0` / `3` | 168 / 123 Mo/s | 476 / 400 Mo/s (×2,8 / ×3,3) |
| motif de 44 positions encadré de `N`, `e = 3` | 298 Mo/s | 794 Mo/s |

Le coût du motif encadré est celui du cœur seul. **Le noyau générique à tableau est lent** : avant génération des noyaux déroulés, `e = 4` / `5` / `6` / `8` donnaient ~390 / 355 / 260 / 210 Mo/s (et un premier essai du noyau générique était même plus lent qu'`obiapat` à `e = 5`). Les noyaux déroulés ne montrent pas de palier jusqu'à 8 niveaux (pas de débordement de registres observé) ; au-delà de 8 on retombe sur le générique, non mesuré après coup. Les tests tirent `e` de 0 à 10 pour couvrir les deux.

## Suite

1. Voie stricte plus rapide que Shift-And (`e = 0`) : par exemple code 2 bits roulant avec table de hachage ou bitmap des expansions du motif (à mesurer).
2. `PatternSet` : empaquetage de plusieurs motifs (et des deux brins) dans les mots de 64 bits, occurrences étiquetées ; puis SIMD par voies de motifs.
3. Port de `ManberIndel` (avec la logique `obialign.LocatePattern` inchangée), pour `obiannotate` et `obingslibrary`.
4. Migration des appelants (`obipcr`, `obisplit`, `obigrep`, `obiannotate`, `obingslibrary`), un par un ; `obiapat` reste jusqu'à la fin.
5. Question ouverte : `X` dans une séquence (traité comme ni IUPAC ni joker). Piste de gain à mesurer : une variante des noyaux sans le masque `allow` quand le motif n'a aucun `#` (une opération de moins par niveau).
