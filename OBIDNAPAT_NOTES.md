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

### Vrai début sans programmation dynamique : recherche arrière à bits (2026-11-05, scan_backward.go)

Idée de l'utilisateur : le balayage avant donne la **fin exacte** (`pos+1`) et le nombre d'erreurs, pas le début. `backwardStart(seq, pos, e, lo)` relit les ≤ len+e bases qui précèdent la fin avec le même automate à bits, motif lu à l'envers (bit j du niveau k : le motif de la position j à la fin est aligné avec les bases lues, ≤ k erreurs ; insertion entre j-1 et j soumise à `allow << 1`, rien n'est inséré avant la première position), et rend le début **le plus tardif** (fenêtre la plus courte) dont l'alignement de tout le motif coûte ≤ e. `BestMatchBackward` l'utilise à la place de `LocatePattern` ; `BestMatch` est inchangé.
- **Exactitude** : `TestBackwardStartAgainstTruth` compare à une programmation dynamique globale par début candidat (`truthStart`, indépendante) sur 125 993 fenêtres avec erreurs : identique ; sans `#`, les deux s'accordent aussi sur l'existence d'un début. Avec `#` dans les premières positions le balayage avant autorise des suppressions de tête hors `#` (comme obiapat) et la recherche arrière non : elle peut ne rien trouver (`ok` faux) là où l'avant a signalé une fenêtre.
- **Coût** (`BenchmarkRefine`, 1 000 lectures de 150 bases, motif de 20 planté avec ≤ 2 indels, e = 3, par lecture) : balayage avant seul 0,35 µs ; avec `LocatePattern` 1,7 à 2,1 µs (réalignement ≈ 1,3 à 1,8 µs) ; avec la recherche arrière 0,46 µs (≈ 0,11 µs). Réalignement ≈ 12 à 16 fois moins cher, `BestMatch` complet ≈ 4 à 5 fois plus rapide.
- **Différences avec `LocatePattern`** (`TestBackwardAgainstLocatePattern`, motifs en lettres IUPAC simples, ~4 000 cas) : 72 % identiques ; sinon départage entre alignements de **même coût** : `LocatePattern` prend toujours la fenêtre la plus longue (début plus tôt dans 338 cas sur 338 à fin égale ; fin plus tard dans 762 cas sur 774 où la fin diffère), mais pas le début le plus précoce non plus (mon essai « début le plus précoce » le dépassait d'une base dans 449 cas sur 452) : son choix est un artefact de l'ordre de préférence de son retour en arrière. Dans 15 cas il annonce moins d'erreurs (son modèle de coût diffère aux extrémités). Non asserté.
- **Bogue trouvé hors du paquet** : `pkg/obiseq/iupac_nog.go`, table `_iupac` : la ligne de `v` vaut 13 (A,G,T = le code D) au lieu de 7 (A,C,G). `SameIUPACNuc('V', ·)` accepte donc `agt` au lieu de `acg` ; cela fausse `LocatePattern` pour tout motif contenant V. **Corrigé par l'utilisateur** (ligne de `v` : 7) ; la comparaison avec `LocatePattern` refaite avec V : 2 960 identiques sur 4 030, même départage. Autre limite : `LocatePattern` lit le **texte brut** du motif tronqué à sa longueur, donc faux avec `!`, `[..]`, `#`.
- Départage : l'utilisateur confirme qu'il n'y a **pas de bonne réponse** (plusieurs alignements de même coût ; "la plus longue" n'est pas démontrée meilleure) et que le « vrai » début qui a produit le succès n'est pas défini de façon unique. Piste proposée, non faite : rendre l'**intervalle** des débuts (et des fins) de coût minimal, et laisser l'appelant choisir (pour rogner une amorce, l'union est le choix sûr). À décider : garder `LocatePattern` ou le remplacer.

### Fraction de débuts ambigus (simulation, ambiguity_test.go, 2026-11-05)

`TestAmbiguityOfTheStart` : 3 000 lectures de 150 bases par modèle, amorce aléatoire de 20 bases plantée (substitutions, suppressions, insertions par base), e permis = 2, indels permis. Ambigu = plusieurs fenêtres de coût minimal (début : intervalle `[plus tôt, plus tard]` par oracle ; fin : autre occurrence de même coût à ±3 bases). Le début de `BestMatchBackward` est vérifié égal au « plus tard » de l'oracle à chaque lecture.

| Modèle | amorce avec erreurs | début ambigu (de toutes les amorces trouvées / de celles avec erreurs) | fin ambiguë | l'un ou l'autre |
|---|---|---|---|---|
| substitutions 1 % | 12,6 % | 1,0 % / 7,9 % | 0,7 % | 1,7 % |
| subst. 1 %, indels 0,2 % | 19,5 % | 1,8 % / 9,2 % | 1,4 % | 3,2 % |
| synthèse d'oligos (sub 0,5 %, suppr. 1 %, ins. 0,2 %) | 26,0 % | 2,3 % / 8,7 % | 2,8 % | 5,0 % |
| dégradé (sub 3 %, indels 1 %) | 55,1 % | 4,8 % / 8,7 % | 5,3 % | 9,9 % |

Largeur moyenne de l'intervalle des débuts : 1,2 à 1,5 base. Lecture : l'ambiguïté ne vient pas que des indels (une **substitution sur une base de bord** donne déjà deux alignements de même coût quand les indels sont permis) ; ≈ 8-9 % des amorces avec erreurs, 1 à 5 % de toutes. **Limites** : lectures simulées, amorces aléatoires (une amorce dont le bord est un homopolymère ou un motif répété serait plus ambiguë), taux d'erreurs supposés ; à refaire sur de vraies lectures.

### Fraction de débuts ambigus sur de vraies lectures MinION (realdata_test.go, 2026-11-05)

Données : `/Users/coissac/travail/MinIon/Metabarcoding/nanopore_test_plate_A.fastq.gz` (lectures de ~160 bases en médiane, 28 à 334 sur les 1 000 premières) avec `plate1_pool-SPER05.Pierre.ngsfilter` (amorces `GGGCAATCCTGAGCCAAATC` / `CTTCCRTTGAGTCTCTGCACCT`, tags de 7 bases, délimiteur `G`, indels). Lancer : `OBIDNAPAT_FASTQ=<fichier> OBIDNAPAT_PRIMERS=<fwd>,<rev> OBIDNAPAT_READS=20000 go test -v -run AmbiguityOnRealReads ./pkg/obidnapat/` (ignoré sans la variable).
20 000 premières lectures, quatre motifs d'`obimultiplex`, e = 3 (indels permis), part des amorces trouvées dont le début/la fin est ambigu (plusieurs fenêtres de coût minimal ; début vérifié par l'oracle) :

| motif | trouvé | avec erreurs | début ambigu (de celles avec erreurs) | fin ambiguë | l'un ou l'autre |
|---|---|---|---|---|---|
| directe | 61 % | 18,5 % | 2,1 % (11 %) | 3,6 % | 5,5 % |
| compl. inverse de l'inverse | 55 % | 29,4 % | 3,9 % (13 %) | 2,3 % | 6,1 % |
| inverse | 51 % | 27,8 % | 1,5 % (5 %) | 2,1 % | 3,6 % |
| compl. inverse de la directe | 51 % | 26,5 % | 2,7 % (10 %) | 6,4 % | 8,9 % |

Largeur moyenne de l'intervalle des débuts : 1,4 à 1,9 base. e = 2 : début ambigu 1,2 à 2,8 %, l'un ou l'autre 2,8 à 5,7 % ; e = 4 : début 1,9 à 4,9 %, l'un ou l'autre 4,4 à 11,6 %. Rôles dans `obimultiplex` (e = 3) : bord côté tag = début de la directe (2,1 %) et fin du compl. inverse de l'inverse (2,3 %) ; bord côté code-barres = fin de la directe (3,6 %) et début du compl. inverse de l'inverse (3,9 %). Non mesuré : combien de ces lectures changent d'échantillon selon la règle de départage. Réserve : les taux « trouvé » sommés pour une amorce et son complément inverse dépassent 100 % (58,8 % + 47,6 % à e = 2), soit des lectures avec les deux orientations (chimères, concatémères) ou des occurrences approchées sans rapport avec une amorce ; non examiné.

### Toutes les occurrences, début nominal contre vrai début (`TestAllOccurrencesOnRealReads`, mêmes données, e = 3, 2026-11-05)

