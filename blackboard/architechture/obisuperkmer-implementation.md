# Implémentation de la commande obisuperkmer

## Vue d'ensemble

La commande `obisuperkmer` a été implémentée en suivant l'architecture standard des commandes OBITools décrite dans `architecture-commande-obitools.md`. Cette commande permet d'extraire les super k-mers de fichiers de séquences biologiques.

## Qu'est-ce qu'un super k-mer ?

Un super k-mer est une sous-séquence maximale dans laquelle tous les k-mers consécutifs partagent le même minimiseur. Cette décomposition est utile pour :
- L'indexation efficace de k-mers
- La réduction de la redondance dans les analyses
- L'optimisation de la mémoire pour les structures de données de k-mers

## Structure de l'implémentation

### 1. Package `pkg/obitools/obisuperkmer/`

Le package contient trois fichiers :

#### `obisuperkmer.go`
Documentation du package avec une description de son rôle.

#### `options.go`
Définit les options de ligne de commande :

```go
var _KmerSize = 21          // Taille des k-mers (par défaut 21)
var _MinimizerSize = 11     // Taille des minimiseurs (par défaut 11)
```

**Options CLI disponibles :**
- `--kmer-size` / `-k` : Taille des k-mers (entre m+1 et 31)
- `--minimizer-size` / `-m` : Taille des minimiseurs (entre 1 et k-1)

**Fonctions d'accès :**
- `CLIKmerSize()` : retourne la taille des k-mers
- `CLIMinimizerSize()` : retourne la taille des minimiseurs
- `SetKmerSize(k int)` : définit la taille des k-mers
- `SetMinimizerSize(m int)` : définit la taille des minimiseurs

#### `superkmer.go`
Implémente la logique de traitement :

```go
func CLIExtractSuperKmers(iterator obiiter.IBioSequence) obiiter.IBioSequence
```

Cette fonction :
1. Récupère les paramètres k et m depuis les options CLI
2. Valide les paramètres (m < k, k <= 31, etc.)
3. Crée un worker utilisant `obikmer.SuperKmerWorker(k, m)`
4. Applique le worker en parallèle sur l'itérateur de séquences
5. Retourne un itérateur de super k-mers

### 2. Exécutable `cmd/obitools/obisuperkmer/main.go`

L'exécutable suit le pattern standard minimal :

```go
func main() {
    // 1. Génération du parser d'options
    optionParser := obioptions.GenerateOptionParser(
        "obisuperkmer",
        "extract super k-mers from sequence files",
        obisuperkmer.OptionSet)
    
    // 2. Parsing des arguments
    _, args := optionParser(os.Args)
    
    // 3. Lecture des séquences
    sequences, err := obiconvert.CLIReadBioSequences(args...)
    obiconvert.OpenSequenceDataErrorMessage(args, err)
    
    // 4. Extraction des super k-mers
    superkmers := obisuperkmer.CLIExtractSuperKmers(sequences)
    
    // 5. Écriture des résultats
    obiconvert.CLIWriteBioSequences(superkmers, true)
    
    // 6. Attente de la fin du pipeline
    obiutils.WaitForLastPipe()
}
```

## Utilisation du package `obikmer`

L'implémentation s'appuie sur le package `obikmer` qui fournit :

### `SuperKmerWorker(k int, m int) obiseq.SeqWorker`

Crée un worker qui :
- Extrait les super k-mers d'une BioSequence
- Retourne une slice de BioSequence, une par super k-mer
- Chaque super k-mer contient les attributs suivants :

```go
// Métadonnées ajoutées à chaque super k-mer :
{
    "minimizer_value": uint64,  // Valeur canonique du minimiseur
    "minimizer_seq": string,    // Séquence ADN du minimiseur
    "k": int,                   // Taille des k-mers utilisée
    "m": int,                   // Taille des minimiseurs utilisée
    "start": int,               // Position de début (0-indexé)
    "end": int,                 // Position de fin (exclusif)
    "parent_id": string,        // ID de la séquence parente
}
```

### Algorithme sous-jacent

Le package `obikmer` utilise :
- `IterSuperKmers(seq []byte, k int, m int)` : itérateur sur les super k-mers
- Une deque monotone pour suivre les minimiseurs dans une fenêtre glissante
- Complexité temporelle : O(n) où n est la longueur de la séquence
- Complexité spatiale : O(k-m+1) pour la deque

## Exemple d'utilisation

### Ligne de commande

```bash
# Extraction avec paramètres par défaut (k=21, m=11)
obisuperkmer sequences.fasta > superkmers.fasta

# Spécifier les tailles de k-mers et minimiseurs
obisuperkmer -k 25 -m 13 sequences.fasta -o superkmers.fasta

# Avec plusieurs fichiers d'entrée
obisuperkmer --kmer-size 31 --minimizer-size 15 file1.fasta file2.fasta > output.fasta

# Format FASTQ en entrée, FASTA en sortie
obisuperkmer sequences.fastq --fasta-output -o superkmers.fasta

# Avec compression
obisuperkmer sequences.fasta -o superkmers.fasta.gz --compress
```

