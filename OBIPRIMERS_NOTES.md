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
suffirait. Non modifié (code partagé) : à décider.

Suite : étape 3 (recherche approchée des mots stricts, <= e substitutions, deux brins, dans exemples
et contre-exemples), puis paires et indices.