**CORRECTION (2026-11-05)** : la version d'abord écrite ici affirmait que le mode multi-code-barres d'`obimultiplex` prend des fenêtres nominales. C'est **faux** : l'`AllMatches` d'`obiapat` (pkg/obiapat/pattern.go) regroupe les fenêtres (`FilterBestMatch`) puis **réaligne chacune avec `obialign.LocatePattern`** (fragment `[début-2e, début+len+4e)`), recalcule début, fin et erreurs, et supprime la fenêtre si le réalignement échoue ou dépasse `maxerr`. L'erreur venait de ce que, dans ce paquet, `AllMatches` déléguait alors à `FilterBestMatch`. Les statistiques ci-dessous « début nominal hors de l'intervalle » (55-67 % des fenêtres avec erreurs) décrivent donc le **`FilterBestMatch` nominal**, pas ce qu'obimultiplex reçoit ; elles restent vraies pour ce qu'elles mesurent. La comparaison pertinente est dans la section suivante. `lookForTag` (tag délimité) part de `begin`, saute les non-délimiteurs puis la série de délimiteurs, donc tolère un petit décalage de `begin` mais pas un décalage à gauche qui entre dans le délimiteur. Une lecture porte plusieurs amplicons bout à bout.
Résultats (20 000 lectures) :
- Fenêtres par lecture, motif direct : 0 : 7 760, 1 : 8 421, 2 : 2 503, 3 : 903, 4 : 285, 5+ : 128 (≈ 19 % des lectures ont 2 fenêtres ou plus).
- **Début nominal hors de l'intervalle du vrai début**, parmi les fenêtres avec erreurs : directe 54,6 %, compl. inverse de l'inverse 66,7 %, inverse 63,8 %, compl. inverse de la directe 62,6 % (soit 12 à 22 % de toutes les fenêtres), à 1 base 1 651 / 2 438 / 1 929 / 1 724, à 2 bases 394 / 867 / 666 / 581, à 3 bases 76 / 261 / 151 / 182.
- Début ambigu (intervalle de plus d'une valeur) : 457 / 655 / 262 / 406 fenêtres. L'indice « GG juste avant le début » ne retient **qu'un** candidat dans 138 / 224 / 112 / 53 de ces cas (30 %, 34 %, 43 %, 13 %), aucun dans 267 / 422 / 144 / 334 (les erreurs de nanopore touchent aussi le contexte), plusieurs dans 52 / 9 / 6 / 19 (l'amorce directe commence par GGG, une série de 5 G avec le délimiteur).
Conclusion **retirée** (voir la correction ci-dessus) : le début nominal n'est pas ce qu'obimultiplex reçoit.

### Position du tag comme ancre du début (`TestTagAnchoredStartOnRealReads`, mêmes données, e = 3, 2026-11-05)

Idée de l'utilisateur : ne pas s'appuyer sur le `GG` collé à l'amorce (la série de G se confond avec le `GGG` de l'amorce directe) mais repérer le **tag** : segment sans G de ~7 bases, bordé d'un G à gauche et du délimiteur à droite ; le début de l'amorce est alors `fin du tag + 2`. `tagAnchoredStart(r, s, delim, delimLen, tagLen)` : à gauche de `s`, saute la série de G, lit le segment sans G (longueur tagLen ± 1) borné à gauche par un G. Paramètres `OBIDNAPAT_TAGLEN` (7), `OBIDNAPAT_DELIM` (G), `OBIDNAPAT_DELIMLEN` (2). Amorces dont le tag est à gauche : directe, et inverse (lectures inverses).
- **Validation du modèle sur les fenêtres au début non ambigu** : tag repéré pour 88 % (directe) et 90 % (inverse) ; début prédit = vrai début dans 93 % et 96 % de ces cas ; écarts +1 : 822 et 381, +2 : 149 et 39, −1 : 71 et 85 (longueur de la série de G altérée par les indels de nanopore).
- **Fenêtres au début ambigu** : directe 457, tag repéré pour 226 (49 %), prédiction dans l'intervalle pour 111 (24 % des ambiguës : égale au plus tôt 30, au plus tard 78, entre 3) ; inverse 262, tag repéré pour 34, prédiction dans l'intervalle pour 10. L'ancre ne règle donc qu'environ le quart des débuts ambigus de la directe et presque rien pour l'inverse (les erreurs qui rendent le début ambigu sont proches du bord gauche, donc souvent aussi du tag). Quand elle règle, elle retient le début le plus tard 70 % du temps (directe).
- Non mesuré : côté droit (compl. inverses, tag à droite, délimiteur C), effet sur l'attribution d'échantillon. La logique du tag relève d'`obingslibrary`, pas d'`obidnapat`.

### `AllMatches` à début réel et comparaison avec `obiapat` sur de vraies lectures (2026-11-05)

