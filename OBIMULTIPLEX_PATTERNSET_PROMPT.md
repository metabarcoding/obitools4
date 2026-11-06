# Prompt de reprise : recherche simultanée des amorces dans `obimultiplex` (`PatternSet` avec indels)

Tu travailles dans le dépôt Go `/Users/coissac/Sync/travail/__MOI__/GO/obitools4` (module
`git.metabarcoding.org/obitools/obitools4/obitools4`). Go : `/usr/local/go/bin/go` (**absent du `PATH`**) ; lance toujours
`go build`/`go test` **depuis le dépôt** (hors du module, `go build` échoue). Machine : MacBook M3 Max, 16 cœurs, macOS arm64, charge
souvent élevée : les temps sont bruités, compare des séries entrelacées. La documentation est dans un autre dépôt,
`/Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc` (Hugo ; `hugo --buildDrafts --config hugo.yaml --destination /tmp/hugo_out --quiet`
pour vérifier une page ; il est versionné avec `jj`).

**Objectif de la session, de bout en bout** : faire chercher **toutes les amorces de tous les marqueurs d'un coup** par
`obimultiplex`, mode indels compris, avec `obidnapat.PatternSet`, au lieu des 4 recherches `AllMatches` par marqueur de
`ExtractMultiBarcode`. L'utilisateur a donné son accord de principe pour l'implémentation (« oui, quand on commence à travailler
avec des multiplexes, c'est un gain réel »), mais voir §6 : propose l'API et le plan de tests **avant** d'écrire le moindre noyau.

## 0. Règles de travail (de l'utilisateur, non négociables)

- **Aucune commande VCS qui modifie quoi que ce soit** (git, jj : commit, add, mv, rm, reset, describe, push...) sans demande explicite dans
  le même tour. Lecture seule (`status`, `diff`, `log`) : permise. Déplacer ou supprimer un fichier : `mv`/`rm`/`Write`/`Edit`. L'utilisateur commite lui-même.
- **Pas de code pendant une discussion de spécification** : propose, demande, n'implémente qu'après un feu vert explicite, qui vaut pour ce qu'il nomme.
- **Pas d'approximation silencieuse** sur du code scientifique ; pas de spéculation présentée comme un fait : vérifie (lis le code, mesure, teste) ou
  dis que c'est une hypothèse. Quand tu te trompes, dis-le tout de suite (c'est arrivé plusieurs fois dans la session précédente).
- **Règles simples, pas de réglages ad hoc sur un jeu de données** : face à un choix sans vérité connue (quelle fenêtre parmi des alignements de même
  coût), une règle simple, arbitraire et documentée. Ce qu'on fait n'est que de l'inférence.
- **La logique des tags n'a rien à faire dans la bibliothèque de recherche** (elle dépend des bibliothèques d'échantillons).
- Fichiers source de moins de ~500 lignes (le code généré est séparé). Enrichir l'existant plutôt que du code local. Supprimer le code rendu inutile
  par une décision plutôt que laisser deux chemins coexister.
- **Tests d'abord par un oracle indépendant** ; toute nouvelle logique est vérifiée par **mutation** (casse-la volontairement et vérifie que les tests échouent).
- zsh ne découpe pas une variable non guillemetée : scripts de mesure avec `#!/bin/bash` (`bash script.sh`). BSD `sed -i ''` ; l'adresse `0,/re/` n'existe pas ;
  `pkill -f motif` tue aussi le shell qui le contient.
- L'utilisateur **dicte** (transcription automatique) : « hobby » = obi, « Hall match » = `AllMatches`, « primers 7 » = `PatternSet`. Interprète avec bon sens, confirme si c'est important.
- **`grep -c '^@'` compte aussi des lignes de qualité FASTQ** : pour compter les séquences, lis les enregistrements par quatre lignes.
- Vérifie qu'un remplacement de texte (`python3 str.replace`, `Edit`) a vraiment eu lieu (`grep` après) : il est arrivé que ce ne soit pas le cas sans erreur.

## 1. Lis d'abord

