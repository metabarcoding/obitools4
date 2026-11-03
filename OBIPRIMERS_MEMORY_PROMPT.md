# Prompt : optimisation de l'usage mémoire de `obiprimers`

Tu travailles dans le dépôt Go `/Users/coissac/Sync/travail/__MOI__/GO/obitools4` (module
`git.metabarcoding.org/obitools/obitools4/obitools4`, Go : `/usr/local/go/bin/go`, non présent dans le PATH).
Ta mission : **réduire l'usage mémoire de la commande `obiprimers`** sans changer ses résultats, en procédant
par la mesure.

## 1. Contexte

`obiprimers` est la ré-implémentation OBITools4 d'ecoPrimers (design de paires d'amorces à partir de génomes
entiers, avec évaluation de la qualité du marqueur délimité). Le mode « voie A » est terminé et validé.
Avant toute chose, **lis `OBIPRIMERS_NOTES.md`** (racine du dépôt) : la section « État courant », puis les
sections 11 à 13 (appariement approché), 16 et 16 bis (taxonomie embarquée ; **une erreur de mesure à ne pas
refaire**), 25 (définitions de Bc, Bs, Bd et élagage de la phase 2).

Code concerné :
- `cmd/obitools/obiprimers/main.go`, `pkg/obitools/obiprimers/` (options, lecture, sortie CSV) ;
- `pkg/obiprimer/` : `corpus.go` (`LoadCorpus`), `taxa.go`, `words.go` (`FindStrictWords`), `match.go` et
  `match_seed.go` (`MatchWords`), `pairs.go` (`selectPrimers`, `countPairs`), `pairs_eval.go` (`FindPairs`),
  `amplicons.go` (`collectAmplicons`), `resolution.go`, `search.go` (`Search`) ;
- `pkg/obikmer/seqquorum*.go` (`SeqKmerCounter`, comptage par partitions sur disque) ;
- `pkg/obithermo/` (Tm), code partagé : `pkg/obiformats`, `pkg/obiiter`, `pkg/obiseq`, `pkg/obitax`.

## 2. Décisions de l'utilisateur à respecter (non négociables)

- **Tout est chargé en mémoire** : toutes les séquences de la base (cibles, contre-exemples, non-cibles) sont
  gardées, parce que la lecture sur l'**entrée standard** doit continuer à marcher comme pour tout obitool.
  Tu peux **compacter** ce qui est gardé ; tu ne peux pas passer à un schéma qui exige de relire le fichier.
- **Un seul calcul par indice** : ne touche pas aux définitions de Bc, Bs et Bd (`resolution.go`). Pas de
  colonne ou d'option alternative.
- **Les résultats ne changent pas.** Critère : la sortie CSV doit rester **identique octet pour octet** sur les
  cas de référence du §4. Si tu penses qu'un changement de résultat est justifié, **demande avant**.
- Aucune commande git qui modifie quoi que ce soit (pas de commit, add, mv, rm, reset…). Lecture seule
  (`status`, `diff`, `log`) autorisée. Déplace les fichiers avec `mv`/`Write`/`Edit`.
- Fichiers source de moins d'environ 500 lignes. Pas de `.collect()` d'une grande structure quand un flux
  suffit. Réutilise l'existant avant d'écrire du code ; enrichis une brique partagée plutôt que d'écrire du
  code ad hoc, **mais** toute modification d'un paquet partagé (obiformats, obiiter, obiseq, obitax) doit
  être signalée clairement et validée par les tests de ces paquets.
- **Pas de spéculation présentée comme un fait** : ce qui n'est pas mesuré est une hypothèse, dis-le.
  Une régression de justesse est pire qu'un gain de mémoire : avant d'échanger l'une contre l'autre, demande.

## 3. Ce qui est mesuré aujourd'hui (macOS, M3 Max, 16 cœurs)