`ApatPattern.AllMatches` pour un motif à indels (`search.go`) : mêmes regroupements de fenêtres qui se chevauchent que `FilterBestMatch` (`bestOfOverlaps`, extrait), puis, pour chaque fenêtre avec erreurs, le début est remplacé par le **vrai début le plus tardif** (`backwardStart`, la fin et les erreurs restent ceux du balayage avant). La région s'applique au **vrai** début (`indelHits` renvoie les fenêtres sans filtre sur le début nominal, qui peut précéder la région pour une fenêtre avec suppressions). Hors indels, `AllMatches` = `FilterBestMatch`. Limites documentées : circulaire (début nominal), fenêtre dont le vrai début est introuvable (seulement `#` en tête de motif : début nominal conservé). `BestMatchBackward` utilise aussi `indelHits`. Pas de logique de tag ici (décision de l'utilisateur : elle dépend des bibliothèques et reste hors de ce paquet).
Tests : `TestIndelAllMatchesAgainstOracle` (référence indépendante : programmation dynamique non filtrée, regroupement réécrit, vrai début par `truthStart` ; plus de 300 fenêtres dont le début diffère du nominal) ; `TestIndelAllMatchesAgainstObiapat` (fins et erreurs identiques à `FilterBestMatch` d'`obiapat`). Le motif d'**une seule base** fait échouer `LocatePattern` (`end == -1`), donc l'`AllMatches` d'`obiapat` supprime ces fenêtres.
`TestAllMatchesAgainstObiapatOnRealReads` (20 000 lectures MinION, e = 3, les quatre motifs) : **même nombre de fenêtres** (17 984 / 13 796 / 13 937 / 16 093), aucune fenêtre propre à l'un ou à l'autre ; identiques : 95,9 % (directe), 91,2 % (compl. inverse de la directe), 96,4 % (inverse), 94,7 % (compl. inverse de l'inverse) ; même fin mais début différent : 5 / 251 / 176 / 441, **toujours** le début d'`obiapat` plus tôt et **dans** l'intervalle de coût minimal ; fin différente : 734 / 957 / 321 / 410, et dans **100 %** de ces cas la fenêtre d'`obiapat` a **exactement le même coût** que la mienne (modèle de coût de ce paquet) avec une fin **plus tardive** : ce sont des départages d'égalités (`LocatePattern` étend), pas des erreurs.
Coût (`BenchmarkRefine`, 1 000 lectures de 150 bases, e = 3, par lecture) : `AllMatches` d'`obiapat` 3,7 à 3,9 µs (cgo + `LocatePattern`) ; ici 0,68 µs (×5,6).

### Migration des appelants : `obiannotate` faite (2026-11-05)

`pkg/obitools/obiannotate/obiannotate.go` (`MatchPatternWorker`) : `obiapat` remplacé par `obidnapat` (`MakeApatPattern`, `BestMatch(s, 0, s.Len())` sur la `BioSequence` directement, plus de `MakeApatSequence`). Comportement choisi : **`BestMatch` (réalignement par `LocatePattern`, comme avant)**, pas `BestMatchBackward`, pour ne pas changer les annotations (le départage de la recherche arrière diffère de `LocatePattern` dans ~28 % des cas avec erreurs).
Vérification : ancien et nouveau binaires sur 10 000 lectures MinION (`nanopore_test_plate_A.fastq.gz`, premières lectures), `--pattern` avec les deux amorces (la seconde contient un R), `--pattern-error` 0, 1, 3 et 5, avec et sans `--allows-indels` : **sorties identiques octet pour octet** dans les 14 configurations. Temps sur 100 000 lectures (e = 3, indels, un motif) : 0,20-0,23 s avant, 0,15-0,17 s après (bruité). Un motif de plus de 63 positions est maintenant refusé par un message d'erreur clair (`log.Fatalf`, essayé avec 68 bases) alors que l'ancien binaire l'acceptait sans erreur (son résultat n'a pas été examiné). `go build ./...` passe ; le paquet n'a pas de tests.
Reste à migrer : `pkg/obitools/obipcr` + `pkg/obiapat/pcr.go` (circulaire), `obisplit`, `obigrep` (`predicat.go`), `pkg/obingslibrary` (`marker.go`, `multimatch.go`). `obiapat` reste en place.

### Migration d'`obingslibrary` / `obimultiplex` (2026-11-05)

`pkg/obingslibrary/marker.go` (`Marker`, `Compile`, `Compile2`, `Match`) et `multimatch.go` (`ExtractMultiBarcode`) utilisent `obidnapat` ; plus de `MakeApatSequence`. Le mode multi appelle **`AllMatchesRealigned`** (nouveau, `search.go`) = l'`AllMatches` d'`obiapat` porté tel quel (regroupement des chevauchements puis réalignement par `LocatePattern` sur `[début-2e, début+len+4e)`, suppression si échec ou score > `maxerr`, fenêtres sans filtre sur le début nominal comme obiapat).
- **Vérification de bout en bout** (ancien et nouveau `obimultiplex`, `.ngsfilter` figé en copie car l'utilisateur le modifiait : `@param,primer_mismatches,5`, indels, délimiteur G ; 100 000 lectures MinION) : `-e` 0, 1, 3, 4, 6 et par défaut, avec et sans `--with-indels` : **0 enregistrement différent**, attribués (de 63 155 à 100 788) comme non identifiés (de 76 367 à 101 439). Aussi identique sur les 20 000 premières lectures avant changement de fichier. Test permanent `TestAllMatchesRealignedAgainstObiapat` : fenêtres identiques à celles d'`AllMatches` d'`obiapat` (30 000 tirages, avec et sans indels, début de région variable).
- **Avec `AllMatches` à début réel** (essayé avant d'ajouter `AllMatchesRealigned`, `-e` par défaut du moment, 20 000 lectures, sorties comparées lecture par lecture) : 95,4 % des lectures identiques ; 4,1 % (371) même échantillon et mêmes tags mais frontière du code-barres décalée (début du code-barres 1 à 4 bases plus tôt, fin 1 base plus tard : la fenêtre la plus courte donne un code-barres plus long ; **on ne sait pas** si les bases en plus appartiennent à l'amorce ou au code-barres, les deux fenêtres ont exactement le même coût ; non vérifié, une première formulation affirmait à tort qu'il s'agissait de bases d'amorce) ; environ 25 barcodes de plus (fenêtres à 2 erreurs que l'`AllMatches` d'obiapat supprime quand le score de `LocatePattern` dépasse `maxerr`, avec des paires de tags qui existent dans la bibliothèque) et 3 de moins (amorce incomplète par une rafale d'indels : la fin de la fenêtre commande `lookForTag`, 292 fonctionne, 290 non) ; aucun changement d'échantillon sur une lecture qui garde le même nombre de barcodes sauf 1 lecture à échantillon différent. Vitesse : 0,16 s contre 0,18 s sur 20 000 lectures (0,57 contre 0,71 s sur un cœur) ; la mesure de 0,85 s relevée au premier lancement était du bruit.
- **Décision en attente (utilisateur)** : garder `AllMatchesRealigned` (comportement identique, `LocatePattern` conservé) ou passer à `AllMatches` (début réel, ~5 fois moins cher, ~4 % de frontières de code-barres décalées de 1 à 4 bases, quelques barcodes de plus ou de moins). Le choix de la frontière côté code-barres (fenêtre de l'amorce la plus courte ou la plus longue) est la vraie question ; voir « Départage ». Test proposé, non fait : pour les lectures à amorce exacte d'un échantillon, le début du code-barres est certain ; comparer les bases litigieuses des lectures à frontière décalée à ce consensus.
- Le chemin « un seul code-barres » (`Marker.Match`, `ExtractBarcode`, `ExtractBarcodeSliceWorker`) n'a **aucun appelant** hors du paquet (seul `obimultiplex` est utilisateur, et il passe par le mode multi) : migré mécaniquement, non exercé, candidat à la suppression (à décider). Il utilise `BestMatch` avec une région stricte (le début nominal doit être dans la région), alors qu'obiapat laissait passer les débuts nominaux avant la région.

### Départage tranché par la fréquence des codes-barres (suggestion de l'utilisateur, 2026-11-05)

Principe : un code-barres dominant dans le jeu est probablement le vrai ; la version d'un code-barres à frontière litigieuse qui coïncide avec les versions fréquentes est probablement la bonne. Données : 400 000 premières lectures de `nanopore_test_plate_A.fastq.gz`, `.ngsfilter` figé (copie), `obimultiplex` par défaut ; deux binaires temporaires qui ne diffèrent que par l'appel : `AllMatchesRealigned` (= comportement d'obiapat, `LocatePattern`) contre `AllMatches` (début réel le plus tardif). 409 717 codes-barres, 92 921 distincts ; 80,4 % d'entre eux ont une séquence exacte présente au moins deux fois (le plus fréquent : 44 610 copies), donc la déréplication est exploitable. (Les « attribués » comptés par `grep -c '^@'` plus haut sont légèrement surestimés : des lignes de qualité commencent par `@`.)
Référence = les codes-barres des enregistrements **identiques** dans les deux exécutions (386 203). 18 372 enregistrements ont le même échantillon et les mêmes tags mais un code-barres différent ; dans **tous**, la version « début réel » est la version `LocatePattern` **plus 1 à 4 bases** (au début ou à la fin). Verdict par la fréquence dans la référence : version `LocatePattern` plus fréquente : 5 751 ; version début réel plus fréquente : 2 116 ; égalité : 60 ; les deux absentes de la référence (rares) : 10 445. Parmi les cas décidables, `LocatePattern` retrouve le code-barres dominant dans **73 %** des cas, le début réel dans **27 %**. Lecture : la fenêtre d'amorce la plus longue (celle de `LocatePattern`) est plus souvent la bonne que ma fenêtre la plus courte. Réserves : 57 % des paires ne sont pas décidables ; un code-barres fréquent peut aussi être fréquent parce qu'une frontière systématique est fausse ; les lectures dont le nombre de barcodes diffère entre les deux exécutions (5 142 / 6 946 enregistrements) ne sont pas comparées.
Suite possible : évaluer avec ce critère la règle « fenêtre la plus longue » de ce paquet (début le plus précoce, dernière fin de même coût), qui serait plus proche de `LocatePattern` sans en dépendre.

### Quelle version récupère le plus de codes-barres ? (400 000 lectures MinION, mêmes binaires temporaires, 2026-11-05)

Comptage exact par lecture des enregistrements (le `grep -c '^@'` précédent était faux) : `LocatePattern` (= obiapat) **409 717** codes-barres attribués à un échantillon, dans 230 899 lectures ; `AllMatches` (début réel) **411 521** (+1 804, +0,44 %), dans 231 610 lectures (+711) ; non identifiées : 369 952 contre 368 151. Différence par enregistrement (lecture, échantillon, tags, direction) : **2 458 uniquement avec le début réel, 654 uniquement avec `LocatePattern`**. Les 2 458 ont presque tous 3 à 5 erreurs sur l'amorce directe (`primer_mismatches` = 5 dans ce `.ngsfilter`) : l'`AllMatches` d'obiapat les supprime quand le score de `LocatePattern` dépasse `maxerr`. Fiabilité (séquence retrouvée parmi les 390 454 codes-barres sur lesquels les deux versions s'accordent) : 58,5 % des 2 458 (51,3 % vus au moins 5 fois) contre 67,7 % des 654.
Bilan : le début réel récupère plus de codes-barres (surtout à amorce très dégradée), `LocatePattern` place plus souvent la frontière sur le code-barres dominant (73 % des cas décidables) ; les deux critères ne sont pas incompatibles : une règle « fenêtre la plus longue » avec la recherche arrière devrait cumuler les deux (à tester avec ces critères).

### Quatre variantes comparées sur 400 000 lectures MinION (2026-11-05)

Variantes (binaires temporaires, `multimatch.go` restauré sur `AllMatchesRealigned`) : (V0) `LocatePattern` partout = ancien comportement ; (V1) `AllMatches` début réel (le plus tardif) partout ; (V2) `AllMatchesLongest` partout (nouveau, `search.go` : parmi les fenêtres qui se chevauchent et ont le même coût la dernière, début le plus précoce de ce coût, par la recherche arrière avec `earliest`) ; (V3) mélange proposé par l'utilisateur : `forward` et `reverse` en début réel, `creverse` et `cforward` (compléments inverses) en `LocatePattern`. Tests : `TestBackwardEarliestAgainstTruth`, `TestAllMatchesLongestAgainstOracle`. Référence neutre = les 385 538 codes-barres identiques dans les quatre exécutions (78 960 séquences distinctes) ; « confirmé » = séquence vue au moins 5 fois dans cette référence.

| variante | codes-barres | confirmés | non confirmés | par rapport à V0 |
|---|---|---|---|---|
| V0 `LocatePattern` | 409 717 | 306 087 (74,7 %) | 103 630 | |
| V1 début réel | 411 521 | 305 383 (74,2 %) | 106 138 | +2 458 / −654 |
| V2 la plus longue | 409 150 | **307 409 (75,1 %)** | **101 741** | +395 / −962 |
| V3 mélange | 409 691 | 303 975 (74,2 %) | 105 716 | +32 / −58 |

- **Le plus de codes-barres au total : V1** (+1 804). Ses 2 458 codes-barres propres viennent surtout de la fenêtre `cforward` (2 005 ; `reverse` 374, `creverse` 72, `forward` 7) ; 52 % seulement ont une séquence confirmée ; les 654 propres à V0 : 63 %.
- **Le plus de codes-barres confirmés : V2** (+1 322 par rapport à V0, avec 1 889 non confirmés de moins), alors que son total est un peu plus petit (−567). Gagnés (395) et perdus (962) ont la même part de confirmés (60 % / 59 %), le gain vient des codes-barres que V0 et V2 gardent tous deux mais dont la frontière de V2 est plus souvent celle du code-barres dominant.
- **V3 (mélange) ne récupère presque rien** (+32 / −58) et a le moins de confirmés : les codes-barres récupérés par V1 passent par les compléments inverses (surtout `cforward`), que V3 laissait à `LocatePattern`. L'hypothèse « début réel pour forward, `LocatePattern` pour les compléments inverses » est donc infirmée sur ces données (le mécanisme exact de la perte par `LocatePattern` sur `cforward` n'est pas vérifié ; hypothèse : le fragment de réalignement est tronqué par la fin de la lecture).
- Limites : « confirmé » favorise les frontières qui reproduisent des versions fréquentes ; il ne prouve pas qu'elles sont vraies ; un seul jeu de données, un seul ngsfilter.

### Où sont les codes-barres erronés : quel bord de quel match ? (400 000 lectures, 2026-11-05)

Le code-barres d'`obimultiplex` est `[fin du match direct, début du match en complément inverse)` en coordonnées de la lecture, dans les deux directions : début du code-barres = **fin** du match direct (`forward` ou `reverse`), fin du code-barres = **début** du match en complément inverse (`creverse` ou `cforward`). Sur les 18 372 paires litigieuses (LocatePattern contre début réel, même lecture/échantillon/tags), la version « début réel » est toujours le code-barres le plus long. « Erroné » = la version minoritaire par rapport à la séquence dominante (>= 5 occurrences dans la référence commune aux quatre variantes) ; seulement ~1/3 des paires sont décidables.
| bord litigieux | paires | début réel erroné (code-barres rallongé) | `LocatePattern` erroné (code-barres mangé) | indécidable |
|---|---|---|---|---|
| début du code-barres = fin du match direct | 10 043 | **2 927** | 184 | 6 932 |
| fin du code-barres = début du match complément inverse | 7 920 | 1 608 | 1 456 | 4 856 |
| les deux bords | 409 | 22 | 3 | 384 |
Détail des cas décidés : fin du match direct, début réel erroné : décalage de 1 base (875), 2 bases (1 779), 3 et plus (273) ; par motif `reverse` 2 046, `forward` 881 ; `LocatePattern` erroné : 184 (`reverse` 101, `forward` 83). Début du match complément inverse : décalage toujours d'**1 base** ; début réel erroné 1 608 (`creverse` 1 086, `cforward` 522) ; `LocatePattern` erroné 1 456 (`creverse` 1 422, `cforward` 34).
Lecture : côté fin du match direct, la fenêtre la plus courte (première fin de coût minimal) est fautive (`LocatePattern` a raison dans 94 % des cas décidables) ; côté début du match complément inverse l'écart est toujours d'une base et les deux méthodes se trompent à peu près autant (52 % / 48 %), donc ni la fenêtre la plus courte ni la plus longue n'y suffit seule (dépend sans doute du contexte, non examiné).

**Correction de vocabulaire (2026-11-05)** : ces notes disent à plusieurs endroits que la fin d'une fenêtre est « exacte » (celle du balayage avant d'`apat`/agrep). Il faut lire : c'est la fin **d'un** alignement de coût minimal pour cette position, pas nécessairement la fin réelle du motif. Plusieurs positions de fin voisines sont signalées avec le même coût (dernière base du motif supprimée : fin une base plus tôt ; ou appariée par substitution avec la base suivante : fin une base plus tard) et `bestOfOverlaps` retient la première, donc la plus précoce ; `LocatePattern` finit plus tard dans 100 % des cas où les fins diffèrent, à coût égal. La fin est ambiguë pour 2,1 à 6,4 % des fenêtres (e = 3, vraies lectures). Exemple : motif `GADGHGCWCMGYGAHCSA`, e = 2 : `gaggtgctccgtgatcc` (17 bases) contre `gaggtgctccgtgatccc` (18 bases), 1 erreur dans les deux cas.

### Proposition de l'utilisateur : fournir le code-barres le plus long, trancher après `obiuniq` (2026-11-05)

Principe : `obimultiplex` ne doit pas perdre de base de code-barres ; les fenêtres d'amorce les plus courtes (début réel, première fin de coût minimal : `AllMatches`) donnent le code-barres le plus long ; les bases en trop (1 ou 2 le plus souvent) se retirent **après la déréplication** (`obiuniq`), où les comptages disent quelle est la version dominante (c'est le critère déjà utilisé ici). Mesure (400 000 lectures) : séquences distinctes 92 921 (`LocatePattern`), 95 292 (début réel, +2,6 %), 91 726 (fenêtre la plus longue) ; part des codes-barres qui ne diffèrent que par les bords d'une version plus fréquente : 1,3 % / 1,4 % / 1,1 % (mesure grossière : regroupement par 16-mer central, bords <= 6 bases). Réserves : les décalages valent 1 base (875 cas décidés côté fin du match direct), 2 bases (1 779), 3 bases ou plus (~9 %, jusqu'à 10) ; l'étape de fusion des variantes de bord en aval reste à fournir (`obiclean` non vérifié) ; les 2 458 codes-barres gagnés par cette variante ne sont confirmés qu'à 52 %.

