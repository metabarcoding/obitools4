# Tests pour obisuperkmer

## Description

Ce répertoire contient les tests automatisés pour la commande `obisuperkmer`.

## Fichiers

- `test.sh` : Script de test principal (exécutable)
- `test_sequences.fasta` : Jeu de données de test minimal (3 séquences courtes)
- `README.md` : Ce fichier

## Jeu de données de test

Le fichier `test_sequences.fasta` contient 3 séquences de 32 nucléotides chacune :

1. **seq1** : Répétition du motif ACGT (séquence régulière)
2. **seq2** : Alternance de blocs homopolymères (AAAA, CCCC, GGGG, TTTT)
3. **seq3** : Répétition du motif ATCG (différent de seq1)

Ces séquences sont volontairement courtes pour :
- Minimiser la taille du dépôt Git
- Accélérer l'exécution des tests en CI/CD
- Tester différents cas d'extraction de super k-mers

## Tests effectués

Le script `test.sh` effectue 12 tests :

### Test 1 : Affichage de l'aide
Vérifie que `obisuperkmer -h` s'exécute correctement.

### Test 2 : Extraction basique avec paramètres par défaut
Exécute `obisuperkmer` avec k=21, m=11 (valeurs par défaut).

### Test 3 : Vérification du fichier de sortie non vide
S'assure que la commande produit une sortie.

### Test 4 : Comptage des super k-mers extraits
Vérifie qu'au moins un super k-mer a été extrait.

### Test 5 : Présence des métadonnées requises
Vérifie que chaque super k-mer contient :
- `minimizer_value`
- `minimizer_seq`
- `parent_id`

### Test 6 : Extraction avec paramètres personnalisés
Teste avec k=15 et m=7.

### Test 7 : Vérification des paramètres dans les métadonnées
S'assure que les valeurs k=15 et m=7 sont présentes dans la sortie.

### Test 8 : Format de sortie FASTA explicite
Teste l'option `--fasta-output`.

### Test 9 : Vérification des IDs des super k-mers
S'assure que tous les IDs contiennent "superkmer".

### Test 10 : Préservation des IDs parents
Vérifie que seq1, seq2 et seq3 apparaissent dans la sortie.

### Test 11 : Option -o pour fichier de sortie
Teste la redirection vers un fichier avec `-o`.

### Test 12 : Vérification de la création du fichier avec -o
S'assure que le fichier de sortie a été créé.

### Test 13 : Cohérence des longueurs
Vérifie que la somme des longueurs des super k-mers est inférieure ou égale à la longueur totale des séquences d'entrée.

## Exécution des tests

### Localement

```bash
cd /chemin/vers/obitools4/obitests/obitools/obisuperkmer
./test.sh
```

### En CI/CD

Les tests sont automatiquement exécutés lors de chaque commit via le système CI/CD configuré pour le projet.

### Prérequis

- La commande `obisuperkmer` doit être compilée et disponible dans `../../build/`
- Les dépendances système : bash, grep, etc.

## Structure du script de test

Le script suit le pattern standard utilisé par tous les tests OBITools :

1. **En-tête** : Définition du nom du test et de la commande
2. **Variables** : Configuration des chemins et compteurs
3. **Fonction cleanup()** : Affiche les résultats et nettoie le répertoire temporaire
4. **Fonction log()** : Affiche les messages horodatés
5. **Tests** : Série de tests avec incrémentation des compteurs
6. **Appel cleanup()** : Nettoyage et sortie avec code de retour approprié

## Format de sortie

Chaque test affiche :
```
[obisuperkmer @ date] message
```

En fin d'exécution :
```
========================================
## Results of the obisuperkmer tests:

- 12 tests run
- 12 successfully completed
- 0 failed tests

Cleaning up the temporary directory...

========================================
```

## Codes de retour

- **0** : Tous les tests ont réussi
- **1** : Au moins un test a échoué

## Ajout de nouveaux tests

Pour ajouter un nouveau test, suivre le pattern :

```bash
((ntest++))
if commande_test arguments
then
    log "Description: OK" 
    ((success++))
else
    log "Description: failed"
    ((failed++))
fi
```

## Notes

- Les fichiers temporaires sont créés dans `$TMPDIR` (créé par mktemp)
- Les fichiers de données sont dans `$TEST_DIR`
- La commande testée doit être dans `$OBITOOLS_DIR` (../../build/)
- Le répertoire temporaire est automatiquement nettoyé à la fin
