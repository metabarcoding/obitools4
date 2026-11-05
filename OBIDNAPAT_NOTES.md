# obidnapat — remplaçant Go pur d'obiapat, spécialisé ADN

Paquet `pkg/obidnapat`. Dernière mise à jour : 2026-11-05. Contexte et semantique d'`obiapat` : annexe D de
`OBIPRIMERS_NOTES.md`. Commité au 2026-11-05 (`git status` propre) ; aucun appelant n'a été migré.

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
`MakeApatPattern(..., allowsIndel=true)` : voir « Indels » plus bas.

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

## `PatternSet` : décisions (2026-11-05)

Motivation et consommateur : `ExtractMultiBarcode` (`pkg/obingslibrary/multimatch.go:573`) cherche, pour chaque marqueur, `forward`, `creverse`, `reverse`, `cforward`
(4 × N `AllMatches`), trie les occurrences par `Begin` (`slices.SortFunc`, **non stable**) puis applique un automate à deux états (jeton positif, puis jeton
`-Marker` de même sens). La seconde recherche est aujourd'hui bornée (`locs[0][0]+1`) pour réduire son coût ; avec une recherche conjointe ce coût disparaît et
l'automate peut être réécrit sur des jetons `{Pattern, Strand}`. Le mode indels de `obimultiplex` attend `ManberIndel` ; `Match`/`BestMatch` (dépendances
séquentielles) ne sont pas un usage de `PatternSet`.

- API : `MakePatternSet([]ApatPattern, bothStrands bool)`, `AppendAll[Circular]`, occurrences `{Pattern, Strand, Start, End, Errors}`. Chaque motif et son complément
  inverse sont deux entrées ; pas de dédoublonnage des palindromes.