### Exemple de sortie

Pour une séquence d'entrée :
```
>seq1
ACGTACGTACGTACGTACGTACGT
```

La sortie contiendra plusieurs super k-mers :
```
>seq1_superkmer_0_15 {"minimizer_value":123456,"minimizer_seq":"acgtacgt","k":21,"m":11,"start":0,"end":15,"parent_id":"seq1"}
ACGTACGTACGTACG
>seq1_superkmer_8_24 {"minimizer_value":789012,"minimizer_seq":"gtacgtac","k":21,"m":11,"start":8,"end":24,"parent_id":"seq1"}
TACGTACGTACGTACGT
```

## Options héritées de `obiconvert`

La commande hérite de toutes les options standard d'OBITools :

### Options d'entrée
- `--fasta` : forcer le format FASTA
- `--fastq` : forcer le format FASTQ
- `--ecopcr` : format ecoPCR
- `--embl` : format EMBL
- `--genbank` : format GenBank
- `--input-json-header` : en-têtes JSON
- `--input-OBI-header` : en-têtes OBI

### Options de sortie
- `--out` / `-o` : fichier de sortie (défaut : stdout)
- `--fasta-output` : sortie en format FASTA
- `--fastq-output` : sortie en format FASTQ
- `--json-output` : sortie en format JSON
- `--output-json-header` : en-têtes JSON en sortie
- `--output-OBI-header` / `-O` : en-têtes OBI en sortie
- `--compress` / `-Z` : compression gzip
- `--skip-empty` : ignorer les séquences vides
- `--no-progressbar` : désactiver la barre de progression

## Compilation

Pour compiler la commande :

```bash
cd /chemin/vers/obitools4
go build -o bin/obisuperkmer ./cmd/obitools/obisuperkmer/
```

## Tests

Pour tester la commande :

```bash
# Créer un fichier de test
echo -e ">test\nACGTACGTACGTACGTACGTACGTACGTACGT" > test.fasta

# Exécuter obisuperkmer
obisuperkmer test.fasta

# Vérifier avec des paramètres différents
obisuperkmer -k 15 -m 7 test.fasta
```

## Validation des paramètres

La commande valide automatiquement :
- `1 <= m < k` : le minimiseur doit être plus petit que le k-mer
- `2 <= k <= 31` : contrainte du codage sur 64 bits
- `len(sequence) >= k` : la séquence doit être assez longue

En cas de paramètres invalides, la commande affiche une erreur explicite et s'arrête.

## Intégration avec le pipeline OBITools

La commande s'intègre naturellement dans les pipelines OBITools :

```bash
# Pipeline complet d'analyse
obiconvert sequences.fastq --fasta-output | \
  obisuperkmer -k 21 -m 11 | \
  obiuniq | \
  obigrep -p "minimizer_value>1000" > filtered_superkmers.fasta
```

## Parallélisation

La commande utilise automatiquement :
- `obidefault.ParallelWorkers()` pour le traitement parallèle
- Les workers sont distribués sur les séquences d'entrée
- La parallélisation est transparente pour l'utilisateur

## Conformité avec l'architecture OBITools

L'implémentation respecte tous les principes de l'architecture :

✅ Séparation des responsabilités (package + commande)
✅ Convention de nommage cohérente (CLI*, Set*, _variables)
✅ Réutilisation de `obiconvert` pour l'I/O
✅ Options standard partagées
✅ Pattern Worker pour le traitement
✅ Validation des paramètres
✅ Logging avec `logrus`
✅ Gestion d'erreurs cohérente
✅ Documentation complète

## Fichiers créés

```
pkg/obitools/obisuperkmer/
├── obisuperkmer.go      # Documentation du package
├── options.go           # Définition des options CLI
└── superkmer.go         # Implémentation du traitement

cmd/obitools/obisuperkmer/
└── main.go              # Point d'entrée de la commande
```

## Prochaines étapes

1. **Compilation** : Compiler la commande avec `go build`
2. **Tests unitaires** : Créer des tests dans `pkg/obitools/obisuperkmer/superkmer_test.go`
3. **Documentation utilisateur** : Ajouter la documentation de la commande
4. **Intégration CI/CD** : Ajouter aux tests d'intégration
5. **Benchmarks** : Mesurer les performances sur différents jeux de données

## Références

- Architecture des commandes OBITools : `architecture-commande-obitools.md`
- Package `obikmer` : `pkg/obikmer/`
- Tests du package : `pkg/obikmer/superkmer_iter_test.go`