Jeu : 7 822 génomes mitochondriaux de vertébrés (131 Mb de bases), `-r 7953` (Cyprinidae : 325 cibles, 7 497
non-cibles). Commande :
`obiprimers -r 7953 -e 3 --three-prime 2 -l 30 -L 150 -c <fichier annoté> -o out.csv` → 3 247 paires
(3 248 lignes avec l'en-tête).

| Mesure | Valeur |
|---|---|
| pic RSS, taxonomie embarquée, fichier | **1 217 à 1 372 Mo** (bruit d'environ ±100 Mo d'une exécution à l'autre), 4,7 à 5,3 s |
| pic RSS, taxonomie embarquée, entrée standard | médiane **1 334 à 1 346 Mo** sur 8 exécutions entrelacées |
| tas des 7 822 séquences chargées | **193,7 Mo** (dont `taxonomic_path` : 19,2 Mo, soit 2,46 Ko par séquence) |
| même exécution **sans** l'élagage de phase 2 | 5 604 Mo (c'est ce qu'on a corrigé, voir §25 des notes) |
| avec `-t` (taxonomie NCBI complète) | +1,07 Go pour le seul chargeur (mesuré avec `obitaxonomy`), pic ~2,8 Go (mesuré avant les nouvelles définitions) |
| `GOGC=25` (ancienne mesure, `-t` NCBI) | pic 2,82 Go → 1,98 Go, mais temps CPU utilisateur ×2 |
| ecoPrimers (C) sur les mêmes jeux | 1,86 Go (Cyprinidae), 2,11 Go (Teleostei) |

Tailles d'enregistrements : `ampRecord` 24 octets, `Match` 12, `seqMatch` 12, `hit` (intermédiaire de
`matchCorpus`) 16. Environ 6,85 millions d'occurrences approchées quand toutes les séquences sont chargées
(cas `-i 1`).

**Le constat qui motive la mission** : les séquences gardées pèsent ~194 Mo, le pic du processus est ~1,3 Go.
Les 1,1 Go restants ne sont **pas expliqués** (hypothèses ci-dessous, aucune n'est vérifiée).

## 4. Garde-fous à mettre en place AVANT de modifier quoi que ce soit

1. **Binaire de référence** : `cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4 &&
   GOEXPERIMENT=simd /usr/local/go/bin/go build -o /tmp/obiprimers.baseline ./cmd/obitools/obiprimers`
2. **Données** (les fichiers de `/tmp` peuvent avoir disparu, regénère-les) : dossier du tutoriel
   `D="$HOME/Documents/Enseignement/DNA metabarcoding spring school/ecoPrimers-Fish"` : `mito.vert.fasta.gz`,
   `ncbitaxo_20250211.tgz`, et pour le test d'échelle `mito.all.fasta.gz` / `mito.one.fasta.gz`. Fichier annoté
   (taxonomie embarquée, évite le chargeur NCBI) :
   `/Users/coissac/travail/__MOI__/GO/obitools4/build/obiannotate -t "$D/ncbitaxo_20250211.tgz"
   --taxonomic-path -Z -o /tmp/mito.vert.path.fasta.gz "$D/mito.vert.fasta.gz"`
   (les binaires déjà compilés sont dans `/Users/coissac/travail/__MOI__/GO/obitools4/build/`, sans `Sync`).
3. **CSV de référence**, à produire avec le binaire de référence et à conserver : Cyprinidae `-r 7953`,
   Teleostei `-r 32443`, un cas avec contre-exemples `-r 7953 -i 1`, un cas `-U`, un cas `-R NC_083213`, et la
   lecture sur entrée standard (`gzcat fichier | obiprimers ...`). Après chaque modification : `cmp` octet pour
   octet.
4. **Tests** : `go test ./pkg/obiprimer/ ./pkg/obitools/obiprimers/ ./pkg/obithermo/ ./pkg/obikmer/ -race
   -count=1`, et les tests opt-in sur les données réelles (durent ~75 s hors `-race`) :
   `OBIPRIMERS_TUTORIAL="$D" go test ./pkg/obiprimer/ -count=1`.
5. **Méthode de mesure** (le piège principal) : le pic RSS bouge de ±100 Mo entre deux exécutions identiques
   (rythme du ramasse-miettes). Mesure **au moins 5 exécutions entrelacées avant/après et compare les
   médianes** (`/usr/bin/time -l` : octets sur macOS). Pour attribuer la mémoire, ne te fie pas au RSS : utilise
   des profils de tas (`--pprof` de la CLI sert `/debug/pprof/heap` sur `localhost:6060`, `net/http/pprof` est
   importé ; ou `runtime/pprof.WriteHeapProfile` aux frontières d'étapes ; `go tool pprof -sample_index=inuse_space`
   et `alloc_space` ; `GODEBUG=gctrace=1`). **Erreur déjà commise (voir la section 16 bis des notes)** : en
   mesurant le tas libéré par la suppression d'un attribut, la tranche de séquences n'était plus référencée
   après la boucle et le GC a libéré *tout* ; utilise `runtime.KeepAlive` sur ce que tu veux garder vivant.

## 5. Pistes à évaluer (toutes des HYPOTHÈSES à confirmer par un profil, classées sans ordre)

- **Retard du ramasse-miettes** (`GOGC=100` laisse le tas monter jusqu'à ~2× le vivant). Pistes : limite
  mémoire douce (`debug.SetMemoryLimit`) ou `debug.SetGCPercent` ajustés par la commande après la lecture ;
  mesurer le coût CPU (l'ancienne mesure avec `GOGC=25` doublait le temps utilisateur).
- **Lecture** (code partagé `obiformats`/`obiiter`) : 131 Mb sont analysés, avec une table d'attributs et des
  chaînes JSON par séquence. L'option `--batch-mem` (lots limités en mémoire) existe déjà : l'évaluer.
- **Séquences gardées** : `BioSequence` garde 1 octet par base et une table d'attributs. Ne garder que
  l'identifiant, la séquence et le `taxid` (retirer les autres attributs, comme déjà fait pour
  `taxonomic_path`, dans `dropTaxonomicPath` de `pkg/obitools/obiprimers/obiprimers.go`) ; ou stocker la
  séquence à 2 bits par base (~33 Mo pour 131 Mb) avec une liste d'exceptions pour les bases ambiguës. Ce
  second point touche `forEachOccurrence`, `ampHash`, `windowAsIs`, `windowRevComp` et `MatchWords`, qui lisent
  les octets : à ne tenter que si le profil le justifie.
- **Occurrences stockées deux fois** : `[]WordMatches` (`matchCorpus`, intermédiaire `hit` de 16 octets par
  occurrence puis recopie) et `primerSet.perSeq` (`selectPrimers`). Vérifier à quel moment chacune est
  libérée, construire directement les structures finales, présizer les tranches (les `append` en doublement
  créent un pic transitoire).
- **Phase 2b** (`collectAmplicons` sur les seules paires survivantes) : `ampRecord` de 24 octets (16 possibles
  en empaquetant le rôle), enregistrements par worker ; **traitement des paires survivantes par lots bornés**
  si le nombre d'enregistrements estimé dépasse un budget (idée notée en 25.2). Cas dur : `-T 0` (aucune paire
  élaguée), où les tests atteignent ~5,5 Go de tas.
- **`SeqKmerCounter`** (`pkg/obikmer/seqquorum*.go`) : environ 16 octets par occurrence de k-mer pour les
  partitions en cours de traitement ; le nombre de partitions est **fixé à 256**, alors que le total de bases
  est connu une fois le corpus chargé : le choisir automatiquement.
- **Lecture en deux passes** pour la taxonomie embarquée (`readSequencesAndTaxonomy`) : la première passe ne
  garde que les chemins, mais analyse tout.
- **Chargeur de taxonomie NCBI** : 1,07 Go à lui seul ; c'est du code partagé (`obitax`), hors périmètre sauf
  si tu y trouves une économie sûre ; sinon, documente-le dans l'aide (la taxonomie embarquée évite ce coût).

## 6. Test d'échelle (jamais fait)

Tout est validé sur des génomes mitochondriaux (5 à 131 Mb). Mesure la croissance sur un fichier plus gros
(`mito.all.fasta.gz`, 172 Mo compressés, avec `--restrict-to-taxon` sur un grand clade, p. ex. les Metazoa) :
temps et pic RSS en fonction de la taille, pour savoir ce qui grossit vite (corpus, occurrences, amplicons,
compteur de k-mers).

## 7. Livrables

1. Un **rapport de profil** : où va la mémoire à chaque étape (lecture, sélection, mots stricts, appariement,
   paires, phase 2a/2b), avec les profils de tas et les chiffres, avant toute modification.
2. Les optimisations **classées par gain mesuré et par risque**, puis implémentées une par une, chacune avec
   mesure avant/après (médianes d'au moins 5 exécutions entrelacées), `cmp` des CSV de référence, tests verts
   sous `-race`.
3. Une **nouvelle section** à la fin de `OBIPRIMERS_NOTES.md` (ne réécris pas l'historique) : mesures,
   décisions, pistes écartées et pourquoi, ce qui reste. Pas de commit.
4. Si une optimisation exige de changer un résultat, une décision ou un paquet partagé : **arrête-toi et
   demande**.

## 8. Pièges connus de l'environnement

- zsh ne découpe pas une variable en arguments : utilise des tableaux (`A=(-r 7953 ...)` puis `"${A[@]}"`)
  ou des scripts `bash`.
- macOS : `sed -i ''` (sans suffixe) ; `/usr/bin/time -l` donne le pic en **octets** (`maximum resident set size`).
- `OBIPRIMERS_WORKERS` n'est lue que par les tests opt-in ; la CLI utilise `--max-cpu`.
- Les avertissements C de `obiformats/fastseq_read.h` sont connus et sans gravité.
- `go test -race` sur les tests opt-in est lent (~90 s) ; réserve-le aux tests unitaires.
- Le test `obiutils/TestSetString` était instable : il est corrigé (tri des membres dans `Set.String`).
