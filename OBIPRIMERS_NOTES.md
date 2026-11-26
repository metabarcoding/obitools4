# obiprimers — état, blocages, essais

Ré-implémentation d'ecoPrimers dans obitools4 (commande `obiprimers`). Ce document décrit **l'état actuel**,
les **points de blocage**, les **essais réalisés et la raison de leur échec ou de leur succès**, et donne ce
qu'il faut pour reprendre le travail. L'historique chronologique a été retiré (il reste dans l'historique git de
ce fichier). Dernière revue complète du code : 2026-11-05.

Conventions : « **établi** » = lu dans le code ou mesuré, avec la source ; « **hypothèse** » = non vérifié.

---

## 0. En bref

- **Le mode A (conservé, ecoPrimers) est complet, validé et rapide** sur les jeux de poissons : il reproduit les
  tables d'ecoPrimers (sauf Bs, redéfini) en 1 à 3 s au lieu de 134 à 2 728 s, avec 0,3 à 1,4 Go.
- **Il ne passe pas à l'échelle des plastes de Spermatophyta** (11 506 exemples, fenêtre 10-300) : la phase 2
  (calcul exact de Bs) doit traiter 1,4 million de paires et 30,4 milliards d'amplicons en 1 522 lots, chaque lot
  repassant sur tout le corpus. Aucun run n'a abouti (§6).
- **Ce qui reste à faire en priorité** : réduire le coût de la phase 2 par des moyens **exacts** (§8.1), après
  avoir relancé le cas avec `-3 2` (le run bloquant utilisait `-3 0`, ce qui gonfle fortement le nombre de paires).
- **Spécifications en attente de décision** : voie B (amorces spécifiques), sélection multi-marqueurs,
  optimisation de Tm, amorces > 31 bases (§9). Aucun code n'existe pour elles.
- **Le dépôt contient des modifications non commitées** (§1.3) : pigeonnier de `countPairs`, correction de la clé de
  paire, hash des barcodes, barres de progression et logs d'étape. Tous les tests, y compris les tests opt-in
  contre ecoPrimers, passent avec elles (§4).

---

## 1. Reprendre le travail

### 1.1 Environnement

| | |
|---|---|
| Go | `/usr/local/go/bin/go` (pas dans le `PATH`) |
| Machine | MacBook M3 Max, 16 cœurs, 128 Go ; le *load average* y est souvent de 70 à 190 (autres processus) : **les temps mesurés sont bruités**, comparer des séries entrelacées et regarder les nombres de paires et la mémoire plutôt que les secondes |
| ecoPrimers (référence) | sources `~/Sync/travail/__MOI__/ecoprimers/src`, binaire `/usr/local/bin/ecoprimers` |
| Tutoriel + données poissons | `~/Documents/Enseignement/DNA metabarcoding spring school/ecoPrimers-Fish/` : `mito.vert.fasta.gz` (7 822 séquences), `ncbitaxo_20250211.tgz`, `Cyprinidae.ecoprimers`, `Teleostei.ecoprimers`, `Cyprinidae_14.fasta.gz`, `Teleostei_11.fasta.gz` |
| Plastes | `plastid/plastid_nov_2026_taxo.fasta.gz` dans le dépôt (619 Mo gz, en-têtes JSON avec `taxonomic_path`, 18 674 taxons embarqués) |
| Article | `~/Sync/travail/__MOI__/ecoprimers/Documents/Riaz et al. 2011 ... .pdf` (NAR 39:e145) |
| Binaires et sorties de référence des essais | `/tmp/obibin/` (**temporaire**, voir §1.2 pour les régénérer) |

Pièges d'outillage rencontrés : **zsh ne découpe pas une variable non guillemetée** (`$C` reste un seul
argument) : écrire les scripts de mesure avec `bash -c`. `pkill -f motif` tue aussi le shell dont la ligne de
commande contient le motif. `go test` imprime des avertissements C venant de `obiformats/fastseq_read.h` : sans
conséquence.

### 1.2 Commandes utiles

```
# construire
/usr/local/go/bin/go build -o /tmp/obibin/obiprimers ./cmd/obitools/obiprimers

# poissons, comme ecoPrimers (-e 3 -3 2 -l 30 -L 150 -c)
obiprimers -t ncbitaxo_20250211.tgz -c -l 30 -L 150 -3 2 -r 7953 mito.vert.fasta.gz            # Cyprinidae : 3 247 paires (T = 0,6)
obiprimers ... -r 7953 -i 1 ...                                                                # avec contre-exemples : 173 paires (200 sans seuil -T), celles d'ecoPrimers
# plastes, sans -t : taxonomie lue dans les séquences
obiprimers -c -l 10 -L 300 -r 58024 --no-progressbar -o sperma.csv plastid_nov_2026_taxo.fasta.gz   # cas bloquant (§6)

# tests : normaux, puis opt-in contre ecoPrimers (~64 s)
go test -count=1 ./pkg/obiprimer/ ./pkg/obitools/obiprimers/ ./pkg/obithermo/ ./pkg/obikmer/
OBIPRIMERS_TUTORIAL="$HOME/Documents/Enseignement/DNA metabarcoding spring school/ecoPrimers-Fish" \
  go test -count=1 -v -run 'AgainstEcoPrimers|SitesAreRealSites' ./pkg/obiprimer/
```

