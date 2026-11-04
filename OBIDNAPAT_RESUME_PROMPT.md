# Prompt de reprise : `obidnapat` (matching ADN en Go pur) et état d'`obiprimers`

Tu travailles dans le dépôt Go `/Users/coissac/Sync/travail/__MOI__/GO/obitools4` (module
`git.metabarcoding.org/obitools/obitools4/obitools4`). Go : `/usr/local/go/bin/go` (**absent du `PATH`**). Machine :
MacBook M3 Max, 16 cœurs, macOS arm64, load average souvent de 70 à 190 à cause d'autres processus : **les temps sont
bruités**, compare des séries entrelacées et fie-toi aux rapports, aux nombres de paires et à la mémoire.

Ce prompt reprend deux chantiers. **Le chantier principal est `obidnapat`** ; `obiprimers` est le contexte d'où il vient et
garde un blocage ouvert.

## 0. Règles de travail (de l'utilisateur, non négociables)

- **Aucune commande git qui modifie quoi que ce soit** (commit, add, mv, rm, reset, push...) sans demande explicite dans
  le même tour. Lecture seule (`status`, `diff`, `log`) : toujours permise. Pour déplacer ou supprimer un fichier, utilise
  `mv`/`rm`/`Write`/`Edit`.
- **Pas de code pendant une discussion de spécification** : réponds, propose, demande, et n'implémente qu'après un feu vert
  explicite. Un feu vert vaut pour ce qu'il nomme.