### Les grands décalages (>= 5 bases) : qui a tort ? (400 000 lectures, 2026-11-05)

Taille du désaccord (somme des décalages de bord, paires litigieuses `LocatePattern` / début réel) : 1 base : 11 880 ; 2 : 4 569 ; 3 : 727 ; 4 : 357 ; 5 : 191 ; 6 : 171 ; 7 : 44 ; 8 : 11 ; 9 : 9 ; 10 : 4 ; deux bords à la fois : 409. Donc 90 % des paires diffèrent de 1 ou 2 bases, 430 (2,3 %) de 5 bases ou plus.
Cas examinés à la main (8 premiers, fenêtres directes `forward`/`reverse`, e = 5) : les deux fenêtres ont **exactement le même coût** (4 ou 5, le maximum permis), et la queue de l'amorce est dégradée par une rafale d'indels. Exemple : amorce `GGGCAATCCTGAGCCAAATC`, lecture `GGGCAATCCTGAGCC GGGGAAGTC ACTTCTGC…` ; début réel : fenêtre de 15 bases (les 5 dernières positions supprimées, coût 5) ; `LocatePattern` : fenêtre de 24 bases (`GGGG` insérés + une substitution sur `AAATC`, coût 5). Les deux sont valables pour le modèle de coût.
Verdict par fréquence sur les décidables (>= 5) : `LocatePattern` dominant (début réel erroné) 129, début réel dominant (`LocatePattern` erroné) 57, soit 69 % / 31 % ; pour les décalages de 8 à 10 bases, 12 cas décidés, **tous** en faveur de `LocatePattern`. Lecture : les grands décalages ne sont pas des absurdités de `LocatePattern` ; ce sont surtout des queues d'amorce dégradées que `LocatePattern` reconnaît (en acceptant des insertions) et que la fenêtre la plus courte laisse dans le code-barres. Réserve : 43 % seulement des paires >= 5 sont décidables ; la fréquence reste un critère indirect.