1. `OBIDNAPAT_NOTES.md` (≈ 390 lignes). Sections utiles : « `PatternSet` : décisions », « noyaux entrelacés branchés », « Indels : port de `ManberIndel` »,
   « Vrai début sans programmation dynamique », « `AllMatches` à début réel », « Migration d'`obingslibrary` », « Simplification : un seul chemin »,
   « `obimultiplex` : recherche simultanée des amorces avec `PatternSet` ? Profil CPU », « tags distincts », « `PatternSet` avec indels : est-ce faisable ? Prototype ».
2. La mémoire du projet (`~/.claude/projects/-Users-coissac-Sync-travail---MOI---GO-obitools4/memory/`) : `project_obidnapat.md`, `feedback_simple_rules_not_dataset_tuned.md`.
3. `git status` et `git log --oneline -5` (lecture seule) : l'état du dépôt peut avoir changé (l'utilisateur commite par morceaux). À la fin de la session précédente,
   seuls `OBIDNAPAT_NOTES.md` (modifié), `pkg/obidnapat/proto_indel_set_test.go` et ce fichier étaient non commités.
4. `OBIDNAPAT_RESUME_PROMPT.md` est le **vieux** prompt (début de projet) : l'état y est périmé, ne t'y fie pas.

## 2. Ce qui existe (état de `pkg/obidnapat`)

Remplaçant Go pur, spécialisé ADN, d'`pkg/obiapat` (cgo). Migrés et vérifiés : `obiannotate`, `obigrep`, `obingslibrary`/`obimultiplex`. `obisplit` a été supprimé
(décision de l'utilisateur). `obipcr` est migré aussi (moteur dans `pkg/obidnapat/pcr.go`) : plus aucun programme ne dépend d'`obiapat`, qui n'est plus importé que par des tests différentiels ; sa suppression se fait sur demande de l'utilisateur. Décision en attente : le traitement d'un `n` de la séquence (voir `OBIDNAPAT_NOTES.md`, « Migration d'`obipcr` »).