- **Pas d'approximation silencieuse** sur du code scientifique : toute perte de résultat se demande d'abord, avec une mesure
  (rappel contre l'exhaustif). L'utilisateur a dit : « il faut une version sans perte ».
- **Pas de spéculation présentée comme un fait** : vérifie (lis le code, mesure, teste) ou dis que c'est une hypothèse.
  Quand tu te trompes, dis-le tout de suite (j'ai dû corriger deux fois des affirmations sur `obiapat` après un test).
- Fichiers source de moins de ~500 lignes (le code généré aussi, on le sépare). Enrichir l'existant plutôt que du code local.
- zsh ne découpe pas une variable non guillemetée : écris les scripts de mesure avec `bash -c`. `pkill -f motif` tue aussi
  le shell qui contient le motif. `go test -bench 'a/b'` découpe l'expression sur `/`. BSD `sed -i` exige `-i ''`.
- Les tests passent d'abord par un **oracle indépendant** ; `obiapat` sert de second oracle sur le terrain commun.

## 1. Lis d'abord

1. `OBIDNAPAT_NOTES.md` (état, décisions, mesures, essais écartés, suite).
2. `OBIPRIMERS_NOTES.md` : annexe D (sémantique **vérifiée** d'`obiapat`), sections 6 à 8 (blocage d'`obiprimers`, essais,
   pistes).
3. La mémoire du projet (`~/.claude/projects/-Users-coissac-Sync-travail---MOI---GO-obitools4/memory/`), surtout
   `project_obidnapat.md` et `project_obiprimers_scale_blocker.md`.
4. `git status` et `git log --oneline -5` (lecture seule) : l'utilisateur commite lui-même, par morceaux ; ne suppose pas
   l'état du dépôt, regarde-le. Les fichiers `OBIDNAPAT_NOTES.md`, `OBIDNAPAT_RESUME_PROMPT.md` et le commentaire de
   `gen_kernels.go` sur la variante abandonnée étaient les derniers changements, peut-être non commités.

## 2. `obidnapat` : ce qui existe

Paquet `pkg/obidnapat` : remplaçant Go pur, spécialisé ADN, d'`pkg/obiapat` (cgo, agrep). **Aucun appelant n'est migré.**
`obiapat` reste en place jusqu'à la fin.

Décisions de l'utilisateur : cœur de 63 positions au plus (jokers `N` de tête et de queue mis à part, un motif plus long est
**refusé**) ; moteur ADN ; règle de **recoupement** pour les lettres ambiguës de la séquence (`n` correspond à tout, `r` à
A ou G, `u` = `t`) et pas de mode « ancien » (l'ancienne règle est un défaut) ; un octet hors IUPAC ne correspond qu'à une
position joker ; plus d'`ApatSequence` : méthodes sur `*obiseq.BioSequence`, variantes `...Circular`, formes `Append...` ;
`NNNNmotifNNNN` recodé (cœur + flancs) ; circulaire : une occurrence par site, `start ∈ [0, n)`, `end` pouvant dépasser `n`,
motif plus long que la séquence circulaire = aucun résultat ; SIMD **réservé au multi-motifs** ; garder une voie stricte.

Fichiers : `syntax.go`, `pattern.go` (`MakeApatPattern`, `ReverseComplement`...), `scan.go` (Shift-And de Wu-Manber, `e = 0`
à la main, générique à tableau pour `e ≥ 9`), `scan_unrolled.go` (**généré** par `gen_kernels.go`, `go generate ./pkg/obidnapat/`,
`e = 1..8` à états en registres), `strict.go` (BNDM, choisi si `advance() ≥ minAdvance = 3,6`, séquence linéaire seulement),
`search.go` (API). `MakeApatPattern(..., allowsIndel=true)` renvoie une erreur : **les indels ne sont pas portés**.

Tests : `oracle_test.go` (oracle naïf + `TestAgainstOracle` + `TestStrictAgainstOracle`), `diff_test.go` (contre `obiapat` sur
séquences ACGT), `pattern_test.go`, `bench_test.go`. `OBIDNAPAT_ITER=300000 go test -race ./pkg/obidnapat/` : **tous verts**
(`e` de 0 à 10). Mesures (un cœur, motif de 20, 1 Mb) : `e = 1..3` ×2,6-2,8 sur `obiapat`, `e = 4..8` ×2,6-3,5, strict ×2,3
(20 positions) jusqu'à ×8-10 (50-63 positions), lectures de 150 bases ×3,3-3,7. Détails et essais écartés dans les notes.

**Ne pas refaire** : noyaux sans masque `allow` (plus lents à cause du compilateur), noyau générique à tableau pour `e ≤ 8`,
BNDM sans critère d'avance.

## 3. Suite d'`obidnapat`, par ordre

1. **`PatternSet` multi-motifs, deux brins d'un coup.** Plusieurs motifs concaténés dans un mot de 64 bits (le décalage qui passe
   d'un motif au suivant est inoffensif car le bit de départ est forcé à 1 par `| départ` ; cela reste à **démontrer par le
   test**), masque de fin par motif et par niveau, occurrences étiquetées `{motif, brin, début, fin, erreurs}`, listes par
   motif et par brin identiques à celles de scans séparés (pas de dédoublonnage des palindromes). Au-delà de 64 symboles :
   plusieurs mots. **Chaîne d'oracles** : oracle naïf → noyau mono-motif → version empaquetée, y compris des motifs qui
   totalisent 63, 64 et 65 positions. Puis **SIMD par voies de motifs** (`simd` de Go 1.27, `GOEXPERIMENT=simd`, repli
   scalaire `*_nosimd.go` comme dans `pkg/obikmer`) : les tables en `[octet][voie]` n'exigent aucun `gather`.
   Critère de décision : mesurer contre le scan séparé et contre `MatchWords` (le filtre par graines d'`obiprimer`, meilleur
   pour des centaines de motifs).
2. **Voie stricte multi-motifs** : Wu-Manber (décalage par q-gramme) ou Aho-Corasick ; l'utilisateur attend qu'elle batte
   Manber, notamment en multi-motifs.
3. **BNDM sur séquence circulaire** (aujourd'hui Shift-And) ; Shift-And à plusieurs flux pour les motifs courts ; mesurer les
   séquences riches en `N` (un `n` correspond à tout, BNDM ne s'arrête plus) et les textes à nombreuses occurrences ;
   recaler `minAdvance` sur d'autres machines.
4. **Port de `ManberIndel`** (≈60 lignes de C, `apat_search.c`) avec la logique de `obialign.LocatePattern` **inchangée**
   (la mémoire signale un bug ouvert : garde « motif ≥ fragment » qui masque un décalage d'un). Nécessaire à `obiannotate` et
   `obingslibrary`. Vérifier le comportement « may return shifted pos » par test différentiel.
5. **Migration des appelants**, un par un, une fois le test différentiel vert : `pkg/obitools/obipcr` + `pkg/obiapat/pcr.go`
   (circulaire), `obisplit`, `obigrep` (`predicat.go`), `obiannotate`, `pkg/obingslibrary` (`marker.go`, `multimatch.go`).
   Transformation : `MakeApatSequence(s, false)` puis `p.AllMatches(aseq, b, l)` devient `p.AllMatches(s, b, l)`. Rejouer les
   tests de chaque consommateur. Ne supprimer `obiapat` qu'à la fin et sur demande.

## 4. `obiprimers` : état et blocage

Mode A (conservé) complet, validé contre ecoPrimers (tests opt-in : `OBIPRIMERS_TUTORIAL="$HOME/Documents/Enseignement/DNA
metabarcoding spring school/ecoPrimers-Fish" go test -count=1 -v -run 'AgainstEcoPrimers|SitesAreRealSites' ./pkg/obiprimer/`,
≈64 s, verts). Définitions : Bc, Bs (taxons identifiés / amplifiés, **tous** les barcodes uniques parmi tout ce qui est amplifié),
Bd. **Blocage** : sur les plastes de Spermatophyta (`plastid/plastid_nov_2026_taxo.fasta.gz`, `-r 58024 -c -l 10 -L 300`) la
phase 2 doit traiter 1 407 309 paires et 30,4 milliards d'amplicons en 1 522 lots, chacun repassant sur tout le corpus ; aucun run
n'a abouti. Ce run utilisait `-3 0`. Pistes **exactes** non faites : relancer avec `-3 2` ; retirer des occurrences les sites dont
les bases 3' sont en erreur après le calcul de `ex`/`co` ; énumérer à partir des amorces « ancres » par lot ; matrice de bits
amorce×amorce avant la table des paires ; hash de barcode par préfixes ; avertissement quand le total d'amplicons est énorme ;
supprimer `forEachOccurrence` (code mort). **Rejetés après mesure** : élagage des graines par contre-exemples, grossier-fin par
superprimers (rappel 32-93 %), sous-échantillonnage comme borne de Bs. Détails : `OBIPRIMERS_NOTES.md` §6-8 et annexes.
`obidnapat` pourra servir à `obiprimers` (évaluation d'amorces dégénérées, phase de Tm).

## 5. Commandes

```
G=/usr/local/go/bin/go
$G generate ./pkg/obidnapat/                               # régénère scan_unrolled.go
$G vet ./pkg/obidnapat/ && gofmt -l pkg/obidnapat/
$G test -count=1 ./pkg/obidnapat/                          # rapide
OBIDNAPAT_ITER=300000 $G test -count=1 -race ./pkg/obidnapat/ -timeout 3000s   # campagne complète
$G test -run '^$' -bench 'Scan|Strict|Reads|Flanks' -benchtime 60x -count 3 ./pkg/obidnapat/
```

Première action suggérée : lire les notes, lancer la campagne de tests pour confirmer l'état, puis te faire donner le feu vert
sur l'étape 1 (`PatternSet`) en proposant l'API et les tests avant d'écrire le moindre noyau.