### Décision : `AllMatches` simple pour `obingslibrary` (utilisateur, 2026-11-05)

Argument de l'utilisateur : on discute de quelques bases sur quelques pour cent des lectures, sans vérité connue ; des règles ad hoc ajustées à un jeu de données ne valent rien ailleurs ; un `AllMatches` simple est arbitraire mais donne plus de codes-barres et n'est pas déconnant. Les bases en trop éventuelles se règlent après la déréplication. `pkg/obingslibrary/multimatch.go` appelle donc `AllMatches` (début réel le plus tardif, fin = première position de coût minimal) pour les quatre motifs. Vérifié : sorties identiques, enregistrement par enregistrement, à celles du binaire « début réel » mesuré (400 000 lectures, `.ngsfilter` figé) ; 411 521 codes-barres attribués contre 409 717 avec l'ancien (+0,44 %).
Code rendu inutile par cette décision, **non supprimé** (à décider) : `AllMatchesRealigned` (compatibilité exacte avec obiapat, utilisé seulement par un test), `AllMatchesLongest` et ses paramètres (`longest` dans `allMatchesReal`, `last` dans `bestOfOverlapsTie`, `earliest` dans `backwardStart`) avec leurs tests, `BestMatchBackward` ; `BestMatch` (réalignement par `LocatePattern`) reste utilisé par `obiannotate` et par `Marker.Match` (sans appelant).

### Simplification : un seul chemin, `AllMatches` (2026-11-05)

Fait sur demande de l'utilisateur. `pkg/obidnapat/search.go` ne contient plus que `AllMatches` (regroupement des chevauchements, vrai début le plus tardif par `backwardStart`) ; `BestMatch` = première fenêtre de plus petit nombre d'erreurs de `AllMatches` (équivalence vérifiée : 20 000 cas sur 20 000, puis `TestBestMatch`, y compris contre le `BestMatch` d'obiapat pour les motifs sans indels). Supprimés : `BestMatch` à réalignement par `LocatePattern`, `BestMatchBackward`, `AllMatchesRealigned`, `AllMatchesLongest`, les paramètres `longest`/`last`/`earliest`, le champ `raw`, la dépendance de `obidnapat` à `obialign` ; tests supprimés : `TestIndelBestMatchAgainstObiapat`, `TestBackwardAgainstLocatePattern`, `TestAllMatchesRealignedAgainstObiapat`, `TestBackwardEarliestAgainstTruth`, `TestAllMatchesLongestAgainstOracle`. `BenchmarkRefine` compare maintenant `BestMatch`/`AllMatches` à ceux d'obiapat.
Effets : `obimultiplex` inchangé par rapport à `AllMatches` (400 000 lectures : 0 enregistrement différent). `obiannotate` (qui utilise `BestMatch`) : sans `--allows-indels`, annotations identiques à l'ancien binaire (toutes les combinaisons des deux amorces et de `--pattern-error` 0, 1, 3, 5) ; avec `--allows-indels`, 0 % de différences à e = 0, 1,8 % (e = 1), 5,9 % (e = 3, amorce directe : 508 sur 8 590), 8,8 % (e = 5) ; à e = 3 : fin plus tôt de 1 base (324), 2 bases (105), 3 bases (13), début plus tard d'1 base (76), **jamais de changement du nombre d'erreurs**, jamais plus de 3 bases d'écart, même nombre de lectures annotées.
Reste à migrer : `obipcr` + `pkg/obiapat/pcr.go`, `obisplit`, `obigrep`. `Marker.Match` / `ExtractBarcode` (obingslibrary) sans appelant : candidat à la suppression.

### `obimultiplex` : recherche simultanée des amorces avec `PatternSet` ? Profil CPU (2026-11-05)

Question de l'utilisateur : chercher d'un coup toutes les amorces forward et reverse avec `PatternSet` ? Réponse par un profil CPU de `ExtractMultiBarcode` sur un cœur (`BenchmarkExtractMultiBarcode` dans `pkg/obingslibrary/realdata_test.go`, ignoré sans `OBIDNAPAT_FASTQ` et `OBIDNAPAT_NGSFILTER` ; 20 000 lectures MinION, ton `.ngsfilter` : **un seul marqueur**, 100 échantillons, `matching,indel`, `tag_mismatches,1`, `primer_mismatches,5`, indels permis). Parts de `ExtractMultiBarcode` (7,78 s) : `SampleIdentifier` **6,19 s (80 %)**, dont `Levenshtein` 4,51 s (58 %), parcours de la table des échantillons 1,26 s, allocations (`makeslice`) ~1 s ; recherche des amorces (`AllMatches`) **0,60 s (7,7 %)**, dont le balayage d'indels 0,53 s. Pourquoi : `ClosestForwardTag`/`ClosestReverseTag` parcourent la table des **100 paires de tags** (`marker.samples`) et calculent 100 distances de Levenshtein par tag observé, alors qu'il n'y a que ~12 tags distincts par côté.
Conclusion : même si la recherche des amorces devenait gratuite, `ExtractMultiBarcode` gagnerait au plus ~8 % (réaliste ×2 à ×3 sur ces 7,7 % : 4 à 5 %). De plus `PatternSet` ne porte pas les indels (le mode nanopore). L'intérêt croît avec le nombre de marqueurs (4 recherches par marqueur) : non mesuré faute de bibliothèque multi-marqueurs ; sur x86 large (AVX2/AVX-512) : non mesuré non plus. Meilleur levier mesuré : la mise en correspondance des tags (tags distincts au lieu des paires, sortie immédiate sur correspondance exacte, `Levenshtein` sans allocation, cache par tag observé), non faite.
Correctif au passage : `MakePatternSet` **acceptait** les motifs avec indels et appliquait la récurrence des substitutions (résultats faux en silence) ; il les refuse maintenant (`TestPatternSetRefusesIndels`).

### `obimultiplex` : tags distincts au lieu de toutes les paires (2026-11-05)

Correction demandée par l'utilisateur (« aucune raison de relire cinquante fois les mêmes tags »). `ClosestForwardTag`/`ClosestReverseTag` (multimatch.go) cherchent dans les tags **distincts** du marqueur (`Marker.distinctTags`, liste triée, construite à la première utilisation et refaite après un `GetPCR`, `atomic.Pointer`, sans verrou), sortent dès qu'un tag est identique (distance 0), et la règle d'égalité est inchangée (deux tags différents à la même plus petite distance : pas de tag). `Levenshtein` n'alloue plus pour les chaînes de 64 caractères ou moins. Tests : `TestClosestTagAgainstAllPairs` (même résultat que l'ancien parcours de toutes les paires, égalités comprises, 3 000 bibliothèques aléatoires), `TestClosestTagAfterNewSample`, `TestClosestTagConcurrent` (`-race`), `TestLevenshteinNoAllocation` ; mutation de la règle d'égalité détectée.
Mesures (ton `.ngsfilter`, 100 échantillons, un marqueur, `matching,indel`) : `obimultiplex` sur 100 000 lectures et un cœur **10,6 s → 5,0 s**, sorties identiques (0 enregistrement différent) ; profil de `ExtractMultiBarcode` (20 000 lectures) 7,78 s → 2,36 s (×3,3), `SampleIdentifier` 6,19 s → 0,57 s. Nouvelle répartition de `ExtractMultiBarcode` : recherche des amorces (`AllMatches`) 0,76 s (**32 %**, contre 7,7 % avant), `SampleIdentifier` 0,57 s (24 %), affectations d'annotations (`mapassign`) 0,5 s, `Subsequence` 0,31 s, `fmt.Sprintf` 0,20 s. La recherche simultanée des amorces redevient donc intéressante (jusqu'à un tiers du temps de la fonction), mais `PatternSet` ne porte pas les indels, qui sont le mode nanopore.

### `PatternSet` avec indels : est-ce faisable ? Prototype (2026-11-05, proto_indel_set_test.go)

Question de l'utilisateur. Réponse par un prototype jetable : `TestProtoPackedIndel` empaquette plusieurs motifs avec indels dans un mot de 64 bits et exécute la récurrence de `scan_indel.go` avec le bit de départ **de chaque motif** (`| start` à la place de `| 1`, comme pour les substitutions), puis compare chaque motif, occurrence par occurrence, à la recherche d'un seul motif (`indelHits`). Résultat : **7 474 123 occurrences identiques** sur 20 000 mots à plusieurs motifs (motifs de 1 à 22 positions flancs compris, e de 1 à 5, `#`, ambiguïtés, N). Mutations : substitution ou niveau 0 avec le seul bit de départ du premier motif -> le test échoue ; contrôle : retirer du terme de suppression `(n << 1)` le bit qui passe d'un motif au suivant ne change rien (ce passage est inoffensif : le bit 0 du motif suivant est déjà posé par le terme de substitution, ou masqué par `allow`). Le test de la récurrence est local au test (aucun code de production) : il est la base des tests de la vraie implémentation.
Ce qui reste à écrire : noyaux de mots avec indels (génération pour 1 à 8 niveaux, entrelacement par lots comme pour les substitutions, noyau à tableau au-delà) ; `MakePatternSet` accepte les motifs avec indels et rend, par (motif, brin), les fenêtres comme `AllMatches` (regroupement des chevauchements puis vrai début par `backwardStart` avec les tables de chaque motif), triées par début ; tests contre `AllMatches` motif par motif ; restructuration d'`ExtractMultiBarcode` (une recherche pour tous les marqueurs, les jetons alimentent l'automate). Gain estimé, **non mesuré** : recherche des amorces 0,76 s sur 2,36 s ; 4 motifs de 20-22 positions tiennent dans 2 mots (3 motifs dans le premier) ; accélération attendue ×2 à ×3 de cette part, soit ~20 % d'`ExtractMultiBarcode` ; davantage avec plusieurs marqueurs.

