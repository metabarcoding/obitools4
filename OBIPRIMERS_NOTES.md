# obiprimers — notes de travail

Ré-implémentation d'ecoPrimers dans obitools4. Ce fichier garde l'état des recherches,
des mesures et de l'avancement. Dernière mise à jour : 2026-11-03.

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
| `-t/--taxonomy` (obligatoire) | `-d` (base) | |
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