- **Ordre de sortie : `Start` croissant, puis `(Pattern, Strand)`** (choix de l'utilisateur : ordre naturel pour l'automate ; déterministe). Le scan produit les fins dans l'ordre :
  un court tampon de réordonnancement (borné par l'écart de longueur des motifs) rétablit l'ordre par début. Un rappel en ligne `Scan(seq, func(Occurrence))` en ordre de
  fin pourra s'ajouter plus tard (changement de comportement possible pour amorces de longueurs différentes ou chevauchantes : à décider et mesurer, non fait).
- Limite **63 positions par motif** inchangée en mono comme en multi-motifs. La capacité de 64 bits d'un mot empaqueté est un détail interne. Passer à 64 par motif serait
  une décision séparée ; le bit de garde ne semble pas nécessaire dans le code actuel (hypothèse **non vérifiée**, à tester si on y vient).
- Récurrence empaquetée : `|1` devient `| S` (S = bit de départ de chaque motif du mot) ; le bit qui déborde d'un motif sur le suivant est écrasé par `S`. À démontrer par test
  (test de non-fuite dédié, vérifié par mutation). Masque de fin par mot ; erreurs = plus petit niveau dont le bit de fin est posé, émis si ≤ errmax du motif.
  Niveaux du mot = max des errmax. Fenêtres, flancs `N`, région et règle circulaire filtrés à l'émission par motif ; motif de jokers seuls traité à part.
- Plan : (1) version générique (noyau à tableau, empaquetage glouton, tampon de réordonnancement) + tests ; (2) noyaux déroulés ; (3) mesure contre scans séparés et
  `MatchWords` pour N = 1, 2, 4, 8, 32, 200 ; (4) SIMD par voies seulement si la mesure le justifie.
- Tests : oracle naïf contre `PatternSet` (jeux de 1 à 40 motifs, totaux de 63, 64, 65 positions, errmax par motif de 0 à 10, IUPAC, `#`, flancs, ambiguïtés, linéaire/circulaire,
  régions, un ou deux brins) ; mono-motif contre `FindAllIndex` ; non-fuite ; palindromes ; projection par `(Pattern, Strand)` = scans séparés ; liste globale triée par
  `(Start, Pattern, Strand)`, avec motifs de longueurs différentes dont les fins sont inversées par rapport aux débuts.

### `PatternSet` : version générique faite (patternset.go, patternset_scan.go, patternset_test.go)

Fait : empaquetage, noyau à tableau (`word.scan`), API `AppendAll[Circular]` / `FindAll[Circular]`, `Len`, `Words`. `compiled.span` a été extrait de `search` (réutilisé par les deux).
Écarts par rapport au plan, à savoir :
- **Tri final avec `slices.SortFunc`** sur la partie ajoutée, pas de tampon de réordonnancement (plus simple ; à optimiser seulement si la mesure le montre).
- **Empaquetage par classe de niveaux** (même `levels` ensemble, first fit decreasing par longueur de cœur à l'intérieur) pour qu'un motif exact ne paie pas les niveaux de ses voisins. Capacité d'un mot : 64.
- Au plus **une `Occurrence` par (motif, brin, début)**, avec le plus petit nombre d'erreurs, comme `ApatPattern`.
Tests (tous verts, `-race` vert avec `OBIDNAPAT_ITER=30000`, ~520 s) : `TestPatternSetAgainstOracle` (jeux de 1 à 40 motifs, errmax de 0 à 10, 1 ou 2 brins, linéaire/circulaire, régions),
`TestPatternSetWordBoundary` (totaux 63/64/65, nombre de mots vérifié : 1, 1, 2), `TestPatternSetSingle`, `TestPatternSetNoLeak`, `TestPatternSetOrder`.
**Vérification par mutation** : forcer le bit de départ au seul bit 0 du mot fait échouer 4 tests ; retirer le filtre de plage par cœur en fait échouer 2.
Reste : noyaux déroulés, mesure (N = 1..200 motifs contre scans séparés et `MatchWords`), SIMD si justifié, rappel en ligne éventuel.

### `PatternSet` générique : première mesure (`BenchmarkSet`, bench_set_test.go)

N motifs de 20 positions, deux brins (2N entrées, 3 cœurs par mot), 1 Mb, M3 Max, load ~6, `-benchtime 5x -count 2` (**bruité**, ±30 % sur les petits N à e=0). Mo/s de séquence pour tout le jeu :

| N | e=0 séparés / jeu | e=2 séparés / jeu |
|---|---|---|
| 1 | 540-610 / 340-380 | 350-375 / 265-350 |
| 2 | 350-370 / 208 | 195-200 / 250 |
| 4 | 177-180 / 138 | 101 / 165 |
| 8 | 86 / 70-182 | 49-50 / 84 |
| 32 | 23,5 / 32-52 | 13 / 23,5 |
| 200 | 3,7 / 5-6 | 2,1 / 3,4-3,7 |

Lecture : gain ×1,6-1,7 à e=2 dès N ≥ 4 ; à e=0 le jeu **perd** pour N ≤ 4 (les scans séparés utilisent BNDM, qui saute, alors que le jeu lit chaque octet) et gagne ~×1,4-2 à N ≥ 8 (très bruité). Le gain reste très en deçà du ×3 que donnerait le partage des mots (3 cœurs par mot) : le noyau du jeu est le générique à tableau, comparé à des noyaux déroulés pour les scans séparés (hypothèse sur la cause, non vérifiée par profil). **Non fait : comparaison avec `MatchWords`** (travaille sur un `Corpus`, pas sur une séquence : banc à construire).

### `PatternSet` : noyaux déroulés (2026-11-05)

`gen_kernels.go` produit maintenant aussi `patternset_unrolled.go` : `word.run0..run8` (`| start` à la place de `| 1`, états en registres, `word.emit` appelé sur les seuls coups au but
après recopie de l'état dans `st`) ; `word.run` choisit le noyau déroulé, sinon `runN` (tableau, niveaux ≥ 9). Tests identiques, verts (`-race`, `OBIDNAPAT_ITER=10000`, 177 s).

`BenchmarkSet` après déroulage (Mo/s de séquence pour tout le jeu, deux séries concordantes à 1 % près, load ~5) :

| N | e=0 séparés → jeu | e=2 séparés → jeu |
|---|---|---|
| 1 | 670-725 → 1 115-1 155 | 419 → 847 |
| 2 | 350 → 620 | 210 → 422 |
| 4 | 171 → 408 | 105 → 279 |
| 8 | 85 → 203 | 52 → 139 |
| 32 | 22,6 → 56 | 13 → 37,8 |
| 200 | 3,6 → 9,1 | 2,1 → 6,2 |

Gain ×1,6 (N=1) à ×2,4-2,7 (N ≥ 4) sur les scans séparés, y compris à e=0 où le jeu bat désormais BNDM (motifs de 20, 3 cœurs par mot). Le noyau générique à tableau était donc bien la cause
du gain faible (cohérent avec l'hypothèse, non vérifiée au profil). Reste : comparaison avec `MatchWords`, voie stricte multi-motifs, SIMD par voies (non justifié tant que `MatchWords` n'est pas mesuré).

### `PatternSet` : motifs longs (`BenchmarkSetLen`, 8 motifs, deux brins, 1 Mb, Mo/s, deux séries concordantes)

| longueur | e=0 séparés → jeu (mots) | e=2 séparés → jeu (mots) |
|---|---|---|
| 10 | 77 → 412 (3) | 50 → 64 (3) |
| 20 | 89 → 212 (6) | 52 → 140 (6) |
| 32 | 148 → 160 (8) | 52,5 → 107 (8) |
| 33 | 148 → 77 (16) | 52,5 → 52 (16) |
| 40 | 176 → 80 (16) | 52,5 → 53,6 (16) |
| 63 | 288 → 80 (16) | 53 → 53 (16) |

Au-delà de 32 positions, deux cœurs ne tiennent plus dans un mot : un cœur par mot, donc aucun partage. À e ≥ 1 le jeu égale les scans séparés ; **à e=0 il perd d'un facteur 2 à 3,6** parce que les scans séparés
utilisent BNDM (qui saute) et que le jeu relit tout (≈ 80 Mo/s pour 16 mots). À longueur 10, e=2, le gain n'est que ×1,3 malgré 6 cœurs par mot (hypothèse non vérifiée : coups au but très fréquents, donc coût d'`emit`).
Correctif envisagé, non fait : un cœur seul à e=0 pour lequel BNDM est choisi reste un scan autonome (liste « solo », comme les motifs de jokers).

### Direction SIMD du `PatternSet` (discussion du 2026-11-05, rien d'écrit)

- Parallélisme **sur les mots** (une voie = un mot de 64 bits, éventuellement plusieurs petits cœurs empaquetés), pas sur les motifs : la limite « somme des longueurs ≤ 64 » disparaît pour les motifs longs (un motif de 63 = une voie).
  Table de masques en `[octet][voie]` (une charge de vecteur par octet, pas de `gather`). Vérifié : `simd.Uint64s` (Go 1.27.3, arm64) a `ShiftAllLeft`, `Or`, `And`, `AndNot`, `Equal`, `Len` ; vecteur ≥ 128 bits donc **2 voies sur M3** (gain borné à ×2, non mesuré), 4 à 8 sur x86.
- Points ouverts : détection des fins/`emit` scalaire ; pression sur les 32 registres vectoriels à 8-9 niveaux ; BNDM reste probablement meilleur que tout balayage à e=0 pour un motif seul (288 Mo/s contre 80 pour 16 mots) → liste « solo » ou voie stricte multi-motifs.
- **Optimisation du remplissage (à faire après le noyau SIMD, voulue par l'utilisateur)** : bin packing à critères multiples : capacité 64 bits, même `levels` dans un vecteur (toutes les voies font le maximum), nombre de mots multiple du nombre de voies (dernier vecteur plein), équilibre des coups au but (motifs courts/dégénérés/`e` élevé à part : hypothèse).
- Prototype proposé : noyau SIMD à `e = 2` seulement, repli scalaire existant, mesuré contre `run2` avec `BenchmarkSetLen` avant de généraliser.

### `PatternSet` : liste « solo » et prototype SIMD à e = 2 (2026-11-05)

- **Liste « solo »** (faite) : après l'empaquetage, un mot à un seul cœur, e = 0, avec BNDM choisi, redevient un scan autonome (`PatternSet.solo`, avec les motifs de jokers). `BenchmarkSetLen`, e = 0 : longueur 33 / 40 / 63 : séparés 140 / 183 / 277 Mo/s, jeu 144 / 184 / 284 Mo/s (la perte ×2-3,6 a disparu). `Words()` ne compte pas les solos.
- **Prototype SIMD** (`patternset_vec.go`, `patternset_simd.go` sous `goexperiment.simd`, `patternset_nosimd.go`) : groupes de `lanes` mots à 2 niveaux, un mot par voie ; le noyau vectoriel lit des blocs de 64 octets, accumule les bits de fin, et seuls les mots dont une voie a vu une fin sont relus par le noyau **scalaire** `word.run` depuis l'état sauvegardé en début de bloc : les occurrences sont donc exactement les scalaires. Les mots qui ne remplissent pas un groupe restent scalaires. `makePatternSet(..., vector bool)` permet de comparer.
- Tests : `TestPatternSetVector` (e = 2, motifs de 4 à 63), toute la suite verte avec et sans `GOEXPERIMENT=simd` (2 voies, 8 783 groupes sur le test) ; `-race` en simd (`OBIDNAPAT_ITER=6000`) ; mutation (suppression des transitions d'erreur du noyau vectoriel) : 2 tests échouent.
- **Bogue du compilateur** (Go 1.27.3, simd) : `internal compiler error: missing Types entry: block@simd0` quand le code vectoriel est dans une **méthode** ; une fonction ordinaire à arguments simples (`vectorBlockKernel`) compile.
- **`ShiftAllLeft(1)` compile mal sur arm64** (vu à `go tool objdump`) : 3 `VMOV R, V.D[0]` + `VUSHL` par octet (compte de décalage variable, pas de `SHL #1`) ; avec `x.Add(x)` à la place, la boucle n'a plus que des `VADD/VORR/VAND` (≈ 14 opérations vectorielles par octet pour 2 voies). Gain de 0,71× à 1,06× du scalaire.
- **Résultat final sur M3 (2 voies)** (`BenchmarkSetVector`, e = 2, 8 ou 64 motifs, deux brins, Mo/s scalaire → vecteur) : longueur 20 : 138 → 145 et 19,3 → 20,2 ; 33 : 52 → 55 et 6,5 → 6,8 ; 63 : 52 → 55 et 6,5 → 6,8 ; **longueur 10 : 58 → 51 et 3,0 → 2,8 (perte)**. Soit ≈ +5 % pour les motifs de 20 et plus, −13 % quand les fins sont fréquentes (relectures scalaires des blocs).
  Cause plausible (hypothèse, non vérifiée) : 2 voies sur 4 pipelines NEON contre 6 ALU entières, et le scalaire fusionne le décalage dans l'`ORR`. L'intérêt attendu du SIMD (voies plus larges) est à mesurer sur x86 AVX2/AVX-512 (4-8 voies), non disponible ici.
- Décision à prendre : garder le prototype (désactivé hors `GOEXPERIMENT=simd`), le généraliser aux autres niveaux, ou le retirer tant qu'aucune machine à voies larges ne le justifie.

### Entrelacement de mots scalaires (expérience, interleave_exp.go + interleave_exp_test.go, 2026-11-05)

Idée : la boucle scalaire d'un mot est limitée par la chaîne de dépendance d'un octet à l'autre ; faire avancer 2 ou 3 mots **indépendants** dans la même boucle donne au processeur de quoi recouvrir les latences. Noyaux écrits à la main pour e = 2 seulement (`run2x2`, `run2x3`), non branchés sur `PatternSet`.
`TestInterleaveSame` : mêmes occurrences que `appendAll` (longueurs 8, 20, 63 ; k = 1, 2, 3). `BenchmarkInterleave` (e = 2, 1 Mb, deux séries concordantes, Mo/s, k = mots par boucle) :

| longueur (mots) | k=1 | k=2 | k=3 |
|---|---|---|---|
| 10 (4) | 28,2 | 32,9 (×1,17) | 30,2 (×1,07) |
| 20 (8) | 106 | 157 (×1,48) | 168 (×1,58) |
| 63 (24) | 35,2 | 52,9 (×1,50) | 57,6 (×1,64) |

Gain ×1,5 à ×1,65 pour les motifs de 20 et plus, contre +5 % pour le SIMD à 2 voies : **le goulot du noyau scalaire est bien la chaîne de dépendance** (confirmé par la mesure, pas par le profil). k=3 n'ajoute que 5-10 % à k=2 : le débit des ALU commence à plafonner. Gain faible à longueur 10 (fins fréquentes, donc `emit`).
Suite possible : générer les noyaux entrelacés pour les niveaux 0 à 8 (k = 2 ou 3, restes en k plus petit) et les brancher sur `PatternSet`, groupés par niveaux ; réévaluer ensuite le SIMD (qui pourrait entrelacer aussi plusieurs vecteurs) ; mesurer sur x86.

### `PatternSet` : noyaux entrelacés branchés (2026-11-05)

`gen_kernels.go` génère `patternset_il*.go` : `run<L>x<K>` (L = 0..8 niveaux, K = 2 ou 3 mots de même niveau avancés dans la même boucle, un état par variable) et `runWords` (répartition). `PatternSet.batches` regroupe les mots par lots de même niveau (`makeBatches`, `batchFor(levels)`) ; `scanBatch` remplace `word.scan` (supprimé) ; `interleave_exp*.go` (expérience) supprimés. Le SIMD (`vgroup`) reste **optionnel** (`makePatternSet(..., vector=true)`, build `goexperiment.simd`), **désactivé par défaut** : l'entrelacement scalaire gagne bien plus.
Taille de lot (`BenchmarkBatch`, 8 motifs, deux brins, k mots par boucle, rapport à k=1) :

| e | len 20 : k=2 / k=3 | len 63 : k=2 / k=3 |
|---|---|---|
| 1 | ×1,97 / ×2,27 | ×1,82 / ×1,98 |
| 2 | ×1,52 / ×1,65 | ×1,53 / ×1,55 |
| 3 | ×1,21 / ×1,40 | ×1,42 / ×1,42 |
| 4 | ×1,28 / ×1,15 | ×1,32 / ×1,15 |
| 6 | ×1,05 / ×0,93 | ×1,10 / ×0,87 |
| 8 | ×1,09 / ×1,09 | ×0,94 / ×0,64 |

Réglage : `batchFor` = 3 mots jusqu'à 3 niveaux, 2 à 4-5 niveaux (5 non mesuré), 1 à partir de 6. Les cas à très nombreuses occurrences (motifs de 10, e ≥ 2 : des millions d'occurrences) mesurent surtout l'émission et le tri final ; ils ne sont pas comparables (rapports erratiques, non expliqués).
`BenchmarkSet` après entrelacement (Mo/s de séquence, jeu / scans séparés) : e = 0 : N = 1 ×1,6 ; N = 2 ×3,4 ; N ≥ 4 ×6 ; e = 2 : N = 1 ×2,0 ; N = 2 ×3,0 ; N ≥ 4 ×4,4 à ×4,9. `BenchmarkSetLen` (8 motifs, deux brins) : e = 0 : longueur 10 / 20 / 32 : ×12,8 / ×6,1 / ×2,6, 33 et plus ×1,0 (solo BNDM) ; e = 2 : 10 / 20 / 32 : ×4,0 / ×3,8 / ×3,2 ; 33 à 63 : ×1,6 (un mot par cœur, seul l'entrelacement joue).
Tests : `TestPatternSetBatches` (e = 0..10, tailles de lot 1 à 3 imposées, flancs de `N` pour que les fenêtres des mots d'un lot diffèrent), suite complète verte, `-race` vert (`OBIDNAPAT_ITER=6000`, 233 s ; en build simd `OBIDNAPAT_ITER=3000`, 117 s). **Mutations** (`fs[1]` → `fs[0]` dans `run0x2`, `run2x2`, `run2x3`) : détectées à chaque fois. (Un premier essai de mutation n'avait rien modifié : `sed` BSD n'accepte pas l'adresse `0,/re/`.)
Reste : voie stricte multi-motifs (e = 0, motifs ≥ 33), comparaison avec `MatchWords`, mesure sur x86, réévaluation du SIMD avec entrelacement de vecteurs, optimisation du remplissage (bin packing), BNDM circulaire, `ManberIndel`, migration des appelants.

### Indels : port de `ManberIndel` (2026-11-05, scan_indel.go, scan_indel_unrolled.go, indel_test.go)

`MakeApatPattern(..., allowsIndel=true)` n'est plus refusé. `compiled.indel = allowsIndel && errormax > 0` (sans erreur permise, c'est la voie stricte, comme `ManberAll`) ; `allowsIndel` est gardé pour `ReverseComplement` et `BestMatch`.
Récurrence (Wu-Manber, même convention de bits que `scan.go`) : `n_0 = ((o_0<<1)|1) & mask[b]` ; `n_j = (((o_j<<1)|1) & mask[b]) | ((o_{j-1} | (o_{j-1}<<1) | 1 | (n_{j-1}<<1)) & allow)` (insertion, substitution, suppression, toutes soumises à `#`) ;
état initial `o_j = 2^j - 1` (j suppressions de tête, non soumises à `#`, comme obiapat). Noyau à tableau (`runIndelN`) pour 9 niveaux et plus, noyaux déroulés générés pour 1 à 8. Le terme `| (j > 1)` de l'écriture d'obiapat est redondant avec la substitution (la mutation qui le retirait ne faisait échouer aucun test) : supprimé.
Décisions / écarts (**à valider par l'utilisateur**) :
- **Fenêtre nominale** `[pos-len+1, pos+1)` comme obiapat ("may return shifted pos") ; `BestMatch` la réaligne par `obialign.LocatePattern` (logique inchangée : fragment `[start-nerr, start+len+nerr)`, `from < 0` = pas de match, comme l'`obiapat` actuel).
- **Région stricte** : la fenêtre nominale doit tenir dans la région (obiapat laissait passer des fenêtres de début négatif et dépassait de 64 octets). Circulaire : le début est dans la région, comme en substitution. Conséquence : `BestMatch` ne peut plus choisir un meilleur candidat de début négatif puis le rejeter.
- **Pas de retrait des jokers `N` de tête et de queue** avec indels (la sémantique d'obiapat les traite comme des positions) : la limite de 63 compte alors le motif entier. Sans erreur permise, retrait comme avant.
- **Le nombre d'erreurs peut dépasser la longueur du motif** (insertions) : `levels = min(errormax, 63)`, pas `min(errormax, longueur)` (premier essai faux, trouvé par les deux tests).
- `PatternSet` ne porte **pas** les indels (non fait).
Tests : `TestIndelAgainstOracle` (programmation dynamique de Sellers, indépendante, linéaire/circulaire, régions, `#`, ambiguïtés, e de 1 à 10), `TestIndelAgainstObiapat` (fenêtre par fenêtre, début nominal compris), `TestIndelBestMatchAgainstObiapat` (2 213 cas comparés, 533 réalignés) ; verts, `-race` vert (`OBIDNAPAT_ITER=6000` suite complète 236 s ; `Indel` seul à 60 000 : 26 s). Mutations détectées (suppression des insertions, mauvais niveau dans la suppression de `runIndel3`).
Mesure (`BenchmarkIndel`, motif de 20, 1 Mb, Mo/s, `ManberIndel` contre ce paquet) : e = 1 : 300 → 880 (×2,9) ; e = 2 : 317 → 685 (×2,2) ; e = 3 : 272 → 510 (×1,9) ; e = 5 : 187 → 328 (×1,75). Noyau à tableau seul : ×1,1 à ×1,3.
Reste : `PatternSet` avec indels ; migration d'`obiannotate`, `obingslibrary`.

## Suite

1. ~~Voie stricte~~ faite (BNDM). Reste : strict multi-motifs (Wu-Manber avec décalage par q-gramme, ou Aho-Corasick), BNDM sur séquence circulaire, Shift-And à plusieurs flux pour les motifs courts.
2. `PatternSet` : voir la section ci-dessus (en cours, version générique d'abord) ; puis SIMD par voies de motifs si la mesure le justifie.
3. Port de `ManberIndel` (avec la logique `obialign.LocatePattern` inchangée), pour `obiannotate` et `obingslibrary`.
4. Migration des appelants (`obipcr`, `obisplit`, `obigrep`, `obiannotate`, `obingslibrary`), un par un ; `obiapat` reste jusqu'à la fin.
5. Question ouverte : `X` dans une séquence (traité comme ni IUPAC ni joker). (La variante sans masque `allow` a été essayée et abandonnée, voir « Essais écartés ».)