### Qui utilise encore `obiapat` ? (2026-11-05, `go list -deps` sur chaque programme de `cmd/obitools/`)

Sept programmes en dépendent : `obigrep`, `obipcr`, `obisplit` (usage direct) et `obiannotate`, `obiscript`, `obicleandb`, `obiconsensus` (dépendance indirecte). Paquets qui importent `pkg/obiapat` : `pkg/obitools/obigrep/options.go` (`obiapat.IsPatternMatchSequence`, pour `--approx-pattern`, défini dans `pkg/obiapat/predicat.go`), `pkg/obitools/obipcr/` (`options.go`, `pcr.go` : moteur de PCR in silico, options d'extension et circulaire, `pkg/obiapat/pcr.go`), `pkg/obitools/obisplit/` (`obisplit.go`, `options.go` : `ApatPattern`, `MakeApatSequence`). Les quatre autres programmes ne l'utilisent qu'**à travers `obigrep`** : `obiannotate` (pourtant migré, il importe `pkg/obitools/obigrep`), `obiscript`, `obicleandb` importent `obigrep`, et `obiconsensus` importe `obiannotate`. Migrer `obigrep` (un seul prédicat) retire donc quatre programmes de la liste d'un coup ; il resterait `obisplit` puis `obipcr`. Plus aucun usage de `obiapat` : le code C (cgo) du dépôt pourrait alors partir. `obimultiplex`, `obitagpcr` ne l'utilisent plus.

### Migration d'`obigrep` faite (2026-11-05)

`pkg/obitools/obigrep/options.go` : `obiapat.IsPatternMatchSequence` -> `obidnapat.IsPatternMatchSequence` (nouveau, `pkg/obidnapat/predicate.go`, même signature : motif, erreurs, deux brins, indels ; s'arrête avec un message si le motif est invalide). Correction au passage dans `ApatPattern.IsMatching` : pour un motif à indels, **toute** occurrence trouvée compte, même si son début nominal tombe avant la région (motif tronqué de ses premières bases au tout début de la séquence : obiapat répondait vrai, la version filtrée par le début nominal répondait faux) ; la recherche démarre à l'origine de la région, l'occurrence y est donc. Tests : `TestIsMatchingIndelAgainstOracle` (contre l'oracle, dont des cas où seule une fenêtre au début nominal avant la région existe), `TestIsPatternMatchSequenceAgainstObiapat` ; mutation (retour au filtre sur le début nominal) détectée par les deux.
Vérification de bout en bout (ancien et nouveau `obigrep`, 100 000 lectures MinION, `--approx-pattern`, 4 motifs dont l'amorce inverse à R, un motif à `N` de queue et un à `N` de tête ; `--pattern-error` 0, 1, 3 ; avec et sans `--allows-indels` ; deux brins ou `--only-forward`) : **48 configurations, 0 enregistrement différent**. Vitesse (e = 3, indels, un motif) : 0,15-0,16 s avant, 0,09-0,10 s après.
**Décision de l'utilisateur sur les motifs dont on n'a pas le début (2026-11-05)** : un motif avec indels dont les premières (ou dernières) bases manquent parce que la séquence commence (ou finit) là **compte comme une correspondance**, les bases manquantes étant des erreurs comme les autres (distance d'édition, comportement d'obiapat) : option A, retenue. Raison de la question : avec les substitutions seules un motif doit tenir entièrement dans la séquence, avec les indels non ; en bord de lecture une base absente n'est pas une preuve de différence. Option B écartée : exiger que le motif soit entièrement observé (premier et dernier éléments alignés sur des bases), qui aurait aussi rejeté les cas intérieurs à la lecture où la première base d'une amorce est supprimée (oligos tronqués). Conséquence : `IsMatching`, `AllMatches` et `BestMatch` acceptent ces cas ; sans effet sur `obimultiplex` (le tag voisin manque aussi en bord de lecture). La « correction » d'`IsMatching` décrite plus haut rétablit ce comportement, elle ne corrige pas un défaut d'obiapat.

**Défaut d'obiapat trouvé** : son complément inverse d'un motif avec négation est faux (`!C#BGSNV[HC]A` -> `t[gd]bnscvg!#` au lieu de `T[DG]BNSCV!G#`, le `!` est déplacé), donc avec les deux brins `--approx-pattern` d'obiapat ne cherche pas le bon motif sur le brin inverse ; `obidnapat` donne le bon (vérifié à la main sur une fenêtre). Le test différentiel ne compare pas ces motifs sur les deux brins.
Changements de comportement visibles par les utilisateurs d'`obigrep` : un octet ambigu de la séquence (`n`, `r`...) correspond maintenant si ses bases possibles recoupent celles du motif (il comptait comme une erreur) ; un motif de plus de 63 positions (dont les `N` de tête et de queue quand les indels sont permis) est refusé ; négation `!` correcte au complément inverse.
Dépendances restantes (`go list -deps`) : **`obipcr` et `obisplit` seulement**. `obiannotate`, `obiscript`, `obicleandb` et `obiconsensus` n'importaient `obiapat` qu'à travers `obigrep` : ils n'en dépendent plus.

### `obisplit` supprimé (utilisateur, 2026-11-05)

« `obisplit` peut être supprimé, il n'a plus de raison d'être. » Supprimés (avec `rm`, aucune commande de gestion de versions) : `cmd/obitools/obisplit/`, `pkg/obitools/obisplit/`, `obitests/obitools/obisplit/` (son `test.sh` était trouvé par le `find` du Makefile) et l'entrée `obisplit` de `setattribute_refs.json` (fichier d'analyse, JSON toujours valide). Rien d'autre n'importait le paquet ; `go build ./...` passe. Documentation (`obitools4-doc`) nettoyée à la demande de l'utilisateur : page `content/docs/commands/experimental/obisplit/`, `layouts/shortcodes/cmd-options/obisplit/` (4 fichiers), la ligne de `content/_index.md` (dans un tableau commenté), l'entrée de `data/commands.yaml` et les 11 attributs `obisplit_*` de `data/tags.yaml` ; le site Hugo se construit sans erreur et sans aucune trace d'`obisplit` (34 avertissements, aucun lié à `obisplit` ni à `splitting` : pages ou images introuvables, un lien `obikmersim`, dépréciations de Dart Sass dans le thème ; non examinés) ; la catégorie `splitting` n'est plus utilisée nulle part. **Seul `obipcr` dépend encore d'`obiapat`** (`pkg/obitools/obipcr/` et `pkg/obiapat/pcr.go`, séquences circulaires).

### Migration d'`obipcr` faite (2026-11-06)

Le moteur de PCR in silico passe de `pkg/obiapat/pcr.go` à `pkg/obidnapat/pcr.go` + `pcr_options.go`. API : `NewPCR(options ...PCROption) (*PCR, error)` (les amorces sont compilées une fois, une amorce invalide renvoie une erreur), `(*PCR).Amplicons(seq)`, `Slice(seqs)`, `SliceWorker()` ; options `PCRPrimers(fwd, e, rev, e)`, `PCRMinLength`, `PCRMaxLength`, `PCRExtension`, `PCRFullExtension`, `PCRCircular`. Une seule fonction (`amplicons`) sert les deux orientations (l'ancien code avait deux blocs de 80 lignes). Abandonnés car inutilisés : `PCRSim`, `BatchSize`, `ParallelWorkers`. `pkg/obitools/obipcr/{pcr,options}.go` utilisent `obidnapat` (plus de `Free()`).
Défauts de l'ancien moteur corrigés : (1) **en circulaire avec `-D`, l'ancien `obipcr` s'arrêtait** (`from out of bounds -1 < 0`) dès qu'un flanc franchissait la jonction côté gauche (`Subsequence` refuse un début négatif) : le début est maintenant ramené dans la séquence ; (2) en circulaire, pour l'orientation inverse, la longueur d'un amplicon qui franchit la jonction utilisait la longueur de la mauvaise amorce : l'ancien moteur manque des amplicons quand les deux amorces n'ont pas la même longueur (12 cas sur ~15 000 tirages aléatoires, **tous circulaires, aucun en linéaire** : la marge de 64 bases qu'obiapat ajoute à la zone de recherche masque l'autre imprécision, qui utilisait aussi la mauvaise longueur) ; (3) un amplicon avec ses flancs plus long que la molécule circulaire n'est plus rapporté (l'ancien code en rendait un fragment absurde) ; (4) `--max-length` <= 0 veut dire « pas de limite » (l'ancien code ne trouvait rien avec -1, la valeur par défaut d'une option pourtant obligatoire).
Tests : `TestPCRAgainstOracle` (oracle écrit d'après la définition : fenêtres de l'oracle naïf, indices modulo la longueur ; linéaire, circulaire, longueurs, extensions, amorces de longueurs différentes), `TestPCRAgainstObiapat` (amorces de même longueur, hors circulaire avec extension où l'ancien s'arrête) ; mutations détectées (longueur de l'amplicon qui franchit la jonction, décalage du flanc). `-race` vert.
Vérification de bout en bout (ancien et nouveau `obipcr`, 300 génomes mitochondriaux de `mito.vert.fasta.gz`, amorces téléostéens `ACACCGCCCGTCACTCTC` / `CCAAGTGCACCTTCCGGT` ou `CTTCCGGTACACTTACCATG`, `-e` 0, 1, 3, 9 combinaisons de `-l`, `-L`, `-D`, `--only-complete-flanking`, `-c` : **54 configurations, 0 différence** pour les amplicons dont les fenêtres d'amorces sont en ACGT). Sur 4 000 génomes, `-e 2` : ancien 3 037 amplicons, nouveau 110 061 dont 107 025 avec un `n` dans une fenêtre d'amorce (43 génomes) ; les 3 036 autres sont identiques à l'ancien sauf un, dont seul le nombre d'erreurs diffère (`y` dans la séquence : 1 erreur au lieu de 2).
**Problème (résolu, voir « Règle des bases réelles » ci-dessous)** : avec la règle de recoupement, un `n` de la séquence correspond à n'importe quelle base sans erreur ; une suite d'au moins 18 `N` (trou d'assemblage) « correspond » donc exactement à une amorce de 18 bases à chaque position, d'où des milliers d'amplicons `n`, `nn`, `nnn`... (obiapat comptait un `n` comme une erreur). Piste : un `n` (aucune information) compte comme une erreur sauf devant un joker du motif, les codes partiels (`r`, `y`...) restant en recoupement ; effet sur toute la bibliothèque (`obimultiplex`, `obigrep`, `obiannotate`), à mesurer.
**`obiapat` n'a plus aucun utilisateur** hors des tests différentiels de `pkg/obidnapat` (`go list -deps` : aucun programme ; aucun paquet non test ne l'importe). Suppression de `pkg/obiapat` (code C, cgo) : seulement sur demande, et ces tests différentiels devraient alors disparaître ou être gardés autrement.

### Règle des bases réelles (utilisateur, 2026-11-06)

**Décision** : une fenêtre n'est rapportée que si son nombre de bases réelles (ni `n` ni `N`) est **strictement supérieur** à son nombre d'erreurs, pour que le motif ne soit pas apparié par ses seules erreurs. Appliquée à **toute la bibliothèque** (`obipcr`, `obimultiplex`, `obigrep`, `obiannotate`...), **à chaque occurrence, avant tout regroupement** (`bestOfOverlaps`, `FilterBestMatch`, `AllMatches`) : une fenêtre rejetée ne peut pas être choisie comme meilleure d'un groupe et masquer des voisines valides. Variante plus stricte (les bases réelles qui *concordent* plus nombreuses que les erreurs) écartée : 72 amplicons restants au lieu de 385, mais elle jugerait aussi des fenêtres légitimes peu informatives.
Cas particuliers : (1) un motif qui permet au moins autant d'erreurs que de positions ne correspond plus à une base qui diffère (`A`, `-e 1` sur `c` : rejeté), même sans `N` ; (2) avec indels, la fenêtre comptée est la fenêtre **nominale** `[fin - Len() + 1, fin + 1)` (la vraie n'est connue qu'après le regroupement) et la partie qui déborde du début de la séquence ne compte pas comme réelle ; (3) un motif fait de jokers seuls n'est plus rapporté sur une fenêtre de `N`.
**Mise en œuvre** (`noise.go`) : *pas* de post-traitement global, et les noyaux déroulés générés (`gen_kernels.go`) ne sont pas modifiés (la variante « noyaux `…Filtered` distincts » proposée à l'utilisateur a été remplacée par ceci, mêmes propriétés, sans toucher aux fichiers générés). Un motif seul (`search`, `indelHits`) : `hasN(seq)` (deux `bytes.IndexByte`, vectorisés), puis `needsCheck` = séquence avec `N`, ou motif avec indels, ou `levels >= Len()` ; si faux, **aucun surcoût** (chemin d'avant, BNDM compris). Si vrai, le texte est lu par blocs de `noiseChunk` = 4096 octets (l'état du noyau passe d'un bloc à l'autre) et `dropNoise` supprime en place les fenêtres rejetées après chaque bloc : le tampon des occurrences ne dépasse jamais un bloc, même dans une suite de millions de `N`. BNDM est désactivé quand la séquence contient un `N` (Shift-And à la place). `PatternSet` : le filtre est dans `word.emit`, point d'émission unique de tous les noyaux (déroulés, entrelacés, vectoriels), via `windows.seq/noisy/circular` ; une fenêtre rejetée n'entre jamais dans le tampon. Coût par occurrence trouvée : O(m) (`realBases`).
Tests : les oracles (`oracle_test.go`, `oracleIndelRaw` dans `indel_test.go`) appliquent la règle avec leur propre comptage (`naiveRealBases`) ; les comparaisons avec `obiapat` (qui n'a pas la règle) écartent les motifs avec `e >= Len()` et, pour les indels, les séquences plus courtes que `Len() + e` ; `noise_test.go` (suites de `N`, fenêtres tout en `N`, motif court, cohérence `PatternSet` / motif seul). Suite complète verte, y compris `OBIDNAPAT_ITER=30000` (115 s) ; build complet et tests des dépendants (`obingslibrary`, `obialign`, `obiclust`, `obioligotag`, `obiprimers`) verts.
Vérification de bout en bout (`obipcr`, les 4 000 premiers génomes de `/tmp/mito.vert.path.fasta.gz`, mêmes amorces, `-e 2 -l 1 -L 1000`) : ancien 3 038 amplicons, nouveau sans la règle 126 264, **avec la règle 3 532** ; aucun amplicon de l'ancien n'est perdu, un seul diffère entre les deux (nombre d'erreurs, code ambigu), les **494 en plus ont tous un `n` dans une fenêtre d'amorce** (frontières de suites de `N` : `nnnnnnnnnnnnnnnncc` pour 18 bases avec 1 erreur, soit 2 bases réelles > 1) : 494 sur ~123 000 amplicons suspects, **0,40 %** (la mesure faite à part donnait 385 avec `-L` différent, les paramètres exacts de la première mesure n'étaient pas notés). Temps : ancien 0,29 s, nouveau sans la règle 1,26 s, **avec la règle 0,15 s** (les suites de `N` ne produisent plus d'occurrences).
`obimultiplex` ne change pas sur les lectures MinION (pas de `N`) ; non revérifié de bout en bout après la règle.

**Prompt de reprise pour la suite** : `OBIMULTIPLEX_PATTERNSET_PROMPT.md` (recherche simultanée des amorces de tous les marqueurs dans `obimultiplex`, `PatternSet` avec indels, de bout en bout : règles, état, conception à valider, protocole de vérification, pièges).

### `PatternSet` avec indels et recherche simultanée des amorces dans `obimultiplex` (2026-11-06)

**Fait** (feu vert de l'utilisateur sur l'API et les choix ci-dessous).
- `MakePatternSet` accepte les motifs avec indels (le refus et `TestPatternSetRefusesIndels` sont supprimés). Mots de **deux sortes** : `s.words` (substitutions, inchangé) et `s.iwords` (indels), empaquetés par classe de niveaux (`packWords`). Récurrence empaquetée du prototype : `patternset_indel.go` (`initIndel`, `scanIndelWords`, `runIndelN` à tableau, `emitIndel`), noyaux déroulés 1 à 8 niveaux et noyaux **entrelacés** de 2 et 3 mots (`gen_kernels.go` : `genIndelWords`, `genIndelInterleaved`, fichiers `patternset_indel_unrolled.go`, `patternset_indel_il*.go`) ; `batchFor` est réutilisé tel quel (3 mots jusqu'à 3 niveaux, 2 à 4-5, 1 au-delà), **non remesuré pour les indels**.
- Séquences circulaires : les membres avec indels passent par `compiled.search` (voie « solo »), correct mais non optimisé ; `AppendAllMatches` ne traite que le linéaire.
- `AppendAll` (brut) : pour un motif avec indels = `AppendAllIndex` (fenêtres nominales, début nominal >= début de région). `AppendAllMatches` (`patternset_matches.go`) = `AllMatches` par (motif, brin) pour tous les motifs : tri par comptage par (motif, brin), `appendBestOfOverlaps`, `backwardStart` avec les tables du motif, région appliquée au vrai début, tri final `(Start, Pattern, brin +1 d'abord)`. Les tampons sont dans un `sync.Pool` du jeu.
- `obingslibrary` : `primersearch.go`, `NGSLibrary.search` construit par `Compile2` (marqueurs numérotés par ordre de `(Forward, Reverse)` ; avant, ordre d'itération d'une map, donc **non déterministe**, y compris la sortie d'`obimultiplex` avec plusieurs marqueurs : 206 lignes d'écart entre deux exécutions du binaire de référence sur 40 marqueurs). Un seul `AppendAllMatches` par lecture. Pour retrouver la sémantique, les fenêtres `creverse`/`cforward` d'un marqueur et d'une orientation antérieures ou égales au début de la première fenêtre `forward`/`reverse` (la plus petite fin) sont écartées : sinon elles casseraient l'appariement des autres marqueurs. L'ordre des jetons de même début reprend celui de l'ancien tri en pratique (stable pour <= 12 jetons) : `(début, marqueur, rôle)` avec rôle = forward, creverse, reverse, cforward. Le tri par `(Start, Pattern, brin)` proposé d'abord donnait 48 lectures différentes sur 20 000 (synthétique, 1 marqueur) à cause de cet ordre.

**Écart connu** (quantifié, jamais caché) : l'ancienne recherche de l'amorce de la seconde extrémité sur une région tronquée (à partir de `premier début + 1`) gardait parfois une fenêtre parasite (5 erreurs, à l'intérieur de l'amorce de la première extrémité) que, sur la lecture entière, une meilleure fenêtre commençant plus tôt absorbe. `TestPrimerSearchAgainstSeparate` (variables `OBIDNAPAT_FASTQ`, `OBIDNAPAT_NGSFILTER`, `OBIDNAPAT_READS`) compare les jetons : 400 000 lectures réelles **0 différence** ; synthétique 20 000 lectures : 9 (1 marqueur), 0 (4), 1 (10), 2 (40) ; toutes les différences sont des jetons négatifs, jamais une amorce de première extrémité (le test échoue sinon). Cela peut changer l'appariement (le jeton parasite consommait l'état 1 et faisait perdre le vrai amplicon).

**Vérification de bout en bout** (`/tmp/annot/e2e.sh`, 400 000 lectures réelles, un cœur, copie figée du `.ngsfilter`) : sorties attribuées et non identifiées **identiques** (0 ligne différente) pour `indels,true -e 5`, sans indels `-e 2`, `-e 0`, `-e 2 --with-indels`, `-e 4 --with-indels`. Synthétique (`/tmp/synth/gen.py`, lectures simulées avec amplicons encadrés par les amorces, 6 % de substitutions/indels, 10 paires de tags, 2 orientations) : `ok` identique pour 1, 4, 10 marqueurs (0 ligne), non identifiées 22 / 3 / 21 lignes (écart ci-dessus, sur 26 047 / 25 172 / 24 649 enregistrements) ; à 40 marqueurs les écarts sont dans le bruit de la référence.

**Mesures** (un cœur, 100 000 lectures synthétiques, minimum de 3 répétitions entrelacées, M3 Max chargé) : ancien -> nouveau : 1 marqueur 3,03 -> 2,93 s ; 4 : 3,53 -> 3,22 ; 10 : 4,68 -> 3,66 ; 40 : 9,71 -> 5,97 (-39 %). Lectures réelles (un marqueur, 400 000, une exécution chacune, bruité) : 19,2 -> 17,9 s (`-e 5`), 11,3 -> 11,1, 13,0 -> 12,6, 9,2 -> 9,3, 15,5 -> 15,1. Avec un seul marqueur le gain est nul ou faible, comme prévu ; avant les noyaux entrelacés (une seule chaîne par mot) le gain à 40 marqueurs était 9,73 -> 6,64. À 40 marqueurs la recherche reste l'essentiel du temps (`runIndel5` 68 % avant entrelacement).

**Tests** : `TestPatternSetAllMatches` (jeux aléatoires mélangés, e de 0 à 10, deux brins, régions, contre `AllMatches` motif par motif et `expectedAllMatches`), `TestPatternSetIndelRaw` (brut et circulaire), `TestPatternSetIndelBatches` (taille de lot forcée 1 à 3, niveaux 1 à 10, 600 tirages par niveau : 120 laissaient passer une mutation du niveau 6). Mutations détectées : état initial sans `full`, règle des bases réelles, filtre de région, vrai début, regroupement des chevauchements, insertion/suppression dans le noyau générique et dans des noyaux déroulés et entrelacés, bit de départ, niveau de sortie, mélange de niveaux dans un lot. Une mutation (retirer `| start` du premier terme au niveau 6) n'est pas détectée : presque équivalente, le terme de substitution pose déjà ce bit sauf avec `#` en tête de motif. `-race` vert aux tirages par défaut (`OBIDNAPAT_ITER=6000` dépasse 590 s avec `-race` depuis l'ajout de ces tests ; le lancer sur un sous-ensemble). Build `simd`, `gofmt`, `vet` verts.

**Pas fait** : profil CPU de référence refait ; `batchFor` remesuré pour les indels ; `sync.Pool` pour les tampons de `primerMatches` (le tableau `first`/`firstEnd` et `occ` sont alloués par lecture) ; bibliothèques multi-marqueurs réelles (il n'en existe pas).

## Suite

1. ~~Voie stricte~~ faite (BNDM). Reste : strict multi-motifs (Wu-Manber avec décalage par q-gramme, ou Aho-Corasick), BNDM sur séquence circulaire, Shift-And à plusieurs flux pour les motifs courts.
2. `PatternSet` : voir la section ci-dessus (en cours, version générique d'abord) ; puis SIMD par voies de motifs si la mesure le justifie.
3. Port de `ManberIndel` (avec la logique `obialign.LocatePattern` inchangée), pour `obiannotate` et `obingslibrary`.
4. Migration des appelants (`obipcr`, `obisplit`, `obigrep`, `obiannotate`, `obingslibrary`), un par un ; `obiapat` reste jusqu'à la fin.
5. Question ouverte : `X` dans une séquence (traité comme ni IUPAC ni joker). (La variante sans masque `allow` a été essayée et abandonnée, voir « Essais écartés ».)