Sorties de référence à conserver pour toute modification qui ne doit pas changer les résultats (comparer ligne
à ligne **sans la colonne `index`**, qui est un rang) : Cyprinidae `-r 7953 -i 1` (173 paires d'ecoPrimers),
*Salmo* `-r 8028 -i 1` (147 625 paires), *Cyprinus* `-r 7961 -i 1 -x 0.01` (761), Cyprinidae `-r 7953`
(3 247). Les options communes sont `-c -l 30 -L 150 -3 2`.

### 1.3 État du dépôt

HEAD `468a49cf` (commit de l'utilisateur : fenêtre de longueur propre aux contre-exemples,
`--min-length-neg`/`--max-length-neg`, `PairParams.window(role)`). **Non commité** (travail de la session du
2026-11-04) : `amplicons.go`, `match.go`, `pairs.go`, `pairs_eval.go`, `pairs_phase2.go`, `search.go`, `words.go`,
`obitools/obiprimers/obiprimers.go` modifiés ; `pairkey_test.go` et `progress.go` nouveaux. Règle de travail :
**aucun commit ni commande VCS sans demande explicite**.

---

## 2. Ce que fait le programme

### 2.1 Rôles des séquences

| Rôle | Définition | Compte dans les quorums | Ses amplicons comptent dans Bs, Bd |
|---|---|---|---|
| **Exemple** (cible) | dans `-r` (toutes si absent) et hors `-E` | oui (sensibilité) | oui |
| **Contre-exemple** | dans `-i` ou `-E`, et pas exemple | oui (faux positifs) | oui |
| **Non-cible** (`Other`) | toutes les autres, **gardées en mémoire** | non | oui |

Pour exclure des séquences, filtrer l'entrée en amont (`obigrep`). Une séquence sans taxon au rang choisi est
un **taxon anonyme distinct** (un barcode qu'elle partage est ambigu). Filtrer par taxid, jamais par nom
(*Salmostoma bacalia* est un cyprinidé).

### 2.2 Pipeline (`Search`, `search.go`)

1. **Lecture** (`LoadPackedCorpus`) : chaque séquence est compactée à 2 bits/base et son taxon cherché par les
   goroutines de lecture ; l'objet séquence n'est jamais gardé. Taxonomie : `-t`, sinon extraite des
   `taxonomic_path` des séquences (deux passes pour un fichier, en mémoire pour l'entrée standard).
2. **Mots stricts** (`FindStrictWords`) : mots de `k = 18` (≤ 31), canoniques, présents **sans erreur** dans
   au moins `floor(|E| × q)` exemples (`q = 0,7`). Compteur disque par partitions (`obikmer.SeqKmerCounter`).
   `-U` : écarte d'abord les mots répétés dans un exemple (`Repeated`).
3. **Recherche approchée** (`MatchWords`) : chaque mot strict est cherché sur les deux brins avec ≤ `e`
   substitutions dans **toutes** les séquences. Index de graines en deux moitiés (principe des tiroirs).
   `-U` : `DropMultiMatch` écarte les mots à ≥ 2 occurrences dans une cible ou un contre-exemple.
4. **Amorces** (`selectPrimers`) : une amorce est **conservée** si elle touche ≥ `floor(|E| × s)` exemples
   (`s = 0,9`) ; **Good** si en plus elle touche ≤ `floor(|C| × x)` contre-exemples (`x = 0,1`).
5. **Paires, phase 1** (`countPairs`) : par séquence, tous les couples de sites (amont, aval) dont la longueur
   d'amplicon est dans la fenêtre, avec **au moins une amorce Good** (`Good(amont) || Good(aval)`), et les
   `-3` bases 3' de chaque amorce sans erreur. Les paires ne sont **inscrites** que depuis `M+1` exemples (pigeonnier,
   §7.1) ; toutes les autres séquences ne font que compter les paires déjà inscrites. Une paire est retenue si
   `q > s` et `qfp < x` (comparaisons en float32, comme le C).
6. **Phase 2a** (`boundPairs`) : amplicons des **cibles** seules → Tm minimales, sites sur la référence, et une
   **borne supérieure** de Bs (`lenientIdentified`). Les paires dont la borne est < `-T` sont rejetées.
7. **Phase 2b** (`evaluatePairs`) : pour les survivantes, amplicons de **toutes** les séquences amplifiées ;
   `resolve` calcule les taxons identifiés, Bs et Bd. Tri par `Yule × Bs` décroissant (puis `Primer1`, `Primer2`).

Les phases 2a et 2b procèdent **par lots de paires** (budget `DefaultAmpBudget` = 20 millions d'enregistrements de
12 octets, soit 240 Mo) : un **plan de comptage** fixe la taille exacte des tranches, puis chaque lot rescanne
le corpus pour écrire ses enregistrements.

### 2.3 Indices de qualité (définitions en vigueur ; elles remplacent celles d'ecoPrimers)

| Indice | Définition |
|---|---|
| **Yule** | `examples/|E| − counterexamples/|C|` |
| **Bc** | taxons cibles amplifiés / taxons cibles |
| **Bs** | taxons cibles **identifiés** / taxons cibles **amplifiés**. Un taxon est identifié si **tous** ses barcodes ne sont portés que par lui, parmi **tout** ce que la paire amplifie (cibles, non-cibles, contre-exemples). Un barcode partagé rend mal identifiés **les deux** taxons. |
| **Bd** | part des barcodes portés par au moins une cible qui mènent sans ambiguïté à **un seul** taxon |

Exemple de l'utilisateur : A porte A1 et A2, B porte A2 et B1 → Bs = 0/2, Bd = 2/3 (A1 et B1 discriminent, A2 non).
La règle d'ecoPrimers (`taxstats.c`, `taxoncount == 1`) compte un taxon **dès qu'un** de ses amplicons lui est
propre et ne regarde que les exemples : **Bs et `identified_taxa` ne sont donc plus comparables à ecoPrimers**
(toutes les autres colonnes le restent). Le seuil `-T = 0,6` vient de l'ancienne règle, plus indulgente ; son
défaut est peut-être à revoir. L'identification stricte implique l'indulgente, d'où la borne de la phase 2a :
`Bs_nouveau ≤ Bs_borne` (`TestLenientIdentifiedIsAnUpperBound`).

### 2.4 Options et sortie

Options (alias courts d'ecoPrimers) : `-t`, `-r`, `-i`, `-E`, `--rank` (species), `-c`, `--primer-length` (18),
`-e` (3), `-l` (0), `-L` (1000), `--min-length-neg` (0), `--max-length-neg` (0 = `-L`), `-3` (0), `-q` (0,7),
`-s` (0,9), `-x` (0,1), `-T` (0,6), `-U`, `-R`, `--salt-correction` (santalucia), `--salt` (0,05 M),
`--primer-concentration` (8e-7 M ; l'aide affiche 0.000001 par arrondi), `--partitions` (0 = automatique),
plus les options communes d'entrée/sortie. Pas de `-O` court (déjà pris dans obitools). **Non reprises** : `-p`, `-b`,
`-M`, `-k` (voir annexe A.3), `-f`, `-v`, `-A`.

Fenêtres de longueur : cibles et non-cibles `]-l, -L[` ; contre-exemples
`]--min-length-neg, max(--max-length-neg, -L)[` ; en phase 2 **tous** les rôles utilisent la fenêtre des cibles
(des barcodes de longueurs différentes ne peuvent pas être égaux).

CSV (une ligne par paire) : `index, primer1, primer2, tm1, tm1_min, tm2, tm2_min, gc1, gc2, good1, good2,
examples, counterexamples, yule, example_taxa, counterexample_taxa, bc, identified_taxa, bs, bd, min_length,
max_length, mean_length` ; avec `-R` : `reference_location`, `reference_amplicon`. Les résumés vont dans le
journal (stderr). Depuis la session du 2026-11-04, **le journal donne chaque étape** (nombres de mots, d'occurrences,
d'amorces, de paires candidates/retenues/survivantes, lots d'amplicons avec temps restant estimé) et des barres de
progression (`schollz/progressbar`) s'affichent si stderr est un terminal et sans `--no-progressbar` ; **les barres
n'ont pas été vues** (aucun terminal dans la session), seuls les logs sont vérifiés.

---

## 3. Carte du code

`pkg/obiprimer` (4 607 lignes hors tests, aucun fichier au-dessus de 440) :

| Fichier | Rôle |
|---|---|
| `roles.go` | `Role`, `Selection` (`-r/-i/-E`), `BelongingTo` |
| `corpus.go`, `corpus_load.go`, `packed.go`, `taxa.go` | corpus en mémoire, chargement parallèle ordonné, séquences à 2 bits (plages d'exceptions pour tout octet ≠ a/c/g/t), taxons denses (`Taxa.Of`) |
| `words.go` | `FindStrictWords`, `MinSequences` |
| `match.go`, `match_seed.go`, `match_naive.go`, `multimatch.go` | `MatchWords` (index de graines), oracle `matchNaive` (tests), `-U` |
| `pairs.go` | `PairParams`, `selectPrimers`, `forEachOccurrenceIn` (clé de paire, `Good\|\|Good`, contrainte 3'), `countPairs`/`scanPairs`, `sensitivityNeeded` |
| `pairs_eval.go`, `pairs_phase2.go` | `FindPairs`, `PairStats`, `boundPairs` (2a), `evaluatePairs` (2b) |
| `amplicons.go`, `ampstore.go` | énumération des amplicons par lots (`eachAmplicon`, `planAmplicons`, `collectAmplicons`), hash de barcode, stockage 12 octets |
| `resolution.go` | `resolve`, `lenientIdentified` (Bs, Bd) |
| `reference.go`, `search.go`, `progress.go` | emplacements `-R`, `Search`, barres et horloge de lots |

Ailleurs : `pkg/obitools/obiprimers/{options,obiprimers,output,memory}.go` (CLI, CSV, réglage du GC),
`cmd/obitools/obiprimers/main.go`, `pkg/obithermo` (Tm), `pkg/obikmer/seqquorum*.go` (compteur de mots stricts).

Représentations clés : une occurrence de site (`seqMatch`) tient en 8 octets (`position<<1 | brin inverse`) ;
une paire est `pairKey{p1, p2 int32, a1, a2 bool}` (identifiants d'amorces + drapeau « chaîne = mot canonique »).
La clé est **normalisée** : les deux extrémités sont ordonnées par `(mot, drapeau)`, `faux` avant `vrai`, de sorte
qu'une paire lue sur l'un ou l'autre brin donne la même clé (`TestPairKeyIsStrandInvariant`).

**Code mort à supprimer** : `primerSet.forEachOccurrence` (`pairs.go`), plus appelé nulle part depuis le passage à
`scanPairs`. `Corpus.Pack` et le couple `LoadCorpus` + `AssignTaxa` ne servent qu'aux tests.

---

## 4. Validation

### 4.1 Contre ecoPrimers (tests opt-in, 64 s, **tous verts avec le code non commité**)

| Test | Résultat |
|---|---|
| mots stricts (`TestStrictWordsAgainstEcoPrimers`) | Teleostei 255 = 255 ; Cyprinidae **1 207** contre 1 208 (voir §4.3) |
| recherche approchée (`TestMatchWordsAgainstEcoPrimers`) | pour chaque mot, même nombre de séquences que le journal d'ecoPrimers : 0 différence (1 207 et 255 mots) |
| paires (`TestPairsAgainstEcoPrimers`) | Cyprinidae : 25 699 paires (seuil de spécificité coupé), **0 absente, 0 champ différent** sur les 3 448 d'ecoPrimers ; Teleostei : 2 796 paires, 0 absente, **6 à une séquence de moins** (§4.3) |
| `-U` (`TestNoMultiMatchAgainstEcoPrimers`) | 1 366 paires d'ecoPrimers : 0 absente, 0 différente |
| contre-exemples (`TestCounterExamplesAgainstEcoPrimers`, `-r 7953 -i 1`) | 173 paires : 0 absente, 0 différente |
| `-R` (`TestReferenceSitesAreRealSites`) | 3 215/3 247 paires amplifient NC_022183 et 3 204/3 247 NC_083213, 0 emplacement faux |

Variantes d'options validées par comparaison ligne à ligne avec ecoPrimers sur Cyprinidae `-r 7953`
(identiques) : séquences linéaires, `-e 1`, `-e 5`, `-O 20`, `-O 25`, `-q 0.8 -s 0.95`, `-q 0.5 -s 0.8`, `-3 0`,
`-3 4`, `-l 10 -L 80`, `-E 7954`, `-E 7954 -x 0.05 -s 0.8`, `-E 84644 -x 0.95`.
Différences expliquées (mot strict en trop, artefact du `N`, Tm minimales) : `--rank genus`, `-T 0.3`.
**Jamais comparés** : `-e 0` (ecoPrimers n'avait pas terminé), les rangs autres que species et genus.

Bs et Bd ont été recalculés **indépendamment** (`obipcr -e 3 -l 31 -L 149 -c` + script Python, base entière,
`--three-prime 0`, `-T 0`) sur 6 paires : accord exact sur les taxons identifiés, Bs et Bd.

### 4.2 Tests unitaires

`resolution_test.go` (exemple A1/A2/B1, taxon anonyme, bornes), `search_test.go` (paire plantée, `-U`, `-R`,
non-cible partageant un barcode), `pairs_batches_test.go` (tout budget de lot donne le même résultat),
`packed_test.go`, `match_test.go` (oracle `matchNaive` sur 11 configurations), `window_test.go`,
`pairkey_test.go` (clé invariante au brin ; hash du barcode d'une paire `(P, rc(P))` identique sur les deux brins),
`obikmer/seqquorum*_test.go` (oracle force brute), `obithermo/model_test.go` (400 cas contre la libthermo C).

### 4.3 Écarts connus avec ecoPrimers

| Écart | Nature |
|---|---|
| Un mot strict en plus chez ecoPrimers (`TTTATGAAAACAAATAAA`, Cyprinidae) | trouvé dans 225 séquences sur 325, sous le quorum de 227 : ecoPrimers le garde à tort (mécanisme non établi). Il explique les 118/119 paires manquantes de `--rank genus` et `-T 0.3` |
| Teleostei : 6 paires à −1 séquence (NC_033548) | le registre à décalage d'`extractSite` d'ecoPrimers accepte un `N` en 3' par accident. **Établi par expérience** : l'émuler reproduit 100 % des deux tables. Non conservé (une base ambiguë n'est pas une correspondance 3' vérifiée) |
| Tm minimale | ecoPrimers décode les sites ambigus avec le même registre ; nous les ignorons |
| `-R` | celui d'ecoPrimers est faux (numéro de séquence utilisé comme indice) ; nous implémentons l'intention (emplacement GenBank, amplicon lu depuis `primer1`) |
| Tm | 273,15 et non 273,0 (validé) ; Owczarzy corrigé ; pas de gaps ni de bulges ; `NaN` quand aucun duplex stable |
| Bs et `identified_taxa` | redéfinis (§2.3) |
| Fenêtre des contre-exemples | options propres (`--min-length-neg`, `--max-length-neg`) |
| **Clé de paire** | corrigée le 2026-11-04 : pour deux sites en répétition directe (une amorce et son complément inverse), l'ancien code produisait deux clés selon l'orientation de la séquence. 4 paires de *Salmo* ont changé (par exemple 27 → 29 contre-exemples). On ne sait pas si ecoPrimers a le même défaut |
| Circulaire | une fenêtre à cheval sur la jonction est lue au-delà du tableau par le C (comportement indéfini) ; nous reprenons au début |
| `-e 0` | ecoPrimers n'applique pas le filtre de sensibilité aux amorces dans ce chemin ; nous si ; tables finales identiques en théorie, **non vérifié** |

---

## 5. Performances connues

### 5.1 Jeux de poissons (16 workers, M3 Max ; ecoPrimers sur 1 cœur)

| Jeu | obiprimers | ecoPrimers |
|---|---|---|
| Cyprinidae `-r 7953` | 2,2 s, 292 Mo | 134 s, 1,86 Go |
| Teleostei `-r 32443` | 2,6 s, 378 Mo | 350 s, 2,11 Go |
| Cyprinidae `-r 7953 -i 1` | ~1,9 s | 2 728 s, 756 Mo |
| Cyprinidae `-U` | — | 98,7 s, 1,18 Go |

Étape dominante après optimisation : la phase 2 (1 s sur ces jeux) ; le matching ne pèse plus que 1 %
(13 ms pour 1 207 mots × 325 séquences, 43 ms pour 255 mots × 3 909 séquences, contre 3,8 s et 9,4 s avec
`obiapat`).

Cas plus durs (même machine, `-e 3 -3 2 -l 30 -L 150 -c -r ... -i 1`) :

| Cas | Mots stricts | Paires candidates → retenues → finales | Temps |
|---|---|---|---|
| *Salmo salar* (1 séq.) `-x 0,1` | 16 665 | 9,9 M → — → 850 357 | 19,2 s (pairs ~17,6 s) |
| *Salmo salar* `-x 0,01` | 16 665 | 173 180 → — → 42 142 | 2,5 s |
| *Salmo* (7 séq.) `-x 0,1` | 15 534 | 8,3 M → 527 945 → 147 625 | 12,2 s |
| *Cyprinus* (10 séq.) `-x 0,1` | 15 412 | 5,8 M → — → 110 886 | 10,4 s |

(Colonne « candidates » avant le pigeonnier de §7.2 ; avec lui, *Salmo* passe à 597 k candidates.) À `-x 0,1` le
coût est dans les **paires**, pas dans le matching ; la même cible avec `-x 0,01` coûte 8 fois moins.

### 5.2 Mémoire (optimisation terminée, sortie inchangée octet pour octet)

Cyprinidae : 1 376 → 292 Mo ; entrée standard 1 365 → 369 Mo ; Teleostei 1 047 → 378 Mo ; `-T 0` 5 883 → 409 Mo ;
avec `-t` NCBI 2 664 → 1 273 Mo (le chargeur de taxonomie NCBI, partagé, coûte 1,07 Go à lui seul : la
taxonomie embarquée l'évite ; l'aide ne le dit pas) ; échelle 18 061 génomes / 593 Mb : 3 189 → 1 370 Mo.
Moyens : enregistrement d'amplicon de 12 octets en tranches de taille exacte par paire, lots bornés, occurrences
de 8 octets en tranches exactes, séquences à 2 bits, compactage à la lecture, `GOGC = 50`
(`memory.go`), `debug.FreeOSMemory` aux frontières d'étapes (`Release`), `--batch-size-max` à 100 par défaut pour
cette commande, budget d'occurrences du compteur de k-mers (32 M, 512 Mo) et nombre de partitions automatique
(256 à 1 024). Pistes écartées : clés `uint64` pour les tables de paires (aucun gain mesuré), `GOGC = 25` par
défaut (+5 % de CPU pour −150 Mo à l'échelle), relire le fichier (exclu : entrée standard).

Pièges de mesure : le RSS **n'est pas** le tas sur macOS (il suit `Sys` et ne redescend pas ; les très gros
tableaux contigus font mapper de la mémoire neuve) ; le temps CPU utilisateur est trompeur (40 % des
échantillons sont du `usleep` de P inactifs : comparer aussi avec `--max-cpu 2`) ; mesurer en séries entrelacées et
prendre la médiane. À l'échelle des plastes les pics observés sont 11 à 16 Go (§6).

---

## 6. Le blocage : l'échelle des plastes de Spermatophyta

**Cas** : `-c -l 10 -L 300 -r 58024` sur `plastid_nov_2026_taxo.fasta.gz` ; e = 3, `-3 0`, q = 0,7, s = 0,9,
x = 0,1, T = 0,6. 11 506 exemples (11 316 espèces), 1 012 non-cibles, **0 contre-exemple**.

| Étape | Résultat | Durée |
|---|---|---|
| taxonomie embarquée (1re passe) | 18 674 taxons | ~10 s |
| mots stricts | 10 989 (quorum 8 054) | 77 s |
| `MatchWords` | **266 882 613** occurrences (~23 000 sites par séquence) | 18 s |
| `selectPrimers` | 10 840 amorces conservées, **toutes Good** (pas de contre-exemple) | 9 s |
| `countPairs` | **12 947 420** paires candidates, **1 407 309** passent la sensibilité | **19 min 40 s** |
| comptage des amplicons (`planAmplicons`) | **30 410 837 549** amplicons cibles, **1 522 lots** | ~10 min |
| phase 2a | non terminée après 37 min (RSS 11,4 Go ; 16 Go dans un run précédent) | — |

Chronologie établie à partir des horodatages du journal (run du 2026-11-04 à 14:53). Un run précédent est mort sans
message (cause inconnue, aucune trace relevée).

**Profil de la phase 2a** (`sample` macOS, 8 s) : `forEachOccurrenceIn` ~29 500, threads inactifs ~24 000,
`runtime.mapaccess2` ~17 700, `collectAmplicons` ~8 800, `eachAmplicon` ~4 400, hachage ~7 000, Tm (`duplex`) ~2 500.
Le matching n'apparaît pas.

**Ce qui est établi** : (1) la sélection des paires n'est pas le problème ; (2) 1,4 M de paires × > 90 % de
11 506 séquences donne au moins 1,5·10¹⁰ enregistrements ; les 3,0·10¹⁰ observés correspondent à ~2 amplicons par
couple (paire, séquence) (cause non cherchée : sites approchés voisins) ; (3) `-3` n'a pas été donné.
**Établi par lecture du code** : chaque lot (`collectAmplicons` → `eachAmplicon`) repasse sur toutes les séquences,
filtre les sites des amorces du lot, énumère **tous** les couples de sites restants dans la fenêtre — y compris
ceux qui ne forment aucune paire du lot — puis construit, normalise et cherche une clé dans une table de 1,4 M
d'entrées. **Hypothèse, non mesurée** : c'est là que passe le temps (le profil va dans ce sens, mais le temps par
lot n'a jamais été observé : l'horloge de lots a été ajoutée après coup).

**Un sous-échantillon ne donne pas de borne utile** : un taxon non identifié dans un sous-ensemble l'est aussi
dans l'ensemble, donc `Bs ≤ 1 − NI_S/A` ; rejeter sous T = 0,6 exige `NI_S > 0,4·A`, donc plus de 40 % des
taxons dans l'échantillon. Aucune borne bon marché n'est connue.

**Structure des mots stricts** (maximal chains de mots décalés d'une base, sur les deux brins, sans minimiseur) :

| Jeu | Mots | Chaînes | Médiane (mots) | Moyenne | p90 | Max | Mots dans des chaînes ≥ 83 mots | Chaînes si coupe à 100 pb |
|---|---|---|---|---|---|---|---|---|
| **Spermatophyta** | 10 989 | **634** (79 isolées) | **6** | 17,3 | 41 | 334 | 37,6 % | 672 |
| Cyprinidae | 1 207 | 99 | 8 | 12,2 | 24 | 120 | 9,9 % | 100 |
| *Cyprinus* | 15 412 | 80 | 75 | 192,7 | 652 | 1 914 | 93,0 % | 233 |
| *Salmo* | 15 534 | 99 | 87 | 156,9 | 366 | 2 302 | 91,6 % | 249 |
| *Salmo salar* | 16 665 | **1** (anneau) | — | — | — | 16 665 | 100 % | 201 |

Les zones conservées des plastes sont courtes (62 % des mots dans des chaînes de moins de 100 pb) ; presque
aucun ramifiement (6 mots sur 10 989 ont plus de 2 voisins). Pour un clade étroit, la chaîne est le génome entier.

---

## 7. Essais réalisés et verdicts

### 7.1 Réussis (en place)

| Essai | Résultat |
|---|---|
| Index de graines en deux moitiés (`MatchWords`) | ×220 à ×270 sur `obiapat` ; égal à l'oracle et à ecoPrimers. Mon premier calcul l'avait écarté à tort (j'avais considéré des graines de e+1 morceaux) |
| Évaluation en deux phases avec borne de Bs (§2.2) | Cyprinidae : mémoire 5,6 → 1,2 Go sans changer un résultat |
| Optimisation mémoire (§5.2) | −79 % de RSS, sortie inchangée |
| **Pigeonnier dans `countPairs`** | une paire retenue manque au plus `M = E − n_min` exemples (`n_min` : plus petit nombre vérifiant `float32(n)/float32(E) > s`) ; elle apparaît donc dans au moins un de n'importe quels `M+1` exemples. On **enregistre** les paires depuis `M+1` exemples seulement (ceux qui ont le moins de sites), les autres exemples et tous les contre-exemples ne font que **consulter** ; seules les amorces des paires déjà candidates sont examinées. Exact : 4 sorties de référence identiques octet pour octet. Paires candidates : *Salmo* 8,3 M → 597 k, *Cyprinus* `-x 0,01` 30 606 → 5 073, Cyprinidae `-i 1` 2 304 → 516 ; *Salmo* 23 → 16 s, 5,2 → 2,6 Go (temps bruités). `info.CandidatePairs` change donc de sens |
| Correction de la clé de paire | voir §4.3 ; test qui échouait avant |
| Hash de barcode symétrique pour `(P, rc(P))` | avant, un même barcode donnait deux hash selon l'orientation de la séquence : un barcode partagé par deux taxons restait invisible et **Bs était surestimé**. Aucune paire de ce type dans les 4 sorties de référence (0 ligne `primer1 == primer2`) |
| Logs d'étape, barres, horloge de lots | en place ; logs vérifiés, barres non vues |

### 7.2 Rejetés ou non retenus

| Essai | Résultat mesuré | Raison de l'abandon |
|---|---|---|
| **Élagage des graines par comptage exact dans les contre-exemples** | part de graines à compte exact > `outQ` : à `-x 0,1` 5,4 à 6,5 % ; `-x 0,05` 8 à 14 % ; `-x 0,01` 23 à 42 % ; `-x 0,001` 62 à 91 % (*Salmo salar*, *Salmo*, *Cyprinus*) | un mot partagé avec C n'est jamais Good mais reste un partenaire légitime (`Good\|\|Good` : une amorce spécifique + une amorce universelle). On ne peut donc l'écarter sans perte. Gain maximal ~10 % à `-x 0,01`. Un scan par table de hachage des graines dans C coûte 0,17 s (le compteur disque : 4,3 s). **Non implémenté** |
| Voie B « une amorce spécifique + un partenaire » | — | **déjà présente dans le mode A** : `Good` signifie « ≤ `outQ` contre-exemples » et une paire exige `Good\|\|Good`. La voie B n'est donc pas un second algorithme mais le réglage `-x` bas |
| **Grossier-fin par superprimers** (minimiseurs `m = 6, 8, 10` ; un représentant Good/le plus couvert par superprimer ; `FindPairs` sur les représentants ; raffinement par produit cartésien des membres des paires retenues) | rappel contre l'exhaustif : Cyprinidae 67 à 93 % ; *Cyprinus* 82 à 83 % mais **top 100 : 32 à 44 %** ; *Salmo* 68 à 76 % (top 100 : 56 à 90 %). Temps ×2 seulement (le raffinement développe 3,6 à 7,8 k paires grossières en 90 à 112 k). Tailles moyennes de superprimer : 3,6 à 6,8 mots (max 9 à 25) | rappel insuffisant sur les **meilleures** paires, gain faible. La métrique compte une paire exacte (une voisine décalée d'une base retrouvée compte comme perte) : la perte réelle de qualité est peut-être moindre, **non mesurée**. La mécanique de raffinement est juste (0 valeur différente sur les paires retrouvées) |
| Composantes connexes de mots (regroupement par chevauchement) | *Salmo salar* : 1 chaîne de 16 665 mots | dégénère pour les clades étroits |
| Recherche approchée SIMD, arrêt anticipé sous le quorum de 90 % | — | inutile : le matching ne pèse plus que 1 % ; arrêt anticipé mesuré sans effet (Teleostei : 255/255 mots conservés) |
| Sous-échantillonnage comme borne de Bs | — | impossible, voir §6 |
| Clés `uint64`, `GOGC = 25` par défaut | voir §5.2 | pas de gain / coût CPU |

---

## 8. Pistes

### 8.1 Exactes (ne changent pas la sortie), à tenter d'abord

1. **Relancer Spermatophyta avec `-3 2`** (paramètre, pas élagage) pour mesurer le vrai ordre de grandeur et lire le
   temps par lot dans le nouveau journal.
2. **Élagage des sites par la contrainte 3' quand `-3 > 0`** : un site dont les bases 3' sont en erreur ne forme
   jamais d'amplicon. Après avoir calculé `ex`/`co` (qui définissent amorces conservées et Good, et ne doivent pas
   changer), le retirer de `perSeq`. Aujourd'hui le test 3' est fait par **couple** en relisant la séquence ; par site
   il coûterait O(`-3`). Les jeux de poissons utilisent déjà `-3 2` : la sortie doit rester identique.
3. **Énumération à partir d'ancres dans la phase 2** : les paires sont déjà triées par première amorce ; si un lot
   regroupe toutes les paires de quelques amorces, il suffit de regarder, dans chaque séquence, les voisins (dans la
   fenêtre) des sites de ces ancres, au lieu de tous les couples du filtre.
4. **Matrice de bits** amorce × amorce × orientation (~59 Mo pour 10 840 amorces) testée avant de construire la clé
   et de chercher dans la table de 1,4 M d'entrées.
5. **Hash de barcode par préfixes** en O(1) (`ampHash` relit tout l'amplicon ; ~10 % du profil).
6. **Avertissement + estimation** quand le total d'amplicons dépasse un seuil (le chiffre est connu après
   `planAmplicons`), avec conseil de resserrer `-s`, `-q`, `-e`, `-3` ou `-r`.
7. Supprimer `forEachOccurrence` (mort).

### 8.2 Idées non validées

- **Famille de paires de chaînes** : à l'intérieur d'une paire de chaînes conservées, les barcodes de tous les
  membres sont des sous-intervalles emboîtés. Calculer l'identification de tous les membres ensemble éviterait de
  les énumérer un par un. Il faut mesurer combien de paires de chaînes coexistent à moins de 300 pb (positions
  des sites nécessaires : matching complet, 18 s sur Spermatophyta).
- **Dominance par inclusion** : si le barcode de A est inclus dans celui de B dans chaque séquence et que A et B
  amplifient **exactement les mêmes séquences**, alors `Bs(B) ≥ Bs(A)` ; si B échoue sous `-T`, A aussi. Condition
  nécessaire non garantie (deux amorces décalées d'une base n'amplifient pas forcément les mêmes séquences) : il
  faudrait un masque de bits de séquences par paire (12 500 bits × 1,4 M ≈ 2,2 Go) et les positions des sites.
  Gain inconnu.

### 8.3 Avec perte (exigent une décision explicite de l'utilisateur)

- **Pré-classement approximatif** de Bs sur un échantillon de séquences pour trier les paires, puis calcul exact sur
  les K premières.
- **Grossier-fin amélioré** : plusieurs représentants par superprimer et seuil grossier abaissé ; le gain de temps
  promet de rester modeste (§7.2). Si on le reprend, mesurer le rappel par **voisinage** (le meilleur score trouvé
  parmi les paires décalées de quelques bases) et non par identité de paire.

---

## 9. Spécifications en attente (aucun code)

### 9.1 Sélection multi-marqueurs (multiplex)

Les options `-b/-M/-k/-p` d'ecoPrimers sont **abandonnées** (annexe A.3). Formulation retenue (décisions de
l'utilisateur) : un ensemble de paires qui, ensemble, discriminent un maximum d'espèces ; **une paire de base**
couvrant tout le groupe visé (Bc proche de 1), choisie d'abord, puis des **paires spécialisées** par clade mal
résolu ; la granularité dépend du budget (peu de paires → clades hauts, beaucoup → genres) ; une espèce est résolue
si **un** marqueur la résout ; l'unicité se juge sur **tout ce qui est amplifié**. L'option « éviter d'amplifier les
autres clades » n'est pas la valeur par défaut. Budget : max d'espèces résolues avec ≤ K paires, ou min de paires
pour un taux donné — même glouton avec deux conditions d'arrêt. **Manquent** : l'ensemble des taxons identifiés par
chaque paire (`PairStats` ne garde que des nombres), un corpus chargé une fois et une sélection refaite en mémoire par
clade, les contraintes de multiplex (Tm voisines, dimères).

### 9.2 Optimisation de Tm

Idée de l'utilisateur : fixer la Tm et ajuster la longueur des amorces à la fin, **pour n'importe quelle paire**, avec
deux paramètres : écart maximal de Tm entre les amorces et Tm cible. Décisions : cible sur l'**apparié imparfait**
(Tm sur les sites réels) ; allongement/raccourcissement du côté le mieux conservé (profil de conservation par
position) ; phase intégrée en post-traitement ; une paire non satisfaisable est écartée. **Points ouverts** : la
statistique des Tm de sites (le minimum est dominé par des aberrants : 12 à 45 °C pour des Tm parfaites de 52 à
62 °C ; moyenne, médiane ou quantile bas ?), les bornes de longueur, et le fait que la contrainte `-3` suit
l'extrémité 3' de l'amorce (étendre de ce côté déplace l'amplicon). Raccourcir par le 5' conserve la sensibilité
mais perd de la spécificité : réévaluer la paire.

### 9.3 Amorces plus longues que 31 bases

L'utilisateur souhaite jusqu'à 61 bases (comme son ancien agrep 64 bits). Limite actuelle : 31, inscrite dans
`match.go`, `words.go`, `pairs.go`, `obikmer` (k-mers sur 64 bits) et des tampons fixes. L'index de graines se
généralise à 128 bits ; le compteur de mots stricts resterait limité à 31 : chercher un cœur conservé de ≤ 31 bases,
puis étendre pendant la phase de Tm.

### 9.4 Voie B (amorces spécifiques d'une espèce ou de quelques espèces)

Constat nouveau (§7.2) : le mode A en fait déjà une version naïve. Questions ouvertes de la spécification : exclusion
stricte ou marge (une tolérance `e_exclusion` sur C distincte de `-e`, score de marge pondéré en 3') ; tailles
typiques de T et de C ; type de séquences ; plusieurs cibles à résoudre entre elles ou seulement à détecter ;
« une amorce spécifique + une générale » acceptable ; sensibilité sur T (0,9 ou toutes). Défaut `-x 0,1` trop
permissif pour une cible étroite. **Échelle** : pour T = un génome, le nombre de mots candidats est celui des
positions (16 000 pour une mitochondrie).

### 9.5 Autres décisions à prendre

`-T` par défaut ; statistique de Tm (9.2) ; Bd dans le classement (aujourd'hui seulement dans la sortie : le
classement est `Yule × Bs`).

---

## 10. Dette technique et points d'attention

- **Code mort** : `forEachOccurrence`. Chemins réservés aux tests : `Corpus.Pack`, `LoadCorpus` + `AssignTaxa`.
- **Palindromes** : `MatchWords` émet un mot égal à son complément inverse **deux fois à la même position** (un
  brin chacun). Les clés de paires correspondantes pourraient se dédoubler. Jamais observé (0 amorce palindromique
  dans les 4 sorties de référence), ni testé, ni gardé.
- **`CandidatePairs`** (`PairInfo`, journal) : nouveau sens (paires présentes dans les `M+1` premiers exemples) ;
  ne pas le comparer aux anciennes valeurs.
- **Aide** : `--batch-size-max` affiche « default 2000 » dans le texte partagé (`obioptions`) alors que le défaut
  de la commande est 100 ; l'aide ne dit pas que `-t` NCBI coûte 1,07 Go.
- **Non faits** : tests shell `obitests/obitools/obiprimers`, documentation utilisateur, comparaison de `-e 0`,
  comparaison des rangs autres que species/genus, vérification visuelle des barres de progression.
- **Entrée standard** : 369 Mo au lieu de 292 (les objets séquences sont gardés pendant l'extraction de taxonomie).
- **`Release`** appelle une collecte complète à chaque frontière d'étape (sans effet mesurable sur le temps).
- **Course de données corrigées hors périmètre** (déjà dans le dépôt) : `obiutils/pipe.go`
  (`globalLockerCounter` en `atomic.Int64`), `obiiter.RebatchBySize`, `obiutils.Set.String()` (ordre stable).
- Un oracle indépendant existe pour presque tout (force brute `matchNaive`, `obipcr` + Python pour Bs/Bd, ecoPrimers
  lui-même) : s'en servir avant de croire un gain.

---

## Annexe A. ecoPrimers (référence)

### A.1 Algorithme (lu dans le code C)

1. **Mots stricts** (`strictprimers.c`, `filtering.c`) : mots de `lp` bases (18, ≤ 31), 2 bits/base, forme canonique
   `min(D,R)`, présents sans erreur dans ≥ 70 % des exemples (`-q`) ; pré-filtre par table de comptage des 13-mers.
2. **Recherche approchée** (`aproxpattern.c`, `apat_search.c`) : chaque mot, deux brins, ≤ `e` substitutions
   (agrep bit-parallèle), dans exemples et contre-exemples. « good » : ≥ 90 % des exemples (`-s`) et ≤ 10 % des
   contre-exemples (`-x`). Seules les amorces « conservées » suffisent pour une des deux d'une paire. `-3` est
   vérifié plus tard, au niveau de la **paire**.
3. **Appariement** (`pairs.c`) : par séquence, occurrences triées ; distance dans `]lmin, lmax[`, ≥ 1 amorce good ;
   clé canonique ; arbre `tsearch` avec **un amplicon par occurrence** (goulot mémoire).
4. **Indices** (`taxstats.c`) : Bc = taxons exemples amplifiés / taxons exemples ; Bs = taxons dont l'amplicon n'est
   partagé avec aucun autre taxon **exemple** / taxons exemples ; `yule = in/|E| − out/|C|` ; tri `yule × Bs` ;
   seuil `-T` sur Bs.
5. **Tm** (`libthermo`) : plus proches voisins, SantaLucia ou Owczarzy ; Tm exacte et pire cas.

### A.2 Mesures d'ecoPrimers (`-e 3 -3 2 -l 30 -L 150 -c`)

| Jeu | Séquences | Paires totales | Retenues | Temps | Mémoire |
|---|---|---|---|---|---|
| Cyprinidae `-r 7953` | 325 | 42 682 | 3 448 | 134 s | 1,86 Go |
| Teleostei `-r 32443` | 3 909 | 9 407 | 407 | 350 s | 2,11 Go |
| Cyprinidae `-r 7953 -i 1` | 325 + 7 497 CE | 2 050 | 173 | 2 728 s | 756 Mo |
| Cyprinidae `-U` | 325 | — | 1 366 | 98,7 s | 1,18 Go |

Sémantique `-r/-i/-E` (testée sur *Salmo*, taxid 8028) : `-r 8028` → 7 exemples (6 espèces), 0 contre-exemple (le
reste est **supprimé** par ecoPrimers) ; `-r 8028 -i 8015` → 76 contre-exemples (les autres Salmonidae) ; `-r 8028
-i 7742` → 7 815 ; `-r 8015 -E 8028` → 76 exemples, 7 contre-exemples. `-U` : un mot est écarté s'il a ≥ 2
occurrences approchées, deux brins confondus, dans **une** séquence quelconque (exemple ou contre-exemple).

### A.3 Options abandonnées : `-b`, `-M`, `-k`, `-p`

Pistes explorées pendant la thèse de T. Riaz, jamais abouties (problème mal posé) : `-p` est commenté dans
`ecoprimer.c`, `-k` n'est lu que par du code inatteignable, `-M` contient un `break` inconditionnel qui compare
chaque paire à la seule première autre paire, `-b` est une heuristique de « liens » sans justification. Décision de
l'utilisateur : **non portées**. Le modèle de la thèse (recouvrement partiel d'espèces entre paires) est quasi
inapplicable : deux paires visant des clades différents partagent presque toujours leurs amorces, donc le
recouvrement est complet ou nul.

## Annexe B. `pkg/obithermo`

Plus proches voisins, corrections de sel, Tm exacte et pire cas. API : `NewModel(primerConc, targetConc, salt,
method)`, `DefaultModel()`, `(*Model).SelfTm`, `(*Model).Tm(primer, site)`, `GCCount`. Conditions par défaut
(celles d'ecoPrimers) : amorce 8e-7 M, sel 0,05 M, SantaLucia. 17,7 ns par appel `Tm`, sans allocation, sûr en
concurrence. `nn_table.go` (112 paramètres) est **généré** depuis `nnparams.c` par un script jetable perdu.
Validation : 400 cas contre la libthermo C (écart ≤ 5·10⁻⁶ K, dû aux constantes `float` du C ; 7·10⁻¹³ K en les
émulant) et 13 Tm d'ecoPrimers à ±0,05 °C. Pas de limite de longueur d'amorce.

## Annexe C. `obikmer.SeqKmerCounter`

`KmerSetGroupBuilder` ouvre n × P fichiers (inutilisable pour des milliers de génomes). Le compteur écrit les
super-kmers de **toutes** les séquences dans un seul jeu de P fichiers (`[uvarint id][uvarint len][2 bits/base]`),
groupés par minimiseur ; `CountAtLeast(q)` traite les partitions en parallèle sous un **budget d'occurrences**
(`DefaultOccurrenceBudget` = 32 millions, 16 octets par occurrence) et renvoie `SeqKmerCount{Kmer, Sequences,
Repeated}`. Les bases non ACGTU coupent la séquence ; une séquence circulaire reçoit ses `k−1` premières bases.
`DefaultPartitions` : 256, jusqu'à 1 024 quand la partition moyenne dépasse ~1 M d'occurrences. À 593 Mb et
526 M d'occurrences : 1 838 Mo, 8,75 s avec le budget de 32 M (16 M : −12 % de RSS pour +16 % de temps).

## Annexe D. `pkg/obiapat` (apat, agrep en C) : sémantique vérifiée et coût

Source : `apat.h`, `apat_parse.c`, `apat_search.c`, `obiapat.c`, et un test jetable (2026-11-05, supprimé).

**Sémantique (vérifiée par exécution)** :
- Le motif accepte les codes IUPAC, des classes `[AT]`, la négation `!` et `#` (position où une erreur est interdite), jusqu'à 64 symboles et 64 erreurs ; `obipcr` l'appelle **sans indel** (`allowsIndel = false`).
- **Dans la séquence, toute lettre autre que a, c, g, t compte comme une erreur** : motif `ACGTAC` contre `ttacntactt` ou `ttacrtactt` → aucune occurrence à `e = 0`, une occurrence à `e = 1`. Un `u` de la séquence n'est pas lu comme `t` (motif `ACGTAC` contre `...guactt` : aucune occurrence), alors qu'un `U` du motif correspond à `t`. Cette règle coïncide avec celle d'ecoPrimers et de `MatchWords` pour les motifs ACGT (sauf `u` ↔ `t`). Les majuscules de la séquence sont acceptées (conversion côté Go).
- L'ancienne version de ce rapport (§11) affirmait qu'un `N` de la séquence correspond à tout : c'était **faux**. Non testé : la négation `!X`, dont le complément est calculé sur 26 bits et pourrait donc accepter `n`, `r`, `y`...

**Coût** : l'alphabet de 26 lettres n'intervient dans le balayage que par la table `smat[26]` indexée par l'octet encodé de la séquence ; la complexité de `ManberSub` est `O(n·(e+1))` mots, indépendante de l'alphabet. Il coûte seulement un passage d'encodage par séquence (`EncodeSequence`, copie en indices 0..25, plus 64 octets de bourrage si circulaire) et la prétable `CreateS` en `O(m·26)`. Les vrais coûts de l'implémentation : les états `r[]` vivent en mémoire (tableau de `2·MAX_PAT_ERR+2` mots, boucle sur `e` à borne dynamique, aucun état en registre), un seul motif par balayage (`MAX_PATTERN = 1` : un passage de texte par motif et par brin), et le passage par cgo (11 % du profil sur Teleostei). L'alphabet général gêne en revanche la vectorisation : une table de 26 entrées par voie exige un `gather` ou deux permutations d'octets, contre un mélange de 4 à 5 valeurs pour un alphabet ADN réduit.

**Longueur maximale (vérifiée par exécution)** : `MAX_PAT_LEN` vaut 64 mais la limite effective de `obiapat` est **63 symboles** (le balayage utilise le bit `patlen` comme garde, `1 << patlen`). `MakeApatPattern` accepte sans erreur des motifs plus longs : à 64 symboles le motif ne trouve **rien**, même dans une copie exacte de lui-même ; à 65 et 70 il renvoie des occurrences avec des **positions négatives**. Aucun message d'erreur. Un remplaçant doit refuser au-delà de 63.

**Remplaçant en cours** : `pkg/obidnapat` (Go pur, ADN, substitutions seules pour l'instant), voir `OBIDNAPAT_NOTES.md`.