- `pattern.go` : `MakeApatPattern(pattern, errormax, allowsIndel)`. `compiled.indel = allowsIndel && errormax > 0` ; avec indels le **motif entier** est le cœur (pas de retrait des `N` de tête
  et de queue), limite de 63 positions ; `levels = min(errormax, 63)` (les erreurs peuvent dépasser la longueur du motif à cause des insertions) ; sans indels le cœur exclut les `N` de tête/queue (`left`, `right`), `levels = min(errormax, coreLen)`.
  `compiled` porte `mask [256]uint64` (bit i : l'octet correspond à la position i), `end`, `allow` (bit i : une erreur est permise à la position i, interdite par `#`), `coreLen`, `positions`.
- `scan_indel.go` : récurrence de Wu-Manber avec indels (`runIndelN` à tableau, `indelHits` = fenêtres **nominales** `[pos-len+1, pos+1)` de chaque fin, **sans filtre** sur le début nominal, état initial `o_j = 2^j - 1`),
  `searchIndel` (= `indelHits` + filtre début nominal >= début de région) ; `scan_indel_unrolled.go` (généré, `runIndel1..8`). `scan_backward.go` : `backwardStart(seq, pos, e, lo)` = vrai début (le plus tardif) par la même
  recherche à l'envers, sans matrice.
- `search.go` : **un seul chemin** pour trouver une fenêtre : `AllMatches(seq, begin, length)` (pour un motif à indels : `bestOfOverlaps(indelHits(...))`, puis pour chaque fenêtre avec erreurs `backwardStart` ; la région s'applique au
  **vrai** début ; début nominal conservé si le vrai est introuvable, ce qui n'arrive qu'avec `#` en tête de motif ; `AllMatches` ne traite que les séquences linéaires ; sans indels : `FilterBestMatch`). `BestMatch` = première fenêtre de plus petit nombre d'erreurs d'`AllMatches`.
  `bestOfOverlaps` : deux fenêtres se chevauchent si `(m.début - m.err) < best.fin + best.err` ; garde la première de plus petit nombre d'erreurs. `obialign.LocatePattern` n'est plus utilisé par ce paquet (supprimé, décision de l'utilisateur).
- `noise.go` : **règle des bases réelles** (2026-11-06) : une fenêtre n'est rapportée que si ses bases non-`N` sont **plus nombreuses que ses erreurs** (le motif ne doit pas être apparié par ses seules erreurs), appliquée à chaque occurrence avant tout regroupement, dans `noise.go` (voir `OBIDNAPAT_NOTES.md`, « Règle des bases réelles »). Un motif seul : `hasN` puis filtre par blocs de 4096 octets (`dropNoise`), actif seulement si la séquence a un `N`, si le motif a des indels ou si `levels >= Len()` ; `PatternSet` : filtre dans `word.emit`. Les tests qui comparent à `obiapat` (sans cette règle) écartent les motifs avec `e >= Len()`.
- `patternset.go`, `patternset_scan.go`, `patternset_unrolled.go`, `patternset_il*.go`, `gen_kernels.go` : `PatternSet` pour les motifs **sans indels** : mots de 64 bits contenant plusieurs cœurs (empaquetage first-fit decreasing, **par classe de
  niveaux**), récurrence avec `| start` (bit de départ de chaque cœur) à la place de `| 1`, masque de fin par cœur, `Occurrence{Pattern, Strand, Start, End, Errors}`, tri final par `(Start, Pattern, brin +1 d'abord)`, liste « solo » (un cœur
  seul à e = 0 avec BNDM reste un scan autonome), noyaux déroulés de 0 à 8 niveaux, noyaux **entrelacés** (2 ou 3 mots de même niveau dans une boucle : ×1,5 à ×1,65 ; `batchFor` : 3 mots jusqu'à 3 niveaux, 2 à 4-5, 1 au-delà), prototype SIMD
  (`makePatternSet(..., vector=true)`, build `goexperiment.simd`) **désactivé par défaut** (2 voies NEON : +5 %, l'entrelacement scalaire fait mieux). **`MakePatternSet` refuse les motifs avec indels** (garde ajoutée en fin de session ;
  elle les acceptait avant et appliquait la mauvaise récurrence).
- `proto_indel_set_test.go` : **prototype validé** de la récurrence empaquetée avec indels (voir §3), code de test seulement.
- Tests : `oracle_test.go` (oracle naïf), `diff_test.go` (contre `obiapat`), `indel_test.go` (`oracleIndel`, `truthStart`, `expectedAllMatches`, `TestBestMatch`, comparaisons à `obiapat`), `patternset_test.go`, `ambiguity_test.go` et
  `realdata_test.go` (mesures sur de vraies lectures, ignorées sans variables d'environnement), `bench*_test.go`.

## 3. Ce qui est établi pour cette tâche

**Récurrence empaquetée avec indels (validée par prototype).** Pour un mot de plusieurs cœurs, `start` = bits de départ de tous les cœurs, `A` = `allow` empaqueté, `mk` = masque de l'octet :

```
n_0 = ((o_0 << 1) | start) & mk
n_j = (((o_j << 1) | start) & mk) | ((o_{j-1} | ((o_{j-1} << 1) | start) | (n_{j-1} << 1)) & A)
état initial  o_j = Σ cœurs  ((2^j - 1) & plein_du_cœur) << décalage
```
Un cœur déclare une occurrence quand le bit de fin d'un niveau <= `levels` du cœur est posé (plus petit niveau). `TestProtoPackedIndel` : 7 474 123 occurrences identiques à la recherche d'un motif seul sur 20 000 mots à plusieurs
motifs (1 à 22 positions flancs compris, e de 1 à 5, `#`, ambiguïtés, `N`) ; mutations détectées (bit de départ du seul premier cœur) ; le bit qui passe d'un cœur au suivant par le terme de suppression est inoffensif (le retirer ne change rien).

**Profil de `ExtractMultiBarcode`** (20 000 lectures MinION, un cœur, `.ngsfilter` à **un seul marqueur**, 100 échantillons, `matching,indel`, `tag_mismatches,1`, `primer_mismatches,5`, indels permis), après la correction des tags distincts :
recherche des amorces (`AllMatches`) **32 %**, `SampleIdentifier` 24 %, affectations d'annotations (`mapassign`) ~21 %, `Subsequence` 13 %, `fmt.Sprintf` 8 %. Avant la correction des tags, la recherche des amorces ne pesait que 7,7 %. Gain attendu de `PatternSet` avec indels
(**non mesuré**) : 4 motifs de 20-22 positions tiennent dans 2 mots (3 dans le premier), ×2 à ×3 sur la recherche, soit ~20 % d'`ExtractMultiBarcode` avec un seul marqueur, davantage avec plusieurs marqueurs (c'est l'objectif).

**Comportement actuel d'`ExtractMultiBarcode`** (`pkg/obingslibrary/multimatch.go`) : pour chaque marqueur `i` (table `library.Markers`), `forward.AllMatches(seq, 0, -1)` ; s'il y a des fenêtres, `creverse.AllMatches(seq, premier début + 1, -1)` ;
idem `reverse` puis `cforward` ; les fenêtres deviennent des `PrimerMatch{Begin, End, Mismatches, Marker (±i), Forward}`, triées par `Begin` (**tri non stable**), puis un **automate à deux états** : état 0 attend un jeton positif (`Marker > 0`), état 1
attend le jeton `-Marker` de même `Forward` (sinon on repart sur un autre jeton positif ou on retombe à 0). Le tag est ensuite lu autour de `Begin`/`End` (`TagExtractor`, `lookForTag`). Le motif `creverse` n'est cherché qu'à partir du premier
`forward` : avec une recherche conjointe cette précaution de coût disparaît, l'automate ignore déjà les jetons négatifs en état 0 (à vérifier aux bords : le regroupement des chevauchements dans la région tronquée peut différer).
Chaque marqueur a ses propres nombres d'erreurs (`Forward_error`, `Reverse_error`) et ses propres indels (`Forward_allows_indels`, `Reverse_allows_indels`) : un jeu mélange des motifs avec et sans indels, avec des niveaux différents.

## 4. À faire, dans l'ordre

1. **Reproduire l'état** : tests (§7), construire le binaire de **référence** avant de toucher à `obingslibrary` (`go build -o /tmp/obimultiplex_ref ./cmd/obitools/obimultiplex`, depuis le dépôt), préparer les données (§5), mesurer les temps de référence.
2. **Proposer à l'utilisateur** l'API et le plan de tests (§6), attendre son accord.
3. **Noyaux de mots avec indels** : étendre `gen_kernels.go` (déroulés de 1 à 8 niveaux, entrelacés par lots comme pour les substitutions avec les mêmes `batchFor` à remesurer, noyau à tableau au-delà de 8). Un mot est d'une seule sorte : classes d'empaquetage
   par `(avec indels ou non, niveaux)`. Garder chaque fichier sous ~500 lignes.
4. **`PatternSet` avec indels** : `MakePatternSet` accepte les motifs avec indels ; sortie par (motif, brin) = fenêtres au sens d'`AllMatches` (regroupement des chevauchements par motif et brin, vrai début par `backwardStart` avec les tables de chaque motif,
   région appliquée au vrai début), toutes triées par début. Retirer le refus ajouté en fin de session et son test (`TestPatternSetRefusesIndels`) en le remplaçant par les vrais tests.
5. **Tests** : (a) jeux aléatoires de motifs avec et sans indels, e de 0 à 10, deux brins, régions, contre `AllMatches` motif par motif et contre `expectedAllMatches` (oracle indépendant) ; (b) totaux de 63/64/65 positions par mot ; (c) mutation de chaque noyau généré
   et de la logique de regroupement/vrai début ; (d) `-race`. S'appuyer sur `proto_indel_set_test.go`.
6. **Restructurer `ExtractMultiBarcode`** : construire le `PatternSet` **une fois** par bibliothèque (dans `Compile2`, il est immuable donc partageable entre les goroutines), une seule recherche par lecture, les jetons alimentent directement l'automate.
   Conserver la sémantique (mêmes fenêtres que les `AllMatches` actuels) ; l'ordre des jetons de même début change (le tri actuel est instable) : vérifier et expliquer toute différence.
7. **Vérifier de bout en bout** (§8), mesurer (§9), mettre à jour les notes et la mémoire. Ne rien commiter.

## 5. Données et outils de mesure

- Lectures MinION réelles : `/Users/coissac/travail/MinIon/Metabarcoding/nanopore_test_plate_A.fastq.gz` (1,9 Go). Lire avec `gzip -dc` (le `zcat` de macOS cherche un `.Z`). Par exemple
  `gzip -dc ... | head -1600000 > /tmp/annot/r400k.fastq` (400 000 lectures). Médiane ~160 bases ; les amorces sont dans les deux orientations, plusieurs amplicons par lecture sont fréquents.
- Description des échantillons : `/Users/coissac/travail/MinIon/Metabarcoding/plate1_pool-SPER05.Pierre.ngsfilter` : **l'utilisateur la modifie pendant les sessions ; travaille sur une copie figée** (`cp ... /tmp/annot/ngs.tab`). À la fin de la session précédente : `primer_mismatches,5`, `indels,true`,
  délimiteurs `G`, `matching,indel`, `tag_mismatches,1`, 100 échantillons, **un seul marqueur** (amorce directe `GGGCAATCCTGAGCCAAATC`, inverse `CTTCCRTTGAGTCTCTGCACCT`, tags de 7 bases).
- **Il n'existe aucune bibliothèque multi-marqueurs** : l'objectif est justement le multiplexage. Fabrique-en une synthétique (N paires d'amorces, par exemple 4, 10 et 40 marqueurs, amorces aléatoires de 18 à 25 bases, tags, lectures simulées avec amplicons
  encadrés par les amorces et des substitutions/indels) pour valider l'équivalence et mesurer le gain. Mesure aussi sur l'unique marqueur réel.
- Banc : `BenchmarkExtractMultiBarcode` (`pkg/obingslibrary/realdata_test.go`) avec `OBIDNAPAT_FASTQ`, `OBIDNAPAT_NGSFILTER`, `OBIDNAPAT_READS` :
  `OBIDNAPAT_FASTQ=... OBIDNAPAT_NGSFILTER=... OBIDNAPAT_READS=20000 go test -run '^$' -bench ExtractMultiBarcode -benchtime 10x -cpuprofile cpu.prof -o ngslib.test ./pkg/obingslibrary/` puis `go tool pprof -top -cum ngslib.test cpu.prof`.
- Mesures sur de vraies lectures dans `pkg/obidnapat/realdata_test.go` (`loadRealReads`, variables `OBIDNAPAT_FASTQ`, `OBIDNAPAT_PRIMERS="fwd,rev"`, `OBIDNAPAT_READS`). Banc des jeux : `BenchmarkSet`, `BenchmarkSetLen`, `BenchmarkBatch` (`bench_set_test.go`) ; en écrire un pour des jeux avec indels.
- Comparaison des sorties d'`obimultiplex` (ancien binaire contre nouveau), triées :
  `norm(){ paste - - - - < "$1" | sort; }` puis `comm -3 <(norm a.fastq) <(norm b.fastq) | wc -l` (sorties attribuées `-o`/stdout et non identifiées `-u ... --keep-errors`). Attention : lancer avec `bash`.
  Options utiles : `--max-cpu 1` pour un temps sur un cœur.

## 6. Décisions à proposer à l'utilisateur avant d'écrire (avec recommandations)

1. **API de sortie.** Recommandation : garder `AppendAll` (occurrences brutes, un par début, pour les motifs sans indels) et ajouter `AppendAllMatches` = sémantique d'`AllMatches` par (motif, brin) pour **tous** les motifs, avec ou sans indels
   (`Occurrence` avec `Start` = vrai début pour les indels). C'est ce que consomme `ExtractMultiBarcode`.
2. **Jeux mixtes** (motifs avec et sans indels, niveaux différents par amorce) : classes d'empaquetage `(indels, niveaux)`.
3. **Ordre des jetons de même début** : `(Start, Pattern, brin)` déterministe, plutôt que le tri instable d'aujourd'hui.
4. **Où vit le `PatternSet`** de la bibliothèque : champ de `NGSLibrary` construit dans `Compile2`, avec la correspondance entrée -> (marqueur, rôle) pour reconstruire les `PrimerMatch`.
5. Séquences circulaires avec indels : pas d'`AllMatches` circulaire aujourd'hui ; non nécessaire ici (obimultiplex est linéaire).

## 7. Commandes

```
G=/usr/local/go/bin/go
$G generate ./pkg/obidnapat/                                   # régénère les noyaux
$G vet ./pkg/obidnapat/ ./pkg/obingslibrary/ && gofmt -l pkg/obidnapat pkg/obingslibrary pkg/obiformats pkg/obitools/obimultiplex
$G build ./... 2>&1 | grep -E ' error|undefined|cannot'       # les avertissements C de obiformats/fastseq_read.h sont anciens et sans gravité
$G test -count=1 ./pkg/obidnapat/ ./pkg/obingslibrary/ ./pkg/obiformats/
OBIDNAPAT_ITER=6000 $G test -count=1 -race ./pkg/obidnapat/ -timeout 590s   # ~4 minutes ; OBIDNAPAT_ITER règle le nombre de tirages
GOEXPERIMENT=simd $G test -count=1 ./pkg/obidnapat/            # le build simd doit rester vert (prototype désactivé par défaut)
$G test -run '^$' -bench 'Set|Batch' -benchtime 5x -count 2 ./pkg/obidnapat/
```

## 8. Critères de réussite

- Tous les tests verts, `-race` vert, build simd vert, `gofmt` et `vet` propres, fichiers < ~500 lignes.
- Mutations détectées pour chaque noyau généré et pour le regroupement / le vrai début dans le jeu.
- **Sorties d'`obimultiplex` identiques** (enregistrement par enregistrement, attribués et non identifiés) entre le binaire de référence et le nouveau, sur 400 000 lectures réelles avec la copie figée du `.ngsfilter`, pour plusieurs `-e` et avec/sans
  `--with-indels`, et sur des bibliothèques synthétiques multi-marqueurs. Toute différence expliquée (ordre des jetons de même début) et quantifiée, jamais passée sous silence.
- Gain **mesuré** (profil + temps de bout en bout sur un cœur) pour 1, 4, 10 et 40 marqueurs ; ne rapporte que des mesures, avec leur bruit.
- Notes (`OBIDNAPAT_NOTES.md`) et mémoire à jour, y compris ce qui est écarté et pourquoi.

## 9. Pièges déjà rencontrés

- La règle des bases réelles (`noise.go`) rejette des fenêtres que les noyaux produisent : tout nouveau chemin d'émission d'occurrences doit l'appliquer (`word.emit` pour `PatternSet`, `dropNoise` pour un motif seul), sinon une suite de `N` redonne une occurrence par position. Ne pas modifier les noyaux générés pour cela.
- Les comptes de séquences par `grep -c '^@'` sont faux (lignes de qualité commençant par `@`).
- Le `.ngsfilter` change pendant que tu travailles : copie figée.
- `go build` hors du dépôt échoue (et un `grep` de filtre peut masquer l'erreur) : toujours depuis le dépôt.
- Ajouter un champ à une structure fait que `gofmt` réaligne toute la structure ; vérifier avec `gofmt -l`.
- Une erreur interne du compilateur Go 1.27 (`missing Types entry`) quand du code `simd` est dans une **méthode** : en faire une fonction ordinaire à arguments simples.
- `ShiftAllLeft(1)` de `simd` compile mal sur arm64 (utiliser `x.Add(x)`).
- Les noyaux générés sont longs : `gen_kernels.go` écrit plusieurs fichiers (`scan_unrolled.go`, `scan_indel_unrolled.go`, `patternset_unrolled.go`, `patternset_il*.go`) ; ne les édite jamais à la main.
- Une fenêtre avec suppressions est plus courte que le motif : son début nominal peut précéder la région, même si le vrai début est dedans. Ne pas filtrer sur le début nominal pour `AllMatches`.
- La fin d'une fenêtre est celle d'**un** alignement de coût minimal, pas nécessairement la fin réelle (plusieurs fins voisines à égalité) ; règle retenue : la première, simple et documentée.
- Les rapports de vitesse très élevés ou erratiques des motifs de 10 positions avec e >= 2 mesurent surtout l'émission et le tri d'un très grand nombre d'occurrences : ne pas les interpréter.
