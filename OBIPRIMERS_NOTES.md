# obiprimers — notes de travail

Ré-implémentation d'ecoPrimers dans obitools4. Ce fichier garde l'état des recherches,
des mesures et de l'avancement. Dernière mise à jour : 2026-11-03.

## État courant (à lire en premier)

**Les trois indices de qualité d'une paire** (définitions en vigueur depuis la section 25 ; elles remplacent
celles d'ecoPrimers, qui figurent encore dans les sections 1 à 18 comme historique) :

| Indice | Définition |
|---|---|
| **Bc** | taxons cibles amplifiés / taxons cibles |
| **Bs** | taxons cibles **identifiés** / taxons cibles amplifiés. Un taxon cible est identifié si **tous** ses barcodes ne sont portés que par lui, parmi **toutes** les séquences amplifiées par la paire (cibles, non-cibles, contre-exemples). Un barcode partagé rend mal identifiés **les deux** taxons. |
| **Bd** (nouveau) | part des barcodes portés par les cibles qui mènent sans ambiguïté à **un seul** taxon |

Un seul calcul pour chacun, aucune variante ni colonne alternative. Les trois sont dans **toutes** les sorties
CSV (`bc`, `bs`, `bd`). Toute séquence sans taxon au rang choisi est un taxon anonyme distinct (un barcode
qu'elle partage est ambigu). Le seuil `-T` reste à 0,6.

**Rôles des séquences** : cible (exemple), contre-exemple, **non-cible** (toutes les autres, qui ne sont plus
écartées : elles sont chargées en mémoire et leurs amplicons comptent dans Bs et Bd, pas dans les quorums).
Pour exclure des séquences de l'analyse, filtrer l'entrée en amont (`obigrep`).

**Conséquence pour les validations contre ecoPrimers (sections 12, 15, 17, 18)** : elles restent exactes sur
ce qu'elles comparent (paires, drapeaux good, séquences et taxons amplifiés, Bc, Yule, longueurs, GC, Tm), mais
**Bs et `identifiés` ne sont plus comparables** à ceux d'ecoPrimers ; ils sont validés autrement (section 25).

Sections en attente de décision de l'utilisateur : voie B (21.4), statistique de Tm (22.3-a), nombre de partitions
automatique (fait, section 26), mémoire (fait, section 26), tests shell, documentation, test `-e 0` (ecoPrimers n'avait pas terminé).

Sources de référence :
- Article : `~/Sync/travail/__MOI__/ecoprimers/Documents/Riaz et al. 2011 ... .pdf` (NAR 39:e145)
- Code C : `~/Sync/travail/__MOI__/ecoprimers/src/` (binaire : `/usr/local/bin/ecoprimers`)
- Tutoriel et données : `~/Documents/Enseignement/DNA metabarcoding spring school/ecoPrimers-Fish/`
  (`fish-primers.qmd`, `vertebrata.*` au format ecoPCR, `mito.vert.fasta.gz`,
  `ncbitaxo_20250211.tgz`, sorties de référence `Teleostei.ecoprimers`, `Cyprinidae.ecoprimers`)

## 1. Algorithme d'ecoPrimers (lu dans le code)

1. **Mots stricts** (`strictprimers.c`, `filtering.c`) : mots de longueur `lp` (18 par défaut, <= 31)
   encodés 2 bits/base, forme canonique `min(D,R)`, dédupliqués par séquence, présents sans
   erreur dans >= 70 % des séquences exemples (`-q`). Pré-filtre : table de comptage des 13-mers
   (4^13 entrées) qui écarte la plupart des mots avant la fusion triée (smoothsort + merge).
2. **Recherche approchée** (`aproxpattern.c`, `apat_search.c`) : chaque mot strict est cherché sur
   les 2 brins, <= `e` substitutions (Manber/agrep bit-parallèle), dans toutes les séquences
   (exemples et contre-exemples). Une amorce est « good » si elle touche >= 90 % des exemples
   (`-s`) et <= 10 % des contre-exemples (`-x`). Les amorces seulement « conservées » restent
   utilisables : une paire n'a besoin que d'une amorce « good ». `omask` vaut 0 dans cette
   recherche : l'option `-3` est vérifiée plus tard, au niveau de la paire.
3. **Appariement** (`pairs.c`) : par séquence, occurrences triées par position ; paires de
   distance dans (`lmin`, `lmax`) avec au moins une amorce « good » ; clé de paire canonique,
   stockage dans un arbre (`tsearch`) avec un amplicon par occurrence.
4. **Indices** (`taxstats.c`) : Bc = taxons exemples amplifiés / taxons exemples (au rang `-t`,
   espèce par défaut) ; Bs = taxons dont l'amplicon n'est partagé avec aucun autre taxon /
   taxons exemples ; yule = in/|E| - out/|C| ; tri sur yule x Bs ; seuil `-T` sur Bs (0,6).
5. **Tm** (`libthermo`) : plus proches voisins, sel SantaLucia ou Owczarzy ; Tm exact et pire cas.

Faiblesses relevées (à ne pas reproduire) :
- Recherche approchée mot par mot sur toutes les séquences : c'est le goulot en temps.
- Un amplicon stocké par occurrence et par paire : c'est le goulot en mémoire.
- `taxonomycoverage` et `taxonomyspecificity` rebalayent toute la base à chaque paire.
- Mots <= 31 bases ; sortie en tableau figé de 21 colonnes.

Non lus (à vérifier avant de les porter ou de les écarter) : `merge.c`, `smothsort.c`,
`ahocorasick.c`, `thermostats.c`, `nnparams.c`, `PrimerSets.c` (sets d'amorces, désactivé dans le
`main` : l'option `-p` est commentée).

## 2. Briques obitools4 réutilisables

| Besoin | Brique |
|---|---|
| Lecture, formats, batch, parallélisme | `obiformats`, `obiiter`, `obidefault` |
| Taxonomie, rangs, sous-clades | `obitax`, `obiseq/taxonomy_*`, options `--taxonomy`, `--restrict-to-taxon`, `--ignore-taxon` |
| K-mers canoniques (k <= 31) | `obikmer.EncodeCanonicalKmers`, `IterCanonicalKmers`, `ReverseComplement` |
| Comptage par partition, quorum, différence | `obikmer.KmerSetGroupBuilder`, `QuorumAtLeast/AtMost`, `Difference` |
| Contraintes sur amorces (GC, runs, IUPAC) | `obikmer.KmerConstraints` |
| Recherche approchée (indels, 2 brins) | `obiapat.ApatPattern` (cgo) |
| Distances entre mots | `obikmer.HammingDistance`, `KmerLCSBatch` (SIMD) |
| Alignement borné | `obialign.BandedLCS` |
| Patterns de CLI | `obipcr`, `obioligotag`, `obiclust` |

**Aucun calcul de Tm n'existe** dans obitools4 (grep santalucia / nearest-neighbor / meltingtemp :
rien). D'où le nouveau paquet générique `pkg/obithermo`.

SIMD : opt-in (`GOEXPERIMENT=simd`), repli scalaire portable. Même règle pour tout nouveau noyau.

## 3. Mesures de référence d'ecoPrimers (cette machine)

Commande : `ecoprimers -d vertebrata -e 3 -3 2 -l 30 -L 150 -c -r <taxid>`, deux processus
lancés en parallèle, un cœur chacun. Sorties dans `/tmp/ecoprimers-bench/` (temporaire).

| Jeu | Séquences | Paires totales | Paires retenues | Temps CPU | Mémoire max |
|---|---|---|---|---|---|
| Cyprinidae (`-r 7953`) | 325 | 42 682 | 3 448 | 134 s | 1,86 Go |
| Teleostei (`-r 32443`) | 3 909 | 9 407 | 407 | 350 s | 2,11 Go |

Les sorties sont **identiques** (diff sur les lignes hors commentaires) à
`Cyprinidae.ecoprimers` et `Teleostei.ecoprimers` du tutoriel. Ce sont les cibles de non-régression du
mode A.

Observation : Cyprinidae a 12 fois moins de séquences mais 4,5 fois plus de paires que
Teleostei, pour une mémoire comparable. Le goulot mémoire est donc le nombre de paires et le
stockage des amplicons, pas la taille de la base.

## 4. Sémantique de `-r`, `-i`, `-E` (testée, cas Salmo, taxid 8028)

`readdnadb.c` ne garde que les séquences qui sont exemples ou contre-exemples.

| Options | Exemples | Contre-exemples |
|---|---|---|
| `-r 8028` | 7 (6 espèces) | 0 : tout le reste est **supprimé** |
| `-r 8028 -i 7742` | 7 | 7 815 (7 697 espèces) : tous les autres vertébrés |
| `-r 8028 -i 8015` | 7 | 76 (69 espèces) : les autres Salmonidae |
| `-r 8015 -E 8028` | 76 | 7 : sens inverse, Salmo devient contre-exemple |

Avec `-r 8028 -i 8015` : 15 534 mots stricts (quorum 4/7 sur 7 séquences seulement), et un quorum
de contre-exemples de 7/76 (le défaut `-x 0,1` laisse passer 9 % du groupe proche, trop peu
exigeant pour une amorce spécifique). Les runs *Salmo* n'ont pas été menés jusqu'au bout (arrêtés
après l'en-tête).

Pièges de la base : *Salmostoma bacalia* est un cyprinidé (nom en « Salmo », taxid différent) : filtrer
par taxid, jamais par nom. Le tutoriel annonce 7 823 génomes ; `mito.vert.fasta.gz` en contient 7 822.

## 5. Plan d'implémentation

Paquets prévus :
- `pkg/obithermo` : plus proches voisins, corrections de sel, Tm exact et pire cas. Générique.
- `pkg/obiprimer` : cœur (`words.go`, `match.go`, `pairs.go`, `score.go`, `types.go`).
- `pkg/obitools/obiprimers` : `options.go`, `obiprimers.go`, `output.go`.
- `cmd/obitools/obiprimers/main.go`.

Deux modes :
- **A, conservé** (ecoPrimers) : k-mers présents dans >= q séquences exemples, via les partitions
  disque d'obikmer (remplace tri en place, smoothsort, table 4^13) ; recherche approchée par
  `obiapat` d'abord (version fidèle), optimisation par index seulement si elle est mesurée plus
  rapide.
- **B, spécifique** (une espèce ou quelques espèces proches) : ancres = k-mers de T absents de C
  (`Difference`) plus k-mers à distance de Hamming >= d de tout k-mer de C (mismatches en 3'
  plus discriminants) ; paires entre ancres spécifiques, ou ancre spécifique + amorce conservée.
  Le critère « good » est inversé. Définition précise de « spécifique » à valider avant le code.
  Cas de test : *Salmo* (cible) contre le reste des Salmonidae.

Sortie : **CSV** comme tous les obitools, ordre des colonnes d'ecoPrimers, en-têtes nommés.
Génomes circulaires (`-c`) gérés de bout en bout. Validation aval : `obipcr -e 5` sur chaque
paire, comme dans le tutoriel.

Ordre :
1. `obithermo` (tests contre `CalculTM.xls` et colonnes Tm de `Teleostei.ecoprimers`)
2. Mode A, mots stricts sur `KmerSetGroup` (liste identique à ecoPrimers)
3. Recherche approchée (`obiapat`, puis index si mesuré utile)
4. Appariement + indices (index taxon -> séquences calculé une fois)
5. CLI et sortie CSV
6. Mode B (spécification validée d'abord)
7. Parallélisme (`Split()`, `obidefault.ParallelWorkers()`)
8. Hors v1 : sets d'amorces (`PrimerSets.c`), options `-b/-M/-k`.

Règles du projet à respecter : fichiers < 500 lignes, pas de commit sans demande, pas de
matérialisation complète en mémoire quand le flux suffit, pas de réimplémentation de l'existant.

## 6. Avancement

- [x] Lecture de l'article et des sources C principales
- [x] Inventaire des briques obitools4
- [x] Mesures de référence ecoPrimers (Cyprinidae, Teleostei) : identiques aux références
- [x] Sémantique `-r/-i/-E` testée sur *Salmo*
- [x] `obithermo` (`pkg/obithermo`) : terminé, tests verts, voir section 8
- [ ] Mode A
- [ ] Mode B

## 7. Questions ouvertes

- Mode B : critère « spécifique » strict (T contre C) ou score continu (distance minimale en 3' à C) ?
- Mode B : valeur par défaut de `-x` (0,1 d'ecoPrimers trop permissive pour une cible étroite) ?
- Recherche approchée : l'index par voisinage bat-il agrep ? À mesurer, pas à supposer.

## 8. `pkg/obithermo` (étape 1, terminée)

Fichiers : `model.go` (modèle, ~230 lignes), `nn_table.go` (112 paramètres, **généré** depuis
`nnparams.c` par un script jetable hors dépôt : `/tmp/obithermo-gen/gen.py`), `model_test.go`,
`testdata/c_reference.tsv` (400 cas calculés par la vraie libthermo C compilée),
`testdata/ecoprimers_tm.tsv` (13 amorces et Tm imprimés par ecoPrimers sur Teleostei).

API : `NewModel(primerConc, targetConc, salt, method)`, `DefaultModel()`, `(*Model).SelfTm(primer)`,
`(*Model).Tm(primer, site)` (site = segment de la cible lu dans l'orientation de l'amorce),
`GCCount(seq)`. Retour en degrés Celsius, `NaN` si le modèle ne prédit aucun duplex stable.
Modèle immuable, sans allocation (17,7 ns par appel `Tm`, M3 Max), sûr en concurrence.

Validation :
- 400 cas aléatoires (amorces de 12 à 25 b, 0 à n mismatches, deux jeux de conditions) :
  écart maximal avec le C ~5e-6 K. Cause vérifiée : le C écrit ses constantes en `float`
  (`-21.3f`, `1.987f`, `-5.9f`). En émulant ces arrondis en float32, l'écart tombe à 7e-13 K.
  Le port est donc fidèle ; il utilise des constantes décimales exactes. Tolérance de test : 1e-4 K.
- Les 13 Tm imprimés par ecoPrimers sont retrouvés à +/- 0,05 °C près.

Écarts volontaires avec ecoPrimers (à confirmer) :
1. **273,15 au lieu de 273,0** : ecoPrimers calcule `Tm(K) - 273.0`, `obithermo` utilise la valeur exacte 273,15 (Tm 0,15 °C plus bas ; le test tient compte du décalage). **Validé par l'utilisateur le 2026-11-03 : on garde 273,15.**
2. **Owczarzy corrigé** : dans ecoPrimers, `gcContent` n'est jamais affecté (grep sur tout `src/`),
   donc `-m 2` lit de la mémoire non initialisée (`CNNParams` est sur la pile dans `main`).
   `obithermo` utilise la fraction GC de l'amorce. Non comparable au C ; testé par invariants
   (les deux corrections s'annulent à 1 M, le sel augmente la Tm).
3. **Pas de gaps, bulges ni extrémités pendantes** : jamais atteints par `SelfTM`/`TwoTM` (les
   séquences sont nettoyées en ACGT). Bases non ACGT : erreur, alors que le C lit hors tableau.
4. Tm minimale d'un modèle en échec : `NaN`, là où le C renvoie 0 K (soit -273 °C affiché).

Pour plus tard : les k-mers d'obikmer sont codés A,C,G,T = 0..3, comme les tables de `obithermo` ;
une variante `TmCode(uint64, k)` évitera le décodage quand le cœur appellera Tm par paire.

## 9. Mode A, étape 2 : `obikmer.SeqKmerCounter` (2026-11-03)

**Constat qui change le plan** : `KmerSetGroupBuilder` crée n x P écrivains `.skm` simultanés (un set par
séquence, P >= 64). Pour 3 909 génomes cela fait > 250 000 fichiers ouverts : `KmerSetGroup` +
`QuorumAtLeast` ne convient pas pour « dans combien de séquences ce k-mer apparaît ».

**Solution** (enrichissement d'obikmer, pas de code local à obiprimers) : `pkg/obikmer/seqquorum.go` et
`seqquorum_count.go`.
- Phase 1, `AddSequence(id, seq, circular)` (concurrent) : super-kmers groupés par minimiseur, écrits avec
  l'id de séquence dans **un seul** jeu de P fichiers (format `[uvarint id][uvarint len][2 bits/base]`).
  Les bases non ACGTU découpent la séquence en segments : aucun k-mer ne contient d'ambiguïté (comme
  le masque d'erreur d'ecoPrimers). Génome circulaire : on ajoute les k-1 premières bases.
- Phase 2, `CountAtLeast(q)` : partitions traitées en parallèle (`obidefault.ParallelWorkers()`),
  paires (k-mer canonique, id) triées, séquences distinctes comptées. Résultat en mémoire, trié
  (il est filtré par le quorum : 255 mots pour Teleostei, 1 208 pour Cyprinidae chez ecoPrimers).
- `SeqKmerCount.Repeated` reproduit le drapeau MULTIWORD d'ecoPrimers (k-mer présent plusieurs fois
  dans une même séquence, les deux brins confondus), utile pour `-U` (pas de multi-match).
- Mémoire de la phase 2 : ~16 octets par occurrence de k-mer des partitions en cours de traitement ;
  P doit croître avec la taille totale des données.
- Refactor neutre dans `skm_writer.go` / `skm_reader.go` : `appendPacked` et `unpackBases` extraits et
  partagés avec le nouveau format (tests skm inchangés et verts).

Tests (`seqquorum_test.go`) : oracle force brute indépendant (`EncodeCanonicalKmer` sur chaque fenêtre),
6 configurations (k = 4, 7, 18, 31 ; linéaire/circulaire ; avec N/R ; m et P variés ; insertion
concurrente ; motifs partagés sur les deux brins), plus cas des deux brins dans une même séquence,
ambiguïtés + jonction circulaire, paramètres invalides. `go test ./pkg/obikmer/ -race` vert.
Un bug trouvé par ces tests dans ma boucle de réduction (compteur de séquences) a été corrigé.

Contrôle de bout en bout prévu (étape suivante) : `Strict primer count` d'ecoPrimers = **1 208**
(Cyprinidae) et **255** (Teleostei) avec `-e 3 -3 2 -l 30 -L 150 -c`, k = 18, quorum 0,7 ; le calcul
d'ecoPrimers est `floor(|E| x 0,7)` séquences (> longueur d'amorce).

## 10. Mode A, étape 2b : mots stricts de bout en bout (2026-11-03)

Code : `pkg/obiprimer/roles.go` (`Role`, `Selection`, `BelongingTo`) et `words.go` (`WordSearch`,
`StrictWords`, `FindStrictWords`, `MinSequences`). Tests : `words_test.go`.

- `Selection` reproduit `-r` (restrict), `-i` (ignore) et `-E` (exception) d'ecoPrimers, vérifié sur les
  quatre cas du tableau de la section 4 : exemple = dans Restrict et hors Except ; sinon contre-exemple si
  dans Ignore ou Except ; sinon ignoré. Prédicats construits avec `obiseq.IsSubCladeOf` (existant).
- `FindStrictWords` lit les séquences une fois (consommateurs parallèles via `Split()`), indexe les
  exemples dans un `SeqKmerCounter`, puis `CountAtLeast(floor(exemples x quorum))`. Quorum calculé avec
  une petite tolérance (10 x 0,7 = 7) au lieu de la multiplication en float32 du C.

**Contrôle contre ecoPrimers** (test opt-in : `OBIPRIMERS_TUTORIAL=<dossier du tutoriel> go test ./pkg/obiprimer/`) :

| Jeu | Exemples | Quorum | Mots obiprimers | Mots ecoPrimers |
|---|---|---|---|---|
| Teleostei | 3 909 | 2 736 | **255** | 255 (identique) |
| Cyprinidae | 325 | 227 | **1 207** | 1 208 |

Écart de Cyprinidae : un mot en plus chez ecoPrimers, `TTTATGAAAACAAATAAA`. Comparaison des deux listes
complètes (le journal d'ecoPrimers imprime ses 1 208 mots) : tous les autres mots sont identiques. Un
comptage par force brute (deux brins, jonction circulaire) trouve ce mot dans **225** séquences sur 325,
sous le quorum de 227 : ecoPrimers le garde à tort. Le mécanisme exact n'est pas établi (`merge.c` élague
avec `remainingSeq + count >= quorum` pendant une fusion en place avec file d'attente ; non investigué
plus loin). Impact : ce mot n'apparaît dans **aucune** des 3 448 paires finales de
`Cyprinidae.ecoprimers`. Le test fige 1 207 avec cette explication.

Mesure : étape « mots stricts » seule, 0,5 s (Cyprinidae) et 1,1 s (Teleostei) sur cette machine, lecture
de la taxonomie et du FASTA non comprise ; ecoPrimers complet : 134 s et 350 s (non décomposé par étape).

**Bug existant repéré hors périmètre** : `go test -race` signale une course dans
`obiutils/pipe.go` (`globalLockerCounter` est un `int` incrémenté et décrémenté depuis plusieurs
goroutines, utilisé seulement pour des `log.Debugln`). Sans rapport avec obiprimers ; un `atomic.Int64`
suffisait. **Corrigé** (accord de l'utilisateur) : `globalLockerCounter` est un `atomic.Int64` dans
`pkg/obiutils/pipe.go` ; le test de bout en bout passe désormais sous `-race`.

Suite : étape 3 (recherche approchée des mots stricts, <= e substitutions, deux brins, dans exemples
et contre-exemples), puis paires et indices.

## 11. Mode A, étape 3 : recherche approchée (2026-11-03)

Code : `pkg/obiprimer/corpus.go` (`Corpus`, `LoadCorpus` : exemples + contre-exemples en mémoire, ordre
d'entrée conservé, ignorés comptés seulement, comme ecoPrimers), `match.go` (types `Match`,
`WordMatches`, `MatchParams`, harnais parallèle `matchCorpus` indépendant du nombre de workers),
`match_naive.go` (oracle en Go pur), `match_apat.go` (variante `obiapat`). `FindStrictWords` travaille
maintenant sur un `Corpus`.

**Règle d'appariement** (celle d'ecoPrimers, figée dans `MatchParams`) : substitutions seules ; toute base
hors A, C, G, T (U lu comme T) dans la séquence compte comme une erreur ; toutes les fenêtres à <= e
erreurs sont rapportées, chevauchantes comprises ; fenêtres à cheval sur la jonction si circulaire.
**`obiapat` ne suit pas cette règle de lui-même** : il lit les codes IUPAC dans la séquence (un N
correspond à toutes les bases). La variante `obiapat` remplace donc les symboles ambigus par `z` (rien
ne lui correspond), sur une copie, seulement pour les séquences qui en contiennent.

Vérification de justesse :
- oracle validé à la main sur 12 cas (brins, N, IUPAC, U, jonction, chevauchement, séquence trop courte) ;
- `obiapat` == oracle position par position et nombre d'erreurs compris, 6 configurations (k = 12, 18,
  25 ; e = 0 à 4 ; linéaire/circulaire ; séquences avec n, r, y, k, m ; mots plantés sur les deux brins
  avec 0 à 3 substitutions et à la jonction) ;
- **bout en bout** : pour chaque mot strict, nombre de séquences exemples touchées == celui que
  ecoPrimers a journalisé (`e = 3`) : **1 207 mots Cyprinidae et 255 mots Teleostei, 0 différence**.
  Les nombres d'ecoPrimers sont dans `testdata/ecoprimers_*.words.tsv` ; test opt-in
  `TestMatchApatAgainstEcoPrimers` (`OBIPRIMERS_TUTORIAL`, `OBIPRIMERS_WORKERS=n` pour fixer les cœurs).
  (Cyprinidae : 1 207 des 1 208 mots d'ecoPrimers, le mot en trop de la section 10 est exclu.)

Mesures, `e = 3`, génomes circulaires, M3 Max :

| Jeu | Mots x séquences | Matches | 16 workers | 1 worker | Mémoire ajoutée |
|---|---|---|---|---|---|
| Cyprinidae | 1 207 x 325 (5,4 Mb) | 416 386 | 3,8 s | non mesuré | 53 Mo |
| Teleostei | 255 x 3 909 (65 Mb) | 999 407 | 10,7 s | **113 s** (293 Mb/s par passe mot-brin) | 151 Mo |

ecoPrimers complet (non décomposé par étape) : 134 s et 350 s, 1,9 et 2,1 Go.

Profil monothread (Teleostei, `go tool pprof`) : 74 % dans `runtime.cgocall` (le balayage C), 11 % en
surcoût de transition cgo (`pthread_cond_signal`), tri/GC marginaux. Il y a 2 millions d'appels cgo
(255 mots x 2 brins x 3 909 séquences).

Variantes du protocole :
1. `obiapat` mot par mot, parallélisé par séquence : **retenue comme référence de production** pour l'instant.
2. Arrêt anticipé sous le quorum de 90 % : **écartée, mesuré inutile** (Teleostei : 255/255 mots
   conservés ; Cyprinidae : 1 201/1 208). Le parallélisme par séquence est déjà dans la variante 1.
3. Bit-parallèle multi-mots en Go (SIMD) : **non écrite**. Cible : les ~74 % de balayage C.
4. Index de voisinage / Aho-Corasick : **j'avais écarté cette piste à tort** (voir section 13) : mon calcul
   ne considérait que le découpage en e+1 morceaux (graines de 4 bases). Le découpage en deux moitiés avec
   voisinage à distance floor(e/2) est très sélectif et a gagné.

Décision en suspens : écrire la variante 3 maintenant, ou d'abord finir le pipeline (paires et indices)
et profiler l'ensemble pour savoir si l'appariement domine encore.

## 12. Mode A, étape 4 : paires et indices (2026-11-03)

Code : `taxa.go` (`Taxa`, `Corpus.AssignTaxa`), `pairs.go` (paramètres, amorces conservées, énumération
des amplicons, contrainte du 3', comptage), `pairs_eval.go` (`FindPairs`, `PairStats`, Bc, Bs, Yule, Tm),
test opt-in `TestPairsAgainstEcoPrimers`.

Fidèle à ecoPrimers : amorce conservée = trouvée dans >= floor(|E| x 0,9) exemples ; « good » si en plus
<= floor(|C| x 0,1) contre-exemples ; une paire a au moins une amorce good ; longueur d'amplicon
strictement entre `-l` et `-L` ; clé de paire (amorce 1, sens, amorce 2, sens) ordonnée par mot
canonique ; paire retenue si q > 0,9 et qfp < 0,1 (comparaisons en float32 comme le C) et Bs >= 0,6 ;
Bc = taxons amplifiés / taxons exemples ; **Bs = taxons identifiés / taxons amplifiés par la paire**
(et non / tous les taxons : c'est ce que fait le code, l'article est ambigu) ; un taxon est identifié si
au moins un de ses amplicons n'est partagé avec aucun autre taxon ; tri par Yule x Bs.

Choix d'implémentation (mémoire) : deux phases. Phase 1 ne garde que des compteurs par paire (séquences
exemples/contre-exemples, longueurs) ; phase 2 ne revisite que les paires retenues (407 / 3 448) et
n'enregistre qu'un hash 64 bits d'amplicon par occurrence au lieu de stocker les amplicons. ecoPrimers
conserve un amplicon par occurrence de chaque paire candidate (42 682 paires pour Cyprinidae).

**Contrôle contre les tables complètes d'ecoPrimers** (mêmes paramètres, `-e 3 -3 2 -l 30 -L 150 -c`) :

| Jeu | Paires | Résultat |
|---|---|---|
| Cyprinidae | 3 448 / 3 448 | **identique sur toutes les colonnes** : amorces, good, séquences, taxons, Bc, Bs, longueurs, GC, Tm, Tm minimale |
| Teleostei | 407 / 407 | mêmes paires ; **6 comptent 1 séquence de moins** (NC_033548) |

Cause de l'écart Teleostei **établie par expérience** : l'`extractSite` d'ecoPrimers lit un `N` (code 4)
dans un registre à décalage ; le 4 déborde sur la base voisine et `~4 & 3` vaut 3 (T), donc un `N` en 3'
peut être accepté par accident. En émulant exactement ce registre dans le test du 3' (modification
temporaire, annulée), les **deux tables sont reproduites à 100 %** sur toutes les colonnes. Non conservé :
une base ambiguë ne peut pas être une correspondance 3' vérifiée. Le test fige donc l'écart (6 paires,
-1 séquence, rien d'autre ne diffère). Une première émulation plus grossière (N lu comme A) était
fausse : elle créait un écart dans Cyprinidae.

Autres écarts connus, non testés sur ces données : fenêtres à cheval sur la jonction circulaire, que
le C lit au-delà du tableau (comportement indéfini) alors que nous reprenons au début ; Tm minimale
ignorant les sites contenant une base ambiguë (le C en tire des valeurs arbitraires).

**Profil du pipeline complet (16 workers, M3 Max)**, le critère de l'option A :

| Étape | Cyprinidae | Teleostei |
|---|---|---|
| lecture + sélection | 0,45 s | 0,44 s |
| taxons | 0,00 s | 0,00 s |
| mots stricts | 0,14 s | 1,04 s |
| **recherche approchée** | **4,14 s** | **10,18 s** |
| paires + indices + Tm | 1,04 s | 1,12 s |
| total | ~5,8 s | ~12,8 s |

ecoPrimers complet : 134 s et 350 s. **L'appariement approché reste 72 % à 80 % du total : la variante
SIMD (variante 3) est justifiée par la mesure.** Mémoire : non isolée (le tas de 2,3 à 2,8 Go du test
contient la taxonomie NCBI complète).

Non fait : option `-U` (pas de multi-match), séquence de référence `-R`, sets d'amorces, sortie CSV,
CLI. Prochaine étape : variante 3 (bit-parallèle multi-mots, SIMD).

## 13. Mode A, étape 5 : recherche approchée par index de graines (2026-11-03)

**Correction d'une erreur de ma part.** À la section 11 j'écartais l'index de graines « par le calcul » :
je n'avais considéré que le principe des tiroirs avec e+1 morceaux (graines de 4 à 5 bases, des
centaines de milliers de candidats). C'était la mauvaise configuration. Avec **deux moitiés** de
k/2 bases : une fenêtre à <= e = 2d+1 erreurs a au plus d = floor(e/2) erreurs dans l'une des deux
moitiés. On indexe donc, pour chaque moitié de chaque motif (mot et son complément inverse), la moitié
et ses voisins à distance d : 1 + 3L valeurs pour d = 1 (28 pour L = 9, dans 4^9 = 262 144). Sur du
texte aléatoire cela fait ~0,0002 x W candidats par position et par moitié au lieu de 2W comparaisons.

Code : `pkg/obiprimer/match_seed.go`, fonction exportée **`MatchWords`** (remplace la variante `obiapat`,
supprimée). Détails : fenêtre glissante 2 bits/base ; bases ambiguës suivies par un masque (erreur
sûre, aucune correspondance) ; fenêtres circulaires par extension de k-1 bases ; index direct pour les
moitiés de <= 11 bases, haché sur 22 bits au-delà (k >= 24) ; doublons évités par recherche binaire
dans la liste de l'index gauche ; vérification exacte par popcount sur le code complet.

Justesse : égal à l'oracle `matchNaive` sur 11 configurations (k = 9, 12, 18, 24, 25, 31 ; e = 0 à 5 ;
linéaire/circulaire ; séquences avec n, r, y, k, m ; cas hachés) ; égal à l'oracle sur les 40
premières séquences réelles de chaque jeu ; égal à `obiapat` sur l'intégralité des 999 407
occurrences de Teleostei ; et mêmes nombres de séquences que le journal d'ecoPrimers, mot par mot
(1 207 et 255 mots, 0 différence).

Mesures `e = 3`, M3 Max :

| Jeu | obiapat 1 worker | MatchWords 1 worker | obiapat 16 workers | MatchWords 16 workers |
|---|---|---|---|---|
| Teleostei (255 mots, 65 Mb) | 113 s | **0,42 s** (x270) | 9,4 s | **43 ms** (x220) |
| Cyprinidae (1 207 mots, 5,4 Mb) | non mesuré | non mesuré | 3,8 s | **13 ms** |

Mémoire allouée par l'appariement : 120 Mo (Teleostei), 51 Mo (Cyprinidae), essentiellement le
stockage des occurrences.

**Profil du pipeline complet, 16 workers** (lecture du FASTA gz comprise, taxonomie exclue) :

| Étape | Cyprinidae | Teleostei |
|---|---|---|
| lecture + sélection | 0,46 s | 0,42 s |
| mots stricts | 0,10 s | 1,02 s |
| recherche approchée | 0,014 s | 0,041 s |
| paires + indices + Tm | 1,09 s | 1,20 s |
| total | ~1,7 s | ~2,7 s |

ecoPrimers complet : 134 s et 350 s, soit environ 80 et 130 fois plus. Les résultats complets restent
ceux de la section 12 (Cyprinidae identique sur toutes les colonnes ; Teleostei : mêmes 407 paires,
6 à -1 séquence).

**Décision SIMD** : l'appariement ne représente plus que 1 % du temps, donc un noyau SIMD pour la
recherche n'a plus d'objet. Les étapes restantes (lecture 0,4 s, mots stricts 1 s sur Teleostei,
paires 1,1 s) sont celles à profiler si l'on veut aller plus loin ; le tri des paires (kmer, id)
dans `SeqKmerCounter` et la phase 2 de `FindPairs` en sont les candidats.

Non fait : option `-U`, `-R`, sets d'amorces, mode B, sortie CSV, CLI `obiprimers`.

## 14. Commande `obiprimers`, mode A (2026-11-03)

Code : `pkg/obiprimer/search.go` (`SearchParams`, `DefaultSearchParams`, `Search`, `SearchInfo` ;
`FindPairs` renvoie maintenant aussi un `PairInfo`), `pkg/obitools/obiprimers/{options,obiprimers,output}.go`,
`cmd/obitools/obiprimers/main.go`, test `obiprimers_test.go`. Le `Makefile` découvre la commande tout seul
(`wildcard cmd/obitools/*/*.go`). Pattern : `InputOptionSet` + `OutputModeOptionSet` + `-t` obligatoire,
sortie par `obiitercsv` + `obicsv.CLICSVWriter` (donc `-o`, `--compress`), comme `obitaxonomy`.

Options (noms alignés sur `obipcr` et `obigrep` quand ils existent) :

| Option | ecoPrimers | Défaut |
|---|---|---|
| `-t/--taxonomy` (facultatif depuis la section 16) | `-d` (base) | taxonomie embarquée |
| `-r/--restrict-to-taxon` | `-r` | tous les exemples |
| `-i/--ignore-taxon` | `-i` | |
| `-E/--exception-taxon` | `-E` | |
| `--rank` | `-t` | species |
| `-c/--circular` | `-c` | non |
| `--primer-length` | `-O` | 18 |
| `-e/--allowed-mismatches` | `-e` | 3 |
| `-l/--min-length`, `-L/--max-length` | `-l`, `-L` | 0, 1000 |
| `-3/--three-prime` | `-3` | 0 |
| `-q/--strict-quorum` | `-q` | 0,7 |
| `-s/--sensitivity` | `-s` | 0,9 |
| `-x/--false-positive` | `-x` | 0,1 |
| `-T/--specificity` | `-T` | 0,6 |
| `--salt-correction`, `--salt`, `--primer-concentration` | `-m`, `-a` | santalucia, 0,05 M, 8e-7 M |
| `--partitions` | (nouveau) | 256 |

`-O` n'a pas d'alias court : dans obitools4 il désigne l'en-tête OBI en sortie. Non reprises : `-U`
(pas de multi-match), `-R` (séquence de référence), `-p/-b/-M/-k` (sets et liens), `-f`, `-v`, `-A`.

Sortie CSV (en-têtes nommés, ordre des colonnes d'ecoPrimers) : `index, primer1, primer2, tm1, tm1_min,
tm2, tm2_min, gc1, gc2, good1, good2, examples, counterexamples, yule, example_taxa,
counterexample_taxa, bc, identified_taxa, bs, min_length, max_length, mean_length`. Tm arrondies à
0,1 °C, indices à 4 décimales, `NA` quand une Tm est indisponible. Le résumé (nombres de mots, de paires,
d'exemples...) va dans le journal (stderr), pas dans le CSV : les lignes `#` d'ecoPrimers sont donc
remplacées par des messages `INFO`.

**Contrôle** : `obiprimers -t ncbitaxo_20250211.tgz -r 7953 -e 3 --three-prime 2 -l 30 -L 150 -c
mito.vert.fasta.gz` : 3 448 lignes, comparées à `Cyprinidae.ecoprimers` champ par champ (script jetable) :
**0 différence**. Teleostei : 407 lignes (les 6 écarts d'une séquence de la section 12 s'appliquent).

Mesures du binaire complet (M3 Max, 16 workers, taxonomie NCBI complète comprise) :

| Jeu | Temps | Pic mémoire | ecoPrimers (temps, pic) |
|---|---|---|---|
| Cyprinidae | 7,4 s | 2,82 Go | 134 s, 1,86 Go |
| Teleostei | 8,4 s | 3,42 Go | 350 s, 2,11 Go |

**Le pic de mémoire est PLUS ÉLEVÉ qu'ecoPrimers**, malgré un temps 18 à 40 fois meilleur. Éléments mesurés :
le chargeur de taxonomie NCBI existant coûte à lui seul 1,07 Go (`obitaxonomy`, mesuré) ; avec
`GOGC=25` le pic de Cyprinidae passe de 2,82 à 1,98 Go (donc ~0,85 Go de ramasse-miettes en retard, au
prix de 2x de temps CPU) ; le reste (~0,9 Go au-dessus de la taxonomie pour 5,4 Mb d'exemples) n'est pas
analysé. Pistes non explorées : ne pas parser les 131 Mb du FASTA des séquences ignorées, libérer la
taxonomie après `AssignTaxa`, régler le GC de la commande.

Avertissements `Taxid ... has to be updated` : viennent du mécanisme existant d'obitools (taxids
fusionnés dans le FASTA) ; `--update-taxid` les corrige à la lecture.

Reste à faire : mode B, tests shell `obitests/obitools/obiprimers` (données dans le dépôt),
documentation, réduction de la mémoire, comparaison systématique `--rank` autre que species.
(`-U` et `-R` : faits, section 15.)

## 15. Options `-U` (no multi-match) et `-R` (référence) (2026-11-03)

Code : `pkg/obiprimer/multimatch.go` (`DropRepeatedWords`, `DropMultiMatch`), `reference.go`
(`referenceSites`, `span`), `Corpus.IndexOfExample`, champs `SearchParams.NoMultiMatch/Reference`,
`PairStats.RefLocation/RefAmplicon`, options CLI `-U/--no-multi-match` et `-R/--reference`, colonnes CSV
`reference_location` et `reference_amplicon` (présentes seulement avec `-R`, `NA` pour une paire qui
n'amplifie pas la référence). Tests : `search_test.go` (synthétiques), `search_real_test.go` (opt-in).

### -U

Sémantique lue dans le C (`aproxpattern.c` et `ahocorasick.c`, qui donnent la même règle) : un mot est
écarté s'il a >= 2 occurrences approchées, **deux brins confondus**, dans **une** séquence quelconque du
corpus (exemple ou contre-exemple). Le drapeau « multi-mot » du comptage strict (`Repeated`, mot présent
plusieurs fois dans un exemple) n'en est qu'un cas particulier ; ecoPrimers l'applique d'abord pour
alléger la recherche, nous aussi (`DropRepeatedWords` puis `DropMultiMatch`).

**Contrôle contre ecoPrimers `-U`** (Cyprinidae, 98,7 s, 1,18 Go) : **1 366 paires, identiques sur tous les
champs** (primers, good, séquences, taxons, Bc, Bs, Yule, longueurs, GC, Tm). Notre journal :
34 mots stricts écartés comme répétés, 469 écartés pour multi-match, 703 amorces conservées.
`testdata/ecoprimers_Cyprinidae_U.pairs.tsv` garde les valeurs de référence (test opt-in
`TestNoMultiMatchAgainstEcoPrimers`).

Détail sans conséquence sur les résultats : avec `-e 0`, ecoPrimers (chemin Aho-Corasick) ne filtre pas les
amorces par le quorum de sensibilité et définit « good » par les seuls contre-exemples ; nous appliquons le
même filtre qu'avec `-e > 0`. Une amorce sous le quorum ne peut pas figurer dans une paire retenue
(une paire amplifie au plus autant d'exemples que ses amorces en touchent), donc les tables finales
sont les mêmes ; seuls les décomptes du journal peuvent différer. Non vérifié par une exécution `-e 0`.

### -R

Contrairement à `-U`, **le `-R` d'ecoPrimers est faux** : `refsequence` reçoit le numéro de la séquence
(`pairs.c`) puis sert d'indice dans la liste des amplicons de la paire (`ecoprimer.c`, `printapair`).
Mesuré sur Cyprinidae : avec la référence `NC_022183` ecoPrimers coïncide avec nous sur les coordonnées
des 3 448 paires (coïncidence : l'indice tombe juste) ; avec `NC_083213`, **seuls 382 de ses 3 405
emplacements sont de vrais sites d'amorces dans la séquence, contre 3 405 sur 3 405 pour obiprimers**
(vérifié par un script indépendant et par `TestReferenceSitesAreRealSites`). Nous implémentons donc
l'intention, pas le C :

- `reference_location` : emplacement GenBank 1-based des deux sites, de gauche à droite,
  `join(a..b,c..d)`, entouré de `complement()` quand `primer1` (tel qu'écrit dans le CSV) se fixe sur le brin
  moins. Fenêtre à cheval sur la jonction circulaire : coupée en deux (`a..n,1..m`).
- `reference_amplicon` : région entre les sites, amorces exclues, en minuscules, lue depuis `primer1`.
- **Orientation** : ecoPrimers oriente l'amplicon par rapport à l'ordre canonique des mots, nous par rapport
  à `primer1` du CSV (plus lisible). Sur NC_022183 : coordonnées identiques pour les 3 448 paires,
  amplicon identique pour 2 572 et complémenté-inversé pour 836, tout le reste égal.
- Plusieurs amplicons de la paire sur la référence : on garde le plus à gauche (le C garde le dernier,
  par écrasement, mais ses valeurs sont de toute façon indéfinies).
- Référence introuvable ou contre-exemple : **erreur** (le C ignore silencieusement).
- Vérifications : 3 408 paires sur NC_022183 et 3 405 sur NC_083213, sites à <= 3 erreurs, amplicon égal à
  la séquence (ou son complément inverse), orientation cohérente (2 184 paires `primer1` sur le brin +,
  1 224 sur le brin -).

### Autres constats

- `obiutils` : `TestSetString` est **instable** (échoue ~4 fois sur 6) : `Set.String()` parcourt une map Go
  et le test attend `[1 2 3]`. Sans rapport avec obiprimers, non corrigé.
- Un défaut de mon test précédent (`TestCLISearchParams` laissait `_salt = -1`) a été corrigé.

## 16. Taxonomie embarquée dans les séquences (2026-11-03)

Demande : comme `obitag`, `obiprimers` doit pouvoir se passer de `-t` quand les séquences ont été annotées
avec `obiannotate --taxonomic-path` (attribut `taxonomic_path` : liste « code:taxid [nom]@rang » de la racine
au taxon).

Code : `pkg/obitools/obiprimers/obiprimers.go` (`readSequencesAndTaxonomy`, `embeddedTaxonomyInMemory`,
`CLIReadSequencesAndTaxonomy`), `-t` rendu facultatif (`LoadTaxonomyOptionSet(options, false, false)`),
`CLIPrimerPairs(sequences, taxonomy)` reçoit désormais la taxonomie. Réutilise l'existant :
`IBioSequence.ExtractTaxonomy` et `BioSequenceSlice.ExtractTaxonomy` (le mécanisme d'`obitag`).

Règles : `-t` donné (taxonomie par défaut déjà chargée) : il l'emporte, les chemins embarqués sont ignorés ;
sinon la taxonomie est extraite des séquences et définie comme taxonomie par défaut, comme `obitag`.
- **Fichiers** : deux passes. La première ne retient que les chemins taxonomiques (mémoire = taxonomie
  seule), la seconde lit le corpus. Une seule passe obligerait à garder toutes les séquences en mémoire,
  puisque la sélection des exemples (`-r`, `-i`, `-E`) a besoin de la taxonomie.
- **Entrée standard** (non relisible) : séquences gardées en mémoire (`Load()`), taxonomie extraite, puis
  réémises par `IBatchOver`.
- **Pas de chemin** : erreur « no taxonomy: use the -t option, or give sequences annotated with their
  taxonomic path (obiannotate --taxonomic-path): path is empty ». L'extraction est stricte, comme celle
  d'`obitag` : une seule séquence sans chemin fait échouer l'extraction.
- `--with-leaves` (séquence = feuille de son taxon), acceptée par le jeu d'options partagé de taxonomie,
  est **sans effet** sur obiprimers.

**Contrôle** (fichier `mito.vert.path.fasta.gz` produit par `obiannotate -t ... --taxonomic-path`, 6,3 s) :

| Cas | Résultat |
|---|---|
| sans `-t`, fichier | CSV **identique octet pour octet** (`cmp`) au CSV obtenu avec `-t` (3 448 paires) |
| sans `-t`, entrée standard | CSV identique |
| `-t` + fichier annoté | CSV identique, pas d'extraction (le `-t` l'emporte) |
| sans `-t`, fichier non annoté | erreur explicite, code de sortie 1 |

La taxonomie embarquée ne contient que 13 039 taxons (ceux des chemins) contre 2,6 millions de nœuds pour
la taxonomie NCBI complète. **Effet sur la mémoire et le temps (Cyprinidae)** :

| | avec `-t` NCBI | taxonomie embarquée (fichier) | embarquée (stdin) |
|---|---|---|---|
| temps | 7,4 s | **2,65 s** | 1,27 s (lecture depuis `gzcat`) |
| pic mémoire | 2,82 Go | **1,50 Go** | 1,23 Go |

(le fichier annoté pèse 43,9 Mo au lieu de 39,0 Mo en gzip.) Ce gain vient du chargeur de taxonomie NCBI
évité ; il ne change rien à l'algorithme.

Tests (`embedded_test.go`) : extraction depuis des fichiers avec restriction `-r` sur un taxid de la
taxonomie embarquée et attribution des espèces, extraction en mémoire, erreur avec conseil, précédence de
`-t`, taxonomie nulle refusée.

**Correctif de code partagé** : en testant sous `-race` la lecture des fichiers, le détecteur a signalé une
course dans `obiiter.IBioSequence.RebatchBySize` (appelée par `CLIReadBioSequences`) : la goroutine
réaffectait la variable capturée `iterator` (`iterator = iterator.SortBatches()`) pendant que la fonction
appelante lisait `iterator.IsPaired()`. Corrigé avec une variable locale (`sorted`), comportement inchangé ;
les tests de `obiiter`, `obiformats` et `obitools/...` passent.

### 16 bis. Suppression de `taxonomic_path` après extraction (2026-11-03, demande de l'utilisateur)

Une fois la taxonomie construite, l'attribut `taxonomic_path` (liste d'une vingtaine de chaînes par
séquence) ne sert plus : la recherche lit les taxons par le `taxid`. Il est supprimé dans les trois
chemins de lecture (`obiprimers.go`) : fichiers (worker `obiseq.AnnotatorToSeqWorker` sur la seconde
passe), entrée standard (en place sur la tranche gardée en mémoire, avant de la réémettre) et `-t` donné
avec un fichier annoté (les chemins sont inutiles). Réutilisation : `AnnotatorToSeqWorker` et
`BioSequence.DeleteAttribute` existants ; seul ajout partagé : la constante
`obiseq.TaxonomicPathAttribute`, utilisée par `SetPath`/`Path` à la place du littéral. Test :
`TestTaxonomicPathIsDroppedAfterExtraction` (le chemin disparaît, le `taxid` reste).

**Mesures, sans gain visible sur le pic de mémoire du processus** :
- tas réellement libéré (7 822 séquences, 131 Mb de bases, `GC` forcé, tranche gardée vivante) :
  **19,2 Mo sur 193,7 Mo (9,9 %), 2,46 Ko par séquence**. Un premier essai donnait « 99,9 % » : il était faux
  (la tranche n'était plus référencée, donc toutes les séquences avaient été libérées) ;
- pic de mémoire du processus (entrée standard, 8 mesures entrelacées chacune) : médiane **1 334 Mo avant,
  1 346 Mo après**, dispersion > 250 Mo : aucune différence détectable.

Interprétation : le gain est réel mais de l'ordre de 10 % des séquences gardées, soit ~1 % du pic ici ; il
deviendrait important sur de très grandes bases (à 2,46 Ko par séquence, ~2,5 Go pour un million de
séquences). **Le pic de 1,3 Go est sans commune mesure avec les 194 Mo de séquences gardées** : l'essentiel
du pic vient d'ailleurs (lecture/analyse des 131 Mb du FASTA, retard du ramasse-miettes : `GOGC=25` avait
fait passer le pic de 2,82 à 1,98 Go). C'est la piste à analyser pour la mémoire.

## 17. Validation avec contre-exemples : `-r 7953 -i 1` (2026-11-03)

Premier cas **avec contre-exemples** comparé à ecoPrimers (la section 1 de l'inventaire : jusque-là tous les
jeux de référence avaient 0 contre-exemple). Les 325 Cyprinidae sont les exemples (`-r 7953`) et toutes les
autres séquences de la base (7 497, 7 387 espèces) sont des contre-exemples (`-i 1`, la racine). Paramètres :
`-e 3 -3 2 -l 30 -L 150 -c`. (« -I 1 » demandé : ecoPrimers n'a pas d'option `-I` ; sa boucle `getopt`
ignore silencieusement une option inconnue, donc c'était `-i`.)

| | ecoPrimers | obiprimers |
|---|---|---|
| exemples / contre-exemples | 325 (316 esp.) / 7 497 (7 387 esp.) | identique |
| quorum de contre-exemples | 749 / 7 497 | identique |
| paires retenues | **173** | **173** |
| temps | **2 728 s (45,5 min)**, 1 cœur | **1,9 s**, 16 workers |
| pic mémoire | 756 Mo | 781 Mo (taxonomie embarquée) |

**Les deux tables sont identiques sur tous les champs** (173/173 paires, 0 différence : amorces, drapeaux
good, séquences et taxons amplifiés côté exemples ET contre-exemples, Yule, Bc, Bs, longueurs, GC, Tm).
Répartition des drapeaux good : 97 `bG` et 76 `Gb` (une seule amorce good par paire, comme attendu avec
des contre-exemples aussi nombreux) ; contre-exemples amplifiés par paire : de 264 à 513 séquences.
Notre journal : 6 850 691 occurrences approchées (contre 227 549 sans contre-exemples), 1 200 amorces
conservées dont **14 good**, 2 037 paires candidates, 200 amplifiant assez d'exemples, 173 retenues
(ecoPrimers : « Total pair count 2050 », écart de 13 attendu : il a un mot strict de plus, section 10).

Fixé dans `testdata/ecoprimers_Cyprinidae_I1.pairs.tsv` ; test opt-in `TestCounterExamplesAgainstEcoPrimers`.
Cela couvre désormais : quorum de faux positifs, drapeau good avec contre-exemples, comptes de taxons
contre-exemples. Reste à comparer (inventaire section 1) : séquences linéaires, `--rank` autre qu'espèce,
`-e 0`, autres `-e` / `-O` / quorums, et `-E` (exception).

## 18. Validation par variantes d'options, Cyprinidae `-r 7953` (2026-11-03)

16 exécutions d'ecoPrimers lancées en parallèle (1 cœur chacune, 2 à 6 minutes sans contre-exemples),
comparées ligne à ligne (multi-ensembles de lignes, car des paires d'amorces AT répétées existent des deux
côtés) à `obiprimers` avec les options équivalentes. Scripts jetables : `/tmp/run_val*.sh`,
`/tmp/run_mine.sh`, `/tmp/compare_val.py` (non versionnés).

**Identiques sur toutes les lignes et tous les champs** :

| Variante | Lignes |
|---|---|
| séquences linéaires (sans `-c`) | 3 448 |
| `-e 1` | 2 031 |
| `-e 5` | 4 080 |
| `-O 20` | 2 288 |
| `-O 25` | 867 |
| `-q 0.8 -s 0.95` | 1 110 |
| `-q 0.5 -s 0.8` | 13 941 |
| `-3 0` | 5 405 |
| `-3 4` | 2 262 |
| `-l 10 -L 80` | 223 |
| `-E 7954` (Danio, 5 contre-exemples hors des exemples) | 11 |
| `-E 7954 -x 0.05 -s 0.8` | 16 |
| `-E 84644 -x 0.95` (Labeo : 304 exemples, 21 contre-exemples dans les exemples) | 54 |

`-E 84644` seul et `-E 84644 -x 0.05 -s 0.8` donnent 0 paire des deux côtés : identiques mais sans valeur,
les Labeo étant amplifiés par toutes les amorces conservées ; d'où la variante `-x 0.95`.

**`--rank genus` et `-T 0.3` : différents, mais entièrement expliqués** (par mesure) :
1. **118 et 119 paires seulement chez ecoPrimers** : toutes contiennent le mot strict en trop d'ecoPrimers
   `TTTATGAAAACAAATAAA` (section 10 : présent dans 225 séquences, quorum 227). Aucune des paires absentes
   ne l'ignore ; aucune ligne des variantes identiques ne le contient. Leurs comptes d'exemples vont de 294
   à 325 : l'hypothèse « effet de seuil à 293 » avait été testée et réfutée.
2. **52 et 65 lignes avec une séquence d'exemple de plus chez ecoPrimers** (exactement +1 exemple, +0
   contre-exemple à chaque fois) : l'artefact du `N` dans la vérification du 3' (section 12).
3. **9 écarts de Tm minimale** (mêmes comptes de séquences) : ecoPrimers décode les sites contenant une
   base ambiguë avec le même registre à décalage ; nous ignorons ces sites (section 12).
Preuve : un binaire temporaire émulant exactement `extractSite` (3' et sites de Tm), sources restaurées
ensuite (diff identique aux sauvegardes), supprime **toutes** les différences de champs ; ne restent que
les 118/119 paires du mot en trop.

## 19. Décisions de l'utilisateur et options `-b`, `-M`, `-k` (2026-11-03)

- **Taxonomie embarquée : extraction stricte** (comme `obitag`) : comportement actuel conservé, rien à changer.
- **Test instable corrigé** : `obiutils.Set.String()` parcourait une map Go (ordre aléatoire) ; elle trie
  désormais ses membres (entiers, flottants et chaînes par valeur, y compris types nommés ; repli sur la forme
  texte pour les autres types). `Members()` est inchangé. Test `TestSetStringIsStable` ; `TestSetString` passe
  200 fois de suite (il échouait ~4 fois sur 6) et tout le paquet passe.
- **`-b`, `-M`, `-k`, `-p`** : lus dans `PrimerSets.c` (voir ci-dessous) ; porter ou abandonner : à trancher.

### Ce que sont ces options dans ecoPrimers (lu dans le code)

Un « lien » entre deux paires = le nombre de séquences que **les deux** paires identifient sans ambiguïté
(`wellIdentifiedSeqs[k] == 1` pour les deux). Le but est de choisir des paires complémentaires (plusieurs
marqueurs) : peu de liens = les paires identifient des taxons différents.
- `-b` : filtre la liste finale des paires avec `primers_changeSortedArray` (puis réimprime). Une paire est
  gardée s'il existe une autre paire avec laquelle elle a au plus 5 liens ou plus de la moitié du plus petit
  nombre de taxons identifiés ; elle n'est écartée que si, avec **toutes** les autres, ses liens sont entre 6
  et cette moitié. Règle heuristique sans justification dans le code.
- `-M` : seuil absolu de liens (nom trompeur : « percent » mais utilisé comme un nombre). Quand il est donné,
  la boucle contient un `break` inconditionnel : chaque paire n'est comparée qu'à la **première** autre paire
  de la liste, donc le filtre ne fait pas ce qu'il annonce.
- `-k` : n'est lu que par `ok_to_add` (construction d'ensembles d'amorces), inatteignable puisque `-p` est
  désactivé : option morte.
- `-p` : sa ligne d'affectation est commentée dans `ecoprimer.c` ; elle ne fait rien. Le code d'ensembles
  associé (glouton, recuit simulé, tabou, force brute) n'est jamais exécuté.
- `primers_filterWithGivenLinks` et `primers_graph_graphviz` ne sont jamais appelées.
Dépendance technique d'un éventuel portage : il faudrait garder, pour chaque paire retenue, l'ensemble des
séquences qu'elle identifie (aujourd'hui seule leur nombre de taxons est conservé).

## 20. Le problème multi-marqueurs : spécification (discussion du 2026-11-03)

**Statut : spécification en cours, aucun code écrit.** Décisions de l'utilisateur ci-dessous ; les points
marqués « proposition » ou « question ouverte » sont de moi et n'ont pas été validés.

### 20.1 Origine de `-b`, `-M`, `-k`, `-p` : abandonnées

Ces options sont des pistes explorées pendant la thèse de Tiayyba Riaz (auteur de l'article), il y a une
quinzaine d'années, jamais menées à terme : le problème était mal posé. L'utilisateur avait oublié qu'elles
traînaient dans le code. **Décision : elles ne sont pas portées** (confirmant l'analyse de la section 19 : deux
options mortes, une buguée, un filtre heuristique non justifié).

Le modèle de la thèse : chaque paire d'amorces couvre et reconnaît une partie des espèces, avec un
recouvrement partiel entre paires ; on cherche la meilleure combinaison (graphe de « liens », glouton, recuit
simulé, tabou, force brute). Ce modèle est quasi inapplicable : un recouvrement d'espèces entre deux paires
visant des clades différents signifie qu'une amorce est partagée entre ces clades, donc que le recouvrement
est presque toujours **complet**. Les ensembles de couverture sont emboîtés ou disjoints par clade.

### 20.2 Formulation correcte

But : un ensemble de marqueurs (un multiplex) qui, ensemble, discriminent un maximum d'espèces.
Structure :
1. **une paire de base** qui couvre la totalité du groupe visé (Bc proche de 1) ;
2. **une collection de paires spécialisées**, chacune sur un clade que les paires précédentes résolvent mal.

### 20.3 Réponses de l'utilisateur

1. **Granularité des clades à raffiner : elle dépend du nombre de paires autorisées dans le multiplex.**
   Avec 2 paires : une paire générique + une seule paire spécialisée, placée assez haut dans la taxonomie,
   autour d'un gros groupe très mal résolu (celle qui récupère le plus d'espèces). Avec 30 paires : on peut
   éclater presque chaque famille ou chaque genre mal résolu.
2. **Éviter d'amplifier les autres clades : pas par défaut, mais une option.** Elle permet de concentrer
   l'effort d'amplification sur le clade d'intérêt sans le disperser, surtout s'il est sous-représenté dans la
   matrice d'ADN.
3. **Critère de résolution : une espèce est résolue si au moins un marqueur la résout.**
4. **Budget : deux formulations possibles, non tranchées** (maximiser les espèces résolues avec au plus K
   paires ; ou minimiser le nombre de paires pour atteindre un taux donné).
5. **La paire de base est choisie d'abord**, ce qui élimine les espèces bien reconnues ; on raisonne ensuite
   sur ce qui reste (glouton hiérarchique).

### 20.4 Ce que cela implique (propositions, à valider)

- **Algorithme glouton hiérarchique** : paire de base sur tout le groupe (tri Bc puis Bs, comme le mode A) ;
  puis, tant que le budget le permet, parmi les nœuds de la taxonomie contenant des espèces non résolues,
  choisir la paire spécialisée au **plus grand gain marginal** (espèces nouvellement résolues). La
  granularité exigée en 20.3-1 en découle sans règle séparée : un petit budget favorise les nœuds hauts, un
  grand budget finit par les genres. Les deux formulations du budget (20.3-4) sont le même gain marginal avec
  deux conditions d'arrêt (K paires atteint, ou taux visé atteint) : le choix n'a pas à être tranché pour
  écrire la boucle.
- **Réutilisation** : le mode A avec `-r <clade>` produit déjà les paires conservées dans un clade et leur Bs.
  Le mode A étant rapide (de 2 à 3 s pour Teleostei), faire une recherche par clade candidat est envisageable ;
  le corpus devrait alors être chargé une seule fois et la sélection exemples/contre-exemples refaite en
  mémoire pour chaque clade (aujourd'hui `LoadCorpus` relit les fichiers).
- **Manque 1 : les ensembles d'espèces résolues.** Aujourd'hui `PairStats` ne garde que des nombres
  (`Identified`, `ExampleTaxa`). Il faut, pour chaque paire, l'ensemble des taxons identifiés et les classes
  de taxons partageant un amplicon (même besoin qu'un portage de `-b`).
- **Question ouverte (moi) : relativement à quel ensemble une espèce est-elle « résolue » par un spécialiste
  qui peut amplifier d'autres clades (20.3-2) ?** Le Bs actuel ne juge l'unicité de l'amplicon que parmi les
  exemples. Si le spécialiste amplifie aussi des espèces extérieures au clade, un amplicon partagé avec un
  extérieur ne permet pas d'identifier l'espèce dans un échantillon réel. Je propose de juger l'unicité sur
  **tout ce que la paire amplifie dans la base de référence**. Cela demande un troisième rôle de séquence,
  « contexte » : non exigée amplifiée, non pénalisée, mais comptée dans le partage d'amplicons.
- **Lien avec la voie B** : avec l'option « éviter les autres clades » (20.3-2), les voisins deviennent des
  contre-exemples ; sans elle, ils sont du contexte. La voie B (amorces spécifiques d'une espèce ou de
  quelques espèces proches) et l'étape de spécialisation ressemblent donc beaucoup : je propose de les
  spécifier ensemble, comme un seul mécanisme avec une option d'exclusion.
- **Non abordé, à demander** : compatibilité des amorces d'un même multiplex (Tm voisines, absence de
  dimères entre amorces de paires différentes). `obithermo` fournit déjà les Tm.
- **Sortie** : un ensemble ordonné de paires avec, pour chacune, les espèces qu'elle ajoute (gain marginal) ;
  la courbe cumulée des espèces résolues en fonction du nombre de paires répond aux deux formulations du
  budget.

### 20.5 Prochaines étapes

Pas de code avant accord sur 20.4. À trancher : (a) le jugement d'unicité (tout ce qui est amplifié, ou
seulement le clade), donc le rôle « contexte » ; (b) si les voies B et de spécialisation sont spécifiées
ensemble ; (c) les contraintes de multiplex. Le reste de l'inventaire du mode A (validation `-e 0`, nettoyage,
mémoire, tests shell, documentation) est indépendant et peut avancer.

## 21. Suite de la spécification : résolution, voie B, optimisation de Tm (2026-11-03)

**Statut : spécification, aucun code.** « Décision » = réponse de l'utilisateur ; « Proposition » et
« Question » = de moi, non validées.

### 21.1 Décisions

1. **Une espèce est résolue par rapport à tout ce qui est amplifié**, dans le clade cible **et** dans le clade
   voisin non cible. Confirme le jugement d'unicité sur tout ce que la paire amplifie dans la référence.
2. **La voie B doit être spécifiée complètement** avant tout code (et avec la spécialisation, 20.4).
3. **Tm** : idéalement oui, des Tm voisines entre amorces. Aujourd'hui la longueur est fixée et la Tm subie.
   Idée de l'utilisateur : **fixer la Tm et optimiser la longueur à la fin pour équilibrer les Tm**, et cela
   **pour n'importe quelle paire**. Ajouter à `obiprimers` une **phase d'optimisation de Tm** avec deux
   paramètres : **un écart de Tm maximal entre les deux amorces** et **une Tm cible pour les paires**.

### 21.2 Conséquence de 21.1-1 : quatre rôles de séquences (proposition)

| Rôle | Exigence sur l'amplification | Compte dans l'unicité des amplicons |
|---|---|---|
| Exemple | doit être amplifiée (quorum de sensibilité) | oui |
| Contre-exemple | ne doit pas l'être (quorum de faux positifs) | **oui, si amplifiée** |
| Contexte | libre, aucune pénalité | **oui, si amplifiée** |
| Ignorée | sans objet | non |

Le rôle « contexte » est nouveau (aujourd'hui : exemple, contre-exemple, ignorée). L'option « éviter les autres
clades » (20.3-2) transforme simplement leurs séquences de contexte en contre-exemples.
**Écart avec ecoPrimers à noter** : le Bs d'ecoPrimers ne juge l'unicité que parmi les exemples, même quand des
contre-exemples sont amplifiés. Proposition : **garder le Bs d'ecoPrimers inchangé** (c'est ce qui reproduit
les tables de référence) et ajouter un second indice calculé sur tout ce qui est amplifié (nom à choisir,
p. ex. `bs_all`), qui sert à la sélection multi-marqueurs ; il coïncide avec `bs` quand il n'y a ni
contre-exemple amplifié ni contexte.

### 21.3 Optimisation de Tm (proposition, paramètres de l'utilisateur)

- **Paramètres** : `Tm cible` et `écart de Tm maximal`. Proposition d'interprétation : la cible porte sur la
  **moyenne des deux amorces** et l'écart borne leur différence, ce qui revient à imposer à chaque amorce
  `cible ± écart/2`.
- **Où** : phase **après** la sélection des paires, sur chaque paire retenue : choisir les longueurs
  `(L1, L2)` des deux amorces satisfaisant la cible et l'écart.
- **Ce qui rend la phase possible** : raccourcir une amorce **par son extrémité 5'** ne peut que conserver ou
  réduire ses erreurs sur chaque site (c'est une sous-chaîne), donc la sensibilité est conservée. Allonger
  par le 5' demande de **revérifier** les sites. Raccourcir fait en revanche perdre de la spécificité
  (l'amorce raccourcie peut s'apparier ailleurs), d'où la nécessité de **réévaluer la paire** (sensibilité,
  faux positifs, Bc, Bs, Tm) après chaque changement de longueur.
- **Réutilisation** : `obithermo` (Tm, Tm minimale), `MatchWords` et `FindPairs` sur une poignée de mots :
  le coût est celui d'une évaluation par variante de longueur et par paire.
- **Intérêt pour le multiplex** : une Tm cible commune à toutes les paires d'un ensemble les rend compatibles
  par construction.
- **Questions** : (a) la cible porte-t-elle sur la Tm de l'appariement parfait ou sur la Tm minimale sur les
  sites (la pire), ou sur une valeur intermédiaire ? (b) bornes de longueur [Lmin, Lmax] ? Le mot actuel est
  limité à 31 bases (codage 2 bits sur 64). (c) extension et raccourcissement seulement par le 5', en gardant
  l'ancrage 3' ? (d) « pour n'importe quelle paire » : phase intégrée à la recherche, ou aussi outil
  indépendant appliqué à des amorces fournies par l'utilisateur ? (e) une paire dont la Tm ne peut pas être
  satisfaite dans les bornes : écartée ou signalée ?

### 21.4 Voie B : esquisse de spécification (proposition)

**But.** Une paire qui amplifie les cibles T (une espèce, ou quelques espèces très proches) et, autant que
possible, aucun de leurs voisins non cibles C. Facteur de conception : l'espace de recherche est élagué en
se concentrant sur les régions **non conservées** (celles où T diffère de C), l'inverse du mode A.

**Rôles.** T = exemples ; C = contre-exemples (voisins à ne pas amplifier) ; le reste, contexte ou ignoré.

**Constat.** Le pipeline actuel fait déjà fonctionner un cas voie B de façon naïve (`-r 8028 -i 8015`, *Salmo*
contre les autres Salmonidae : 7 exemples, 76 contre-exemples, 15 534 mots stricts), mais sans élagage, et le
critère « un quorum de faux positifs de 10 % » y est trop permissif pour une amorce spécifique.

**Principe d'une paire spécifique.** Une paire n'amplifie un voisin que si **ses deux amorces** s'y apparient.
Il suffit donc d'**une amorce spécifique** (loin de tous les sites de C) et d'une **amorce partenaire**
quelconque et conservée dans T, à distance d'amplicon permise. Les candidats se génèrent ainsi : mots de T
(quorum de T) dont aucun site à <= `e_exclusion` erreurs n'existe dans C (au-delà du quorum de faux
positifs) ; puis le partenaire dans la fenêtre d'amplicon de chaque occurrence dans T.

**Deux tolérances (proposition).** `-e` (sensibilité sur T) et une tolérance d'exclusion **`e_exclusion`** sur C,
éventuellement plus grande : un voisin est « amplifié » s'il s'apparie à `e_exclusion` erreurs ou moins. Cela
donne une marge de sécurité sans dégrader la sensibilité sur T.

**Score de marge (proposition).** Pour une amorce : nombre minimal d'erreurs d'un site de C ; pour la paire :
la meilleure des deux amorces (un voisin n'est amplifié que si les deux s'apparient), avec un poids
particulier pour les erreurs proches du 3'. Il compléterait Yule x Bs pour classer.

**Sortie.** Colonnes du mode A plus la marge des deux amorces.

**Échelle (à prévoir).** Pour T = un génome, le nombre de mots candidats est celui des positions du génome
(16 000 pour une mitochondrie, des millions pour une bactérie) : l'index de graines (`MatchWords`) indexe les
**mots** et balaie le texte de C ; pour des millions de mots il faudra les traiter par lots, chacun balayant C
une fois. À dimensionner selon l'usage visé.

**Questions pour finaliser.**
1. **Exclusion stricte ou marge** : une paire est-elle admise seulement si elle n'amplifie aucun voisin à
   `e_exclusion` erreurs, ou classe-t-on par marge sans seuil dur ? Faut-il une contrainte particulière sur
   les 3' derniers nucléotides des voisins (au moins une erreur dans les 5 dernières bases) ?
2. **Tailles typiques** de T et de C (une espèce avec combien de séquences ? voisins = un genre, une famille,
   toute la base de 8 000 génomes mitochondriaux, des génomes entiers de bactéries ?).
3. **Type de séquences** : génomes organellaires entiers (comme le tutoriel), loci courts, génomes complets ?
4. **Plusieurs cibles** (quelques espèces proches) : amplifier le groupe sans les distinguer (détection de
   groupe), ou les amplifier et les **résoudre** entre elles (Bs au sein de T, comme la spécialisation) ?
5. **Une amorce spécifique + une générale** est-il acceptable, ou les deux amorces doivent-elles être
   spécifiques ?
6. **Sensibilité sur T** : quorum comme en mode A (0,9), ou toutes les séquences de T (intra-spécifique) ?

### 21.5 Travaux préalables communs (sans les engager)

Rôle « contexte » ; indice de résolution sur tout ce qui est amplifié ; ensembles d'espèces résolues par
paire ; corpus chargé une seule fois et sélection refaite en mémoire ; réévaluation d'une paire pour des
longueurs d'amorces variables ; marge d'une amorce (erreurs minimales sur C).

## 22. Réponses de l'utilisateur sur la résolution et la Tm (2026-11-03)

**Statut : spécification, aucun code.** La voie B attend les réponses de l'utilisateur (« encore un peu »).

### 22.1 Redéfinir Bs : oui (décision de principe, détails à confirmer)

Question de l'utilisateur : ecoPrimers calcule Bs sur les seules séquences amplifiées parce que ce sont les
seules chargées en mémoire ; ne vaudrait-il pas mieux **charger tout le monde, chercher les amorces
seulement dans une sélection, et tester sur tout le monde** ? Réponse : **oui**, et cela redéfinit Bs. Trois
populations distinctes :

| Population | Sert à |
|---|---|
| exemples (sélection de recherche) | générer les mots conservés ; sensibilité ; dénominateur de Bc et de Bs |
| contre-exemples | quorum de faux positifs |
| **tout le monde** (population d'évaluation) | juger l'unicité des amplicons : une espèce exemple est résolue si son amplicon n'est partagé avec aucun autre taxon amplifié, où qu'il soit dans la base |

Proposition de définition : `Bs = espèces exemples dont l'amplicon est unique parmi tous les taxons amplifiés de
la population d'évaluation / espèces exemples amplifiées`. Elle coïncide avec le Bs d'ecoPrimers quand la
population d'évaluation est la sélection (aucun contre-exemple ni contexte amplifié). **Piège à ne pas
oublier** : avec `-i 1` (validation de la section 17) le Bs d'ecoPrimers ignore les amplicons des
contre-exemples amplifiés ; la nouvelle définition changera donc ces valeurs. Pour garder les tests de
non-régression, proposition : conserver aussi le Bs d'ecoPrimers dans une colonne à part (p. ex.
`bs_examples`) et définir `bs` comme le nouvel indice ; à confirmer. Colonne informative à ajouter :
nombre de taxons hors exemples amplifiés.

**Mémoire (proposition)** : charger tout le monde n'est pas nécessaire pour un passage simple. Après la
recherche (sur la sélection, en mémoire), un **passage d'évaluation en flux** relit la base entière avec les
seules amorces des paires retenues (quelques centaines de mots : `MatchWords` en 0,04 s pour 65 Mb) et ne
conserve que, par paire, un hash d'amplicon et un taxon par séquence amplifiée, jamais les séquences. Coût :
une relecture du fichier (comme la double lecture de la taxonomie embarquée). Pour la sélection multi-marqueurs
(évaluations répétées par clade), une population d'évaluation **en mémoire sous forme compacte** (2 bits par
base, soit ~33 Mo pour les 131 Mb du tutoriel) serait préférable. À trancher à la spécification de la
sélection multi-marqueurs.

### 22.2 Longueur des amorces : 61 nucléotides (question de l'utilisateur)

L'utilisateur rappelle qu'agrep avait été réimplémenté en 64 bits pour des amorces jusqu'à 61 nucléotides
(`MAX_PAT_LEN` vaut 64 dans `obiapat/apat.h`). État **actuel** de obiprimers : **31 bases au maximum**, limite
inscrite dans `match.go:48`, `words.go:44`, `pairs.go:31`, dans `obikmer` (`encodekmer.go`,
`seqquorum.go:63` : k-mers sur 64 bits, 2 bits par base) et par des tampons fixes `[32]byte`
(`pairs_eval.go:208`). `obithermo` n'a pas de limite de longueur (tranches).

La stratégie de l'index de graines **n'est pas bloquée** : elle se généralise à 61 bases (et 64) avec une fenêtre
de **128 bits** (deux mots de 64 bits, décalage et popcount sur les deux) ; les moitiés (<= 31 bases) tiennent
dans un mot et le chemin d'index haché pour les grandes graines est déjà utilisé et testé (k = 25 et 31).
L'oracle de test existant (`matchNaive`) fixe la justesse. Le compteur de mots stricts (k-mers sur 64 bits,
type partagé d'`obikmer`) resterait limité à 31 : **proposition** : chercher les mots conservés avec un k <= 31
(le « coeur » conservé), puis obtenir les amorces plus longues par **extension** pendant la phase de Tm
(22.3), vérifiée par l'appariement généralisé.

### 22.3 Optimisation de Tm : réponses aux questions de 21.3

(a) **Cible sur l'apparié non parfait** (la Tm de l'amorce sur les sites réels, avec leurs erreurs), pas sur
l'appariement parfait. *Question restante* : quelle statistique des Tm de sites ? Le minimum (`tm_min` d'ecoPrimers)
est dominé par des valeurs aberrantes : sur les tables de référence on lit des minimums de 12 à 45 °C pour des
amorces dont la Tm parfaite est de 52 à 62 °C ; une cible sur le minimum imposerait des longueurs absurdes.
Proposition : cible sur la **moyenne ou la médiane** des Tm de sites (éventuellement pondérées par taxon), le
minimum restant rapporté ; ou un quantile bas réglable. À choisir.
(b) bornes de longueur : traitées en 22.2 (jusqu'à 61 bases possible avec la généralisation).
(c) **Extension et raccourcissement du côté le mieux conservé** (et non seulement par le 5', contrairement à ma
proposition). Cela se calcule avec un profil de conservation par position (fraction des sites cibles qui
portent la même base, comme le PSSM du tutoriel) autour de chaque amorce, tiré des occurrences déjà
trouvées. Attention : la contrainte `-3` s'applique à l'extrémité 3' *de l'amorce*, quelle que soit sa
position sur la séquence ; étendre ou raccourcir du côté 3' déplace donc cette extrémité et modifie la
longueur d'amplicon.
(d) **Intégrée à la recherche, en phase de post-traitement** de la même exécution de obiprimers (pas d'outil
indépendant pour l'instant).
(e) Une paire dont la Tm n'est pas atteignable dans les bornes est **écartée**.

### 22.4 En attente

Réponses de l'utilisateur sur la voie B (questions de 21.4) ; choix de la statistique de Tm (22.3-a) ; nom et
sort de l'ancien Bs (22.1) ; en-têtes éventuels à prévoir pour les colonnes de population d'évaluation.

## 23. Décisions : un seul Bs, tout en mémoire (2026-11-03)

**Statut : spécification, aucun code.** Cette section **remplace** la proposition de 22.1 (colonne `bs_examples`,
passage d'évaluation en flux).

### 23.1 Décisions de l'utilisateur

1. **Un seul indice Bs, calculé d'une seule façon.** Pas de colonne `bs_examples`, pas de variante : « on décide
   un indice, il se calcule d'une certaine façon, point barre ». Le Bs d'ecoPrimers disparaît.
2. **Tout charger en mémoire.** Ne pas tout charger interdirait de lire la base de référence sur l'entrée
   standard, que les obitools permettent ; la base est lue par le mécanisme de lecture standard
   (comme `obiconvert`), sans format dédié. Le passage d'évaluation en flux de 22.1 est abandonné.
3. **Définition de Bs** : le nombre de taxons **bien identifiés parmi les cibles**, en considérant **tous les
   barcodes amplifiés, y compris ceux hors des cibles et ceux des contre-exemples**. Une espèce cible est
   mal identifiée si son barcode, unique parmi les cibles, est partagé avec une espèce non cible ou un
   contre-exemple.

### 23.2 Définition formelle retenue (traduction de 23.1-3)

Pour une paire, soit A l'ensemble des séquences qu'elle amplifie, **de n'importe quel rôle** (cibles,
contre-exemples, autres). Pour un amplicon a, T(a) est l'ensemble des taxons (au rang choisi) des séquences de A
qui le portent. Un taxon cible t est **bien identifié** s'il existe un amplicon a produit par t avec T(a) = {t}.

### 23.3 Conséquences

- **Fin de la notion de séquence « ignorée »** : les séquences qui ne sont ni cibles ni contre-exemples ne sont
  plus écartées à la lecture (comportement actuel, hérité d'ecoPrimers) ; elles sont **non cibles** : elles ne
  comptent ni dans les quorums de sensibilité et de faux positifs, mais leurs amplicons comptent dans Bs.
  Pour exclure des séquences de l'analyse, on filtre l'entrée en amont (`obigrep`), à la manière des obitools.
  Le rôle « contexte » de 21.2 devient donc simplement « non cible ».
- **Les valeurs changent, y compris pour le tutoriel** : une paire conservée chez tous les vertébrés amplifie
  aussi des non-Cyprinidae ; son Bs baissera par rapport aux tables d'ecoPrimers, même sans `-i`. Les tests de
  non-régression de la section 17 et de la section 18 ne peuvent plus comparer Bs/identifiés avec ecoPrimers sur
  la base entière. **Proposition, sans option ni colonne supplémentaire** : au niveau de la bibliothèque, un corpus
  construit avec les seules cibles et contre-exemples (c'est ce que fait `LoadCorpus` aujourd'hui) redonne
  exactement la définition d'ecoPrimers, car il n'y a alors aucune séquence hors sélection à amplifier ; les tests de
  non-régression continuent d'utiliser ce corpus, la CLI charge toujours tout.
- **Coût mémoire** : tout le fichier est conservé (131 Mb de bases pour 194 Mo de tas dans le tutoriel, contre
  5 Mb pour Cyprinidae aujourd'hui). Une représentation compacte (2 bits par base) reste possible plus tard ;
  non décidée.
- **Coût de calcul** : l'appariement approché des mots et l'énumération des amplicons passent de la sélection à
  tout le corpus (l'exécution `-r 7953 -i 1` du tutoriel, qui charge déjà tout, prend 1,9 s) ; les quorums ne
  comptent toujours que cibles et contre-exemples.

### 23.4 Points à confirmer (propositions, non validées)

1. **Dénominateur de Bs** : taxons cibles **amplifiés** (comme le code d'ecoPrimers : 2 650 identifiés sur 3 805
   amplifiés = 0,696) ou **tous** les taxons cibles ? Avec le premier, Bc (couverture) et Bs (discrimination)
   restent indépendants et `Bc x Bs` est la part des cibles résolues par la paire, ce qui est le gain mesuré par
   la sélection multi-marqueurs. Avec le second, Bs <= Bc et ce produit disparaît. **Je recommande le
   premier.** L'article (« nombre de taxons identifiés sur |E| ») et le code divergent déjà.
2. **Taxon à plusieurs amplicons** : bien identifié dès qu'**un** de ses amplicons est unique (règle d'ecoPrimers),
   ou seulement si **tous** le sont ? Je propose de garder la règle d'ecoPrimers.
3. **Séquence non cible sans taxon au rang choisi** : la traiter comme un taxon anonyme distinct (prudent :
   elle empêche l'identification d'une cible qui partage son amplicon) ou l'ignorer ? Je propose le premier.

## 24. Règle d'identification de Bs, et l'indice Bd (2026-11-03)

**Statut : spécification, aucun code.** Complète et corrige la section 23.

### 24.1 Décisions de l'utilisateur

1. **Dénominateur de Bs : les taxons cibles amplifiés** (comme ecoPrimers : 2 650 / 3 805 = 0,696).
   Bc (couverture) et Bs (discrimination) restent indépendants ; `Bc x Bs` est la part des cibles résolues.
2. **Règle d'identification : une espèce est bien identifiée si TOUS ses barcodes ne sont associés qu'à
   elle.** Dès qu'un barcode est partagé entre deux espèces, **les deux** sont mal identifiées. Cette règle
   remplace celle de 23.4-2 (que j'avais proposé de garder).
3. **Nouvel indice Bd (barcode discrimination)** proposé par l'utilisateur : la part des **barcodes** qui
   mènent sans ambiguïté à une espèce cible. Exemple de l'utilisateur : A porte les barcodes A1 et A2, B porte
   A2 et B1. Ni A ni B ne sont bien identifiées (elles partagent A2), mais **A1 et B1 sont discriminants**, A2
   ne l'est pas. Ici Bs = 0/2 et Bd = 2/3. (Avec la règle d'ecoPrimers on aurait Bs = 2/2.)

### 24.2 ecoPrimers applique bien l'autre règle (vérifié)

Lecture de `taxstats.c` : pour chaque amplicon distinct (parmi les exemples), on dresse l'ensemble de ses
taxons ; si cet ensemble n'a qu'**un** taxon (`taxoncount == 1`), ce taxon est marqué bien identifié. Un
taxon est donc compté **dès qu'un** de ses amplicons lui est propre. Confirmations : (i) mon implémentation de
cette règle reproduit toutes les colonnes `identifiés` et Bs de plus de vingt tables d'ecoPrimers (sections
12, 15, 17, 18) ; (ii) recalcul indépendant avec `obipcr` + Python sur la paire n° 1 de `Cyprinidae.ecoprimers`
(320 amplicons chez `obipcr`, 318 chez ecoPrimers qui impose `-3 2`) : règle « au moins un barcode unique »
239/312 = 0,766 contre 237/310 = 0,765 pour ecoPrimers ; règle « tous uniques » 238/312 = 0,763 ; **Bd =
240/265 = 0,906** (265 barcodes distincts). **Portée de (ii)** : sur cette paire les deux règles ne diffèrent
que d'**une** espèce sur 312, car les données du tutoriel ont un génome par espèce (4 espèces seulement
portent plusieurs barcodes) : ce recalcul ne tranche pas entre les règles à lui seul ; la preuve est le code
et les tables reproduites.

### 24.3 Conséquences

- **Les différences entre règles seront faibles sur le tutoriel** (un génome par espèce) et **plus fortes**
  sur des bases à plusieurs séquences par espèce (variation intra-spécifique), typiquement des bases de loci courts.
- **Non-régression** : la règle changeant, l'égalité des colonnes `identifiés` et Bs avec ecoPrimers n'est plus
  atteignable, même avec un corpus restreint aux cibles et contre-exemples (annule la proposition de 23.3 sur ce
  point). Toutes les autres colonnes (séquences, taxons, Bc, Yule, longueurs, GC, Tm, drapeaux good) restent
  comparables. Pour Bs, Bd et `identifiés` : tests sur cas synthétiques (dont l'exemple A1/A2/B1 ci-dessus) et
  **vérification indépendante** par `obipcr` + script, comme en 24.2, sur des paires choisies.
- **Seuil `-T` (spécificité minimale, 0,6 par défaut)** : calibré sur l'ancienne règle, plus indulgente. Le
  défaut sera peut-être à revoir ; à décider avec la spécification complète.
- **Implémentation (plus tard)** : la phase 2 de `FindPairs` doit enregistrer le hash d'amplicon de **toutes** les
  séquences amplifiées (cibles, contre-exemples, non-cibles) et non plus des seuls exemples ; un taxon cible est
  mal identifié si l'un de ses amplicons est porté par un autre taxon.

### 24.4 Questions restantes

1. **Bd** : l'inclure dans la sortie, et dans le classement ? Dénominateur proposé : les barcodes distincts
   portés par **au moins une cible** (un barcode propre à une non-cible ne « mène » pas à une espèce cible) ;
   comptage par barcode distinct, sans pondération par le nombre de séquences.
2. **Non-cible sans taxon au rang choisi** (23.4-3, proposé, non confirmé) : taxon anonyme distinct, donc
   barcode partagé = ambigu (prudent).
3. **`-T` par défaut** (voir ci-dessus).

## 25. Mise en œuvre des nouvelles définitions de Bc, Bs et Bd (2026-11-03)

Demande de l'utilisateur : mettre à jour le document et `obiprimers` pour que les trois indices soient calculés
selon leur nouvelle définition. Décisions prises dans la discussion (sections 23 et 24) : un seul Bs ; tout en
mémoire ; dénominateur de Bs = cibles amplifiées ; identification par **tous** les barcodes ; Bd ajouté et
présent dans toutes les sorties ; non-cible sans taxon = taxon anonyme (ambigu) ; seuil `-T` inchangé (0,6).

### 25.1 Changements de code

- **Rôles** (`roles.go`, `corpus.go`) : `Ignored` devient `Other` (non-cible) ; `LoadCorpus` conserve toutes les
  séquences ; `Counts.Others` ; `Corpus.Ignored` supprimé. Les quorums (sensibilité, faux positifs) ne
  comptent que cibles et contre-exemples (`SequencesMatched`, phase 1 de `countPairs` qui saute les
  non-cibles) ; `Taxa.CounterExamples` ne compte que les contre-exemples.
- **`-U`** (`DropMultiMatch(wm, corpus)`) : ne considère que les cibles et contre-exemples, comme ecoPrimers
  (la validation de la section 15 reste exacte) ; une non-cible qui contient deux sites ne fait pas écarter le mot.
- **`resolution.go`** (nouveau) : `resolve` calcule, à partir des amplicons d'une paire, les cibles
  amplifiées, les contre-exemples amplifiés, les taxons identifiés, les barcodes portés par les cibles et ceux
  qui sont discriminants. États de barcode : ambigu / anonyme / porté par une cible. `lenientIdentified` est
  la borne supérieure décrite en 25.2.
- **`amplicons.go`** (nouveau) : enregistrement d'un amplicon par séquence amplifiée, **tout rôle confondu**
  (hash 64 bits, paire, taxon, rôle ; 24 octets). Chaque worker trie ses enregistrements par paire ; ils ne
  sont ni copiés ni regroupés (l'ancienne version les dupliquait) ; les Tm minimales sont des minima par
  worker, fusionnés à la fin ; l'évaluation des paires est parallèle.
- **`pairs_eval.go`** : `PairStats.Bd` ; `evaluate` utilise `resolve`.
- **Sortie CSV** : colonne `bd` ajoutée juste après `bs`. L'aide de `-r` décrit les non-cibles ; le journal
  indique « other sequences » et le nombre de paires qui passent la borne de Bs.

### 25.2 Un problème de mémoire, et son remède exact

Première version : le pic de mémoire passait de **1,36 Go à 5,6 Go** (Cyprinidae, `-r 7953`), car la phase 2
gardait l'amplicon de toutes les séquences amplifiées pour les 25 699 paires qui passent les quorums, alors que
seules ~3 400 seront retenues.

Remède, **sans aucun changement de résultat** : un taxon identifié par la nouvelle règle (tous ses barcodes
propres parmi tout ce qui est amplifié) l'est aussi par la règle d'ecoPrimers (au moins un barcode propre parmi les
seules cibles), avec le même nombre de taxons cibles amplifiés. Donc `Bs_nouveau <= Bs_borne`. La phase 2 est
scindée : **2a**, amplicons des seules cibles (comme avant : Tm minimales, sites de la référence, borne de
Bs) ; **2b**, amplicons de toutes les séquences **pour les seules paires dont la borne atteint le seuil**.
Propriété vérifiée par `TestLenientIdentifiedIsAnUpperBound` (3 000 tirages aléatoires avec taxons anonymes,
plusieurs rôles, enregistrements en plusieurs morceaux).

Résultat (Cyprinidae `-r 7953 -e 3 --three-prime 2 -l 30 -L 150 -c`, M3 Max) :

| | pic mémoire | temps | paires |
|---|---|---|---|
| avant (ancienne définition) | 1 363 Mo | 6,5 s | 3 448 |
| nouvelle définition, sans élagage | 5 604 Mo | non mesuré | 3 247 |
| **nouvelle définition, avec élagage** | **1 217 à 1 372 Mo** | **4,7 à 5,3 s** | **3 247** |

42 506 paires candidates, 25 699 passent les quorums, **3 448** passent la borne (c'est exactement le nombre de
l'ancienne table), **3 247** passent le seuil avec la nouvelle règle. Mise en garde : avec `-T 0` aucune paire
n'est élaguée et la mémoire remonte (les tests à `-T 0` atteignent ~5,5 Go de tas, taxonomie NCBI comprise).
Piste non engagée si cela devient gênant : traiter les paires survivantes par lots bornés.

### 25.3 Valeurs : ancienne contre nouvelle règle (Cyprinidae, paire n° 1)

`CAAGCCAGCCACATAACC` / `TTCTATAACGGGTGATGC` : 310 taxons cibles amplifiés ; ecoPrimers : 237 identifiés,
Bs = 0,7645 ; obiprimers : **233** identifiés, **Bs = 0,7516**, **Bd = 0,8935**. L'écart est faible ici car la
base a un génome par espèce (peu de barcodes multiples) et que les non-cibles amplifiées par cette paire sont
peu nombreuses à partager un barcode ; il sera plus marqué sur des bases à plusieurs séquences par espèce.

### 25.4 Validation indépendante (obipcr + Python, base entière, `--three-prime 0`, `-T 0`)

Six paires (les deux premières de la sortie et quatre tirées au hasard), amplicons obtenus par `obipcr -e 3
-l 31 -L 149 -c` sur les **7 822 séquences**, Bs et Bd recalculés par un script qui n'utilise aucun code d'obiprimers.
Les paires amplifient de 3 202 à 7 796 séquences de la base entière (dont des milliers de non-cibles) :

| Paire | taxons cibles amplifiés | identifiés | Bs | Bd |
|---|---|---|---|---|
| 1 | 314 = 314 | 237 = 237 | 0,755 = 0,755 | 0,892 = 0,892 |
| 2 | 314 = 314 | 237 = 237 | 0,755 = 0,755 | 0,892 = 0,892 |
| 3 | 314 = 314 | 5 = 5 | 0,016 = 0,016 | 0,357 = 0,357 |
| 4 | 315 = 315 | 14 = 14 | 0,044 = 0,044 | 14/27 = 0,5185 = 0,5185 |
| 5 | 315 = 315 | 7 = 7 | 0,022 = 0,022 | 0,233 = 0,233 |
| 6 | 314 = 314 | 15 = 15 | 0,048 = 0,048 | 0,469 = 0,469 |

(format « obipcr+Python = obiprimers ».) Accord exact sur les 6 paires. Ces paires ont des Bs très bas : avec
tout ce qui est amplifié, un amplicon est vite partagé avec une non-cible. C'est l'effet recherché de la nouvelle
définition. Limite : la paire de référence 1 reste un cas où les deux règles diffèrent peu ; le cas qui les
distingue nettement (plusieurs barcodes par espèce) est couvert par les tests synthétiques (exemple A1/A2/B1 de
l'utilisateur, non-cible et contre-exemple partageant un barcode, taxon anonyme).

### 25.5 Tests

Nouveaux : `resolution_test.go` (exemple de l'utilisateur : Bs = 0/2, Bd = 2/3 ; tous uniques ; même espèce
plusieurs séquences ; non-cible ou contre-exemple partageant un barcode ; comptes de taxons ; taxon anonyme ;
ordre et découpage en morceaux ; borne supérieure) ; `TestNonTargetSharingABarcodeLowersBsAndBd` (de bout en
bout : une non-cible copie l'amplicon d'une cible, Bs = 7/8 et Bd = 7/8, quorums inchangés ; puis deux
contre-exemples amplifiés : 6 identifiés, Bd = 0,75, taxons contre-exemples = 2).

Adaptés : les tests opt-in contre ecoPrimers (`TestPairsAgainstEcoPrimers`, `TestNoMultiMatchAgainstEcoPrimers`,
`TestCounterExamplesAgainstEcoPrimers`) désactivent le seuil de spécificité, vérifient que **chaque paire
d'ecoPrimers est trouvée** avec les mêmes comptes de séquences et de taxons, Yule, Bc, longueurs, GC, Tm et
drapeaux good, et que les taxons identifiés ne dépassent jamais ceux de la règle indulgente d'ecoPrimers. Résultats :
Cyprinidae 3 448 paires, 0 absente, 0 champ différent ; Teleostei 407, 0 absente, 6 à une séquence de moins
(l'artefact du `N`, section 12) ; `-U` 1 366, 0 absente, 0 différente ; contre-exemples 173, 0 absente, 0
différente. obiprimers produit en plus 22 251 (Cyprinidae) et 2 389 (Teleostei) paires qu'ecoPrimers avait écartées avec son
propre Bs.

## 26. Optimisation de la mémoire (2026-11-03)

Mission : réduire la mémoire de `obiprimers` **sans changer un octet de la sortie**, en procédant par la mesure.
Contraintes respectées : tout reste en mémoire (l'entrée standard marche), un seul calcul par indice
(`resolution.go` garde ses définitions), aucun paquet partagé modifié (`obiformats`, `obiiter`, `obiseq`,
`obitax`, `obidefault` : `git diff` vide ; seuls `pkg/obiprimer`, `pkg/obikmer/seqquorum*.go`,
`pkg/obitools/obiprimers` et `cmd/obitools/obiprimers` changent). Pas de commit.

### 26.1 Méthode, et pièges

- **Garde-fou** : binaire de référence construit avant toute modification ; six CSV de référence (Cyprinidae
  `-r 7953`, Teleostei `-r 32443`, `-r 7953 -i 1`, `-U`, `-R NC_083213`, entrée standard) comparés par `cmp`
  après **chaque** étape : identiques à chaque fois. Idem pour `-T 0` (25 699 paires), `-t` (taxonomie NCBI) et le
  test d'échelle (Metazoa) comparés au binaire de référence.
- **Mesure** : exécutions entrelacées, médiane (5 à 9 exécutions ; le pic RSS d'une même commande varie de ±100 Mo
  avant, de ±10 Mo après). Le **temps CPU utilisateur est trompeur** : 40 % des échantillons d'un profil CPU sont du
  `runtime.usleep` des P inactifs. Pour juger le vrai coût d'un changement, comparer aussi avec `--max-cpu 2`.
- **Le RSS n'est pas le tas** (observation, pas hypothèse : mesures croisées `ps` + `runtime.MemStats` toutes les
  60 ms). Sur cette machine le RSS ne descend jamais pendant l'exécution et suit `Sys`, même quand le runtime
  annonce avoir *relâché* 258 Mo (`HeapReleased`). Une cause plausible, **non vérifiée** : macOS continue à compter
  les pages rendues par `madvise` tant que le système ne les reprend pas. Conséquence pratique, vérifiée par
  l'expérience du §26.3 (étapes 5 et 6) : ce qui fait monter le RSS, ce sont les très gros tableaux contigus
  alloués successivement, qui n'entrent pas dans les trous du tas et font mapper de la mémoire neuve.
- Les profils de tas par étape (tas après GC forcé, `inuse_space`) sont faits sur une copie du dépôt dans `/tmp`
  avec des marqueurs temporaires (`MemStage`) ; rien n'en reste dans le code.

### 26.2 Rapport de profil, avant toute modification (Cyprinidae, 7 822 séquences)

Tas vivant après GC, par étape, et ce qui le compose :

| Étape | vivant | composition |
|---|---|---|
| lecture terminée | 178 Mo | 165 Mo d'octets de séquence (131 Mo utiles : 26 % de marge laissée par les `append`) + attributs |
| appariement approché | 274 Mo | + 97 Mo de `[]WordMatches` (82 Mo utiles) |
| `selectPrimers` | 283 Mo | `perSeq` 109 Mo (82 utiles), construit par `append` (238 Mo alloués), tri stable en série : **0,74 s** |
| phase 2a | 511 Mo | + **214 Mo** d'`ampRecord` (8,1 millions × 24 o, avec la marge des `append`) |
| phase 2b | 517 Mo | + **220 Mo** (9,2 millions × 24 o) ; 2,17 Go alloués au total par ces `append` |
| fin | pic du tas 1 005 Mo | environ 2 × le vivant (`GOGC=100`) |

Les 1,1 Go « inexpliqués » étaient donc : (a) les enregistrements d'amplicons, 24 octets, marge des `append`,
copiés par pair puis triés ; (b) les occurrences stockées deux fois pendant `selectPrimers` (`WordMatches` puis
`perSeq`) avec des `append` en doublement ; (c) le retard du ramasse-miettes (le tas monte à ~2 × le vivant) ;
(d) la mémoire libérée par une étape qui n'est pas réutilisée par la suivante (§26.1). **Hypothèses réfutées** :
la lecture n'est pas le pic à cette taille (tas ≤ 330 Mo) ; `SeqKmerCounter` n'y est pour rien (il n'indexe que
les 325 exemples, 5 Mo) ; `GOGC=25` ne double pas le temps utilisateur ici (+7 % sur la référence, +5 % avec les
changements ci-dessous).

### 26.3 Optimisations, dans l'ordre où elles ont été faites

Médianes entrelacées (RSS en Mo, temps réel, temps utilisateur) ; « avant » est le binaire de l'étape précédente
mesuré **dans la même série** (les valeurs absolues changent d'une série à l'autre, d'où les deux colonnes).

| # | Changement | RSS avant → après | réel | utilisateur |
|---|---|---|---|---|
| 1 | **Amplicons** (`ampstore.go`, `amplicons.go`, `pairs_phase2.go`) : enregistrement de **12 octets** (deux tableaux parallèles `hash`/`taxon+rôle` au lieu de 24 octets), taille exacte connue avant l'écriture (passe de comptage par worker, puis écriture sans compteur partagé), ni copie ni tri ; **lots bornés** de paires (budget `DefaultAmpBudget` = 20 millions d'enregistrements, 240 Mo) ; maps de `resolve` réutilisées | 1 402 → 866 | 3,76 → 2,77 s | 18,9 → 13,5 s |
| 2 | **Occurrences** (`match.go`, `pairs.go`) : `hit` de 12 octets, regroupement par mot en tranches de taille exacte, `seqMatch` de **8 octets**, tris parallélisés | 949 → 704 | 2,80 → 2,04 s | 13,5 → 13,2 s |
| 3 | **Séquences à 2 bits par base** (`packed.go`) : plages d'exceptions pour tout octet ≠ a/c/g/t (restitution exacte), décodage dans un tampon par worker ; 178 → 42 Mo vivants | 705 → 628 | = | 13,4 → 13,8 s |
| 4 | **Compactage à la lecture** (`LoadPackedCorpus`, `corpus_load.go`) : chaque séquence est compactée et son taxon cherché par les goroutines de lecture, l'objet `BioSequence` n'est jamais gardé | petit jeu : sans effet mesurable ; **échelle : 1 999 → 1 390** | = | = |
| 5 | **Crochet `Release`** (`SearchParams.Release`, branché sur `debug.FreeOSMemory`) aux frontières d'étapes ; `debug.SetGCPercent` (étape 7) | 660 → 532 | = | = |
| 6 | **Tranches exactes par mot / par séquence / par paire** au lieu d'un seul tableau de 55 à 110 Mo | 546 → 365 | 2,12 → 2,19 s | 13,6 → 13,8 s |
| 7 | **Ramasse-miettes** : `SetGCPercent` dès le début de la commande (`memory.go`), sauf si `GOGC` est fixé par l'utilisateur | 370 → 305 avec 25 ; 292 avec 50 (autre série) | = | 13,9 → 14,4 s avec 25 |

**Étape 1, un faux départ instructif** : la première version (comptage exact, écriture dans un tableau partagé,
compteurs atomiques) faisait gagner 15 % de RSS mais coûtait **+18 % de CPU** (la passe de comptage énumère
deux fois les 7 822 séquences). Deux corrections : des curseurs privés par worker (le plan de comptage fixe
aussi quel worker lit quelle séquence, la passe d'écriture le rejoue), et surtout ne garder pour ces deux passes
que les occurrences **des amorces qui figurent dans les paires concernées** (613 amorces sur 1 200 pour les
survivants de la phase 2b) : l'énumération des paires d'occurrences est quadratique en densité. Résultat à
`--max-cpu 2` : réel 10,35 → 7,33 s et CPU 18,8 → 13,3 s **par rapport à la référence**. Même résultat par
construction : une paire dont une des deux amorces est absente de la liste ne peut pas être produite.

**Étape 5, le choix du pourcentage** : `GOGC=50` donne l'essentiel du gain pour presque rien de CPU ; `25` retire
encore 150 Mo à l'échelle pour +5 % de CPU. Retenu : 50 (`gcPercent` dans `memory.go`).

### 26.4 Résultat final (médianes entrelacées ; binaire de référence → binaire final)

| Cas | RSS | temps réel | temps utilisateur |
|---|---|---|---|
| Cyprinidae, fichier (9 exécutions) | **1 376 → 292 Mo** (−79 %) | 3,77 → 2,15 s | 19,2 → 14,2 s |
| Cyprinidae, entrée standard (7) | 1 365 → 369 Mo | 2,91 → 1,33 s | 17,2 → 12,7 s |
| Teleostei `-r 32443` (5) | 1 047 → 378 Mo | 3,41 → 2,58 s | 21,6 → 20,8 s |
| Cyprinidae `-T 0`, aucune paire élaguée (3) | **5 883 → 409 Mo** | 6,06 → 4,82 s | 46,3 → 46,1 s |
| avec `-t` (taxonomie NCBI complète) (3) | 2 664 → 1 273 Mo | 10,2 → 8,7 s | 32,7 → 29,5 s |
| échelle : 18 061 génomes, 593 Mb, `-r 33208` (3) | 3 189 → 1 370 Mo | 9,06 → 8,95 s | 64,5 → 68,2 s |

Pour comparaison, ecoPrimers (C) : 1,86 Go (Cyprinidae), 2,11 Go (Teleostei). Les tests : `go test ./pkg/obiprimer/
./pkg/obitools/obiprimers/ ./pkg/obithermo/ ./pkg/obikmer/ -race -count=1` verts ; tests opt-in sur les données
réelles (`OBIPRIMERS_TUTORIAL`) verts (65 s) : mêmes paires qu'ecoPrimers, 0 absente, 0 champ différent.

Tests ajoutés : `TestAmpBatches` et `TestAmpBudgetDoesNotChangeTheResult` (tout budget, jusqu'à un enregistrement,
donne les mêmes paires, indices et sites de référence) ; `TestPackedSeqRoundTrip` (toutes les longueurs modulo 4,
N, IUPAC, `u`, plages), `TestPackedCorpusGivesTheSameSearch` (recherche sur corpus compacté ou non, avec
ambiguïtés, référence, 3', circulaire) ; `TestLoadPackedCorpusEqualsLoadThenAssign` ; pour le compteur de k-mers :
budget sans effet sur les comptes, décompte des occurrences par partition, budget d'occurrences,
`DefaultPartitions`.

### 26.5 Le test d'échelle (jamais fait avant) et le compteur de k-mers

Sur 593 Mb d'un seul tenant (16 027 exemples), la mémoire n'est plus dominée par la phase 2 (aucune paire n'est
trouvée avec les Metazoa pour cibles : 18-mers conservés chez 70 % des cibles = 0) mais par deux autres postes,
que le petit jeu ne montrait pas :

1. **Le comptage des mots stricts** (`SeqKmerCounter`, 16 octets par occurrence de k-mer, 526 millions
   d'occurrences). Défauts trouvés : le tampon réutilisé par chaque worker gardait la taille de la plus grosse
   partition traitée (jusqu'à 106 Mo) ; la plus grosse partition fait 6 fois la médiane. Corrigé
   (`seqquorum*.go`) : les occurrences de chaque partition sont comptées à l'écriture et la partition est chargée
   dans un tableau de taille exacte ; un **budget d'occurrences** (`DefaultOccurrenceBudget` = 32 millions, 512 Mo)
   borne ce que les workers chargent ensemble, partitions les plus grosses d'abord (une partition plus grosse que le
   budget passe seule) ; **nombre de partitions automatique** (`DefaultPartitions` : 256, jusqu'à 1 024 quand la
   partition moyenne dépasserait ~1 million d'occurrences ; `--partitions 0`, nouveau défaut). La mémoire du comptage
   ne dépend plus de la répartition des k-mers. Le budget est un compromis temps/mémoire, **mesuré** (16 cœurs,
   RSS total / réel) : 8 M → 1 348 Mo / 14,4 s ; 16 M → 1 621 Mo / 10,2 s ; **32 M → 1 838 Mo / 8,75 s** ; sans
   limite → 2 157 Mo / 8,70 s. 32 M ne coûte rien en temps.
2. **La lecture** : les lots de 2 000 séquences (défaut commun des obitools) sont de ~33 Mo de bases chacun ;
   avec 16 workers et les lots qui attendent entre les étapes, 793 Mo restaient vivants pendant la lecture.
   `--batch-size-max` à 100 par défaut pour cette commande (`SetReadingDefaults`, appelée avant l'analyse des
   options, qui l'emportent) : à l'échelle 1 819 → 1 190 Mo (−35 %, budget de k-mers à 32 M), temps réel +2,5 %, CPU +7 % ;
   petit jeu : 306 → 274 Mo, +4 % de CPU. `--batch-mem` (existant)
   a aussi été évalué : 4 M → 2 312 → 1 582 Mo (+6 % CPU), mais le plafond en nombre de séquences est plus efficace.
   **À signaler** : le texte d'aide commun de `--batch-size-max` (paquet `obioptions`, partagé, non modifié) dit
   « default 2000 » alors que le défaut affiché par `getoptions` est maintenant 100 pour `obiprimers`.

### 26.6 Pistes écartées, et pourquoi

- **Clé `uint64` pour les tables de paires** (au lieu de la structure de 12 octets) : aucun gain mesurable
  (22,8 → 22,1 s de CPU à 2 cœurs, bruit). Annulé.
- **`GOGC=25` par défaut** : 5 % de CPU de plus que 50 pour 150 Mo de moins à l'échelle, rien sur le petit jeu.
  `GOGC` fixé par l'utilisateur est respecté, donc 25 reste à portée de main.
- **Budget d'amplicons plus petit** (`DefaultAmpBudget`) : non réglé finement. 20 millions d'enregistrements
  (240 Mo) tiennent en un lot pour les cas usuels ; `-T 0` passe en plusieurs lots sans perte de temps mesurable.
- **Relire le fichier** : exclu (entrée standard).

### 26.7 Ce qui reste, et décisions à prendre

- **Entrée standard** (369 Mo au lieu de 292) : les séquences sont gardées comme objets pendant l'extraction de la
  taxonomie embarquée (`embeddedTaxonomyInMemory`), avant d'être compactées. Les compacter dès la lecture
  demanderait de scinder la sélection (qui a besoin de la taxonomie) : non fait.
- **Taxonomie NCBI (`-t`)** : ~1,07 Go à elle seule (chargeur `obitax`, partagé, hors périmètre). La taxonomie
  embarquée (`obiannotate --taxonomic-path`) l'évite. À dire dans l'aide : non fait.
- **Budget du compteur de k-mers** (32 M occurrences) : décision de l'utilisateur si la mémoire prime sur le
  temps (16 M : −12 % de RSS à l'échelle pour +16 % de temps).
- `Corpus.Pack()` n'est plus appelé que par les tests (et utile à qui bâtit un corpus à la main) ; l'ancien chemin
  `LoadCorpus` + `AssignTaxa` reste pour les tests. À supprimer si l'on veut un seul chemin.
- Le crochet `Release` appelle une collecte complète à chaque frontière d'étape (4 appels dans le cas usuel : après
  la lecture, avant l'appariement, avant chaque lot de 2a et de 2b) ; sans effet mesurable sur le temps.

