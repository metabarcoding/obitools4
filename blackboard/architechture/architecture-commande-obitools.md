# Architecture d'une commande OBITools

## Vue d'ensemble

Une commande OBITools suit une architecture modulaire et standardisée qui sépare clairement les responsabilités entre :
- Le package de la commande dans `pkg/obitools/<nom_commande>/`
- L'exécutable dans `cmd/obitools/<nom_commande>/`

Cette architecture favorise la réutilisabilité du code, la testabilité et la cohérence entre les différentes commandes de la suite OBITools.

## Structure du projet

```
obitools4/
├── pkg/obitools/
│   ├── obiconvert/          # Commande de conversion (base pour toutes)
│   │   ├── obiconvert.go    # Fonctions vides (pas d'implémentation)
│   │   ├── options.go       # Définition des options CLI
│   │   ├── sequence_reader.go  # Lecture des séquences
│   │   └── sequence_writer.go  # Écriture des séquences
│   ├── obiuniq/             # Commande de déréplication
│   │   ├── obiuniq.go       # (fichier vide)
│   │   ├── options.go       # Options spécifiques à obiuniq
│   │   └── unique.go        # Implémentation du traitement
│   ├── obipairing/          # Assemblage de lectures paired-end
│   ├── obisummary/          # Résumé de fichiers de séquences
│   └── obimicrosat/         # Détection de microsatellites
└── cmd/obitools/
    ├── obiconvert/
    │   └── main.go          # Point d'entrée de la commande
    ├── obiuniq/
    │   └── main.go
    ├── obipairing/
    │   └── main.go
    ├── obisummary/
    │   └── main.go
    └── obimicrosat/
        └── main.go
```

## Composants de l'architecture

### 1. Package `pkg/obitools/<commande>/`

Chaque commande possède son propre package dans `pkg/obitools/` qui contient l'implémentation complète de la logique métier. Ce package est structuré en plusieurs fichiers :

#### a) `options.go` - Gestion des options CLI

Ce fichier définit :
- Les **variables globales** privées (préfixées par `_`) stockant les valeurs des options
- La fonction **`OptionSet()`** qui configure toutes les options pour la commande
- Les fonctions **`CLI*()`** qui retournent les valeurs des options (getters)
- Les fonctions **`Set*()`** qui permettent de définir les options programmatiquement (setters)

**Exemple (obiuniq/options.go) :**

```go
package obiuniq

import (
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
    "github.com/DavidGamba/go-getoptions"
)

// Variables globales privées pour stocker les options
var _StatsOn = make([]string, 0, 10)
var _Keys = make([]string, 0, 10)
var _InMemory = false
var _chunks = 100

// Configuration des options spécifiques à la commande
func UniqueOptionSet(options *getoptions.GetOpt) {
    options.StringSliceVar(&_StatsOn, "merge", 1, 1,
        options.Alias("m"),
        options.ArgName("KEY"),
        options.Description("Adds a merged attribute..."))
    
    options.BoolVar(&_InMemory, "in-memory", _InMemory,
        options.Description("Use memory instead of disk..."))
    
    options.IntVar(&_chunks, "chunk-count", _chunks,
        options.Description("In how many chunks..."))
}

// OptionSet combine les options de base + les options spécifiques
func OptionSet(options *getoptions.GetOpt) {
    obiconvert.OptionSet(false)(options)  // Options de base
    UniqueOptionSet(options)              // Options spécifiques
}

// Getters pour accéder aux valeurs des options
func CLIStatsOn() []string {
    return _StatsOn
}

func CLIUniqueInMemory() bool {
    return _InMemory
}

// Setters pour définir les options programmatiquement
func SetUniqueInMemory(inMemory bool) {
    _InMemory = inMemory
}
```

**Convention de nommage :**
- Variables privées : `_NomOption` (underscore préfixe)
- Getters : `CLINomOption()` (préfixe CLI)
- Setters : `SetNomOption()` (préfixe Set)

#### b) Fichier(s) d'implémentation

Un ou plusieurs fichiers contenant la logique métier de la commande :

**Exemple (obiuniq/unique.go) :**

```go
package obiuniq

import (
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obichunk"
)

// Fonction CLI principale qui orchestre le traitement
func CLIUnique(sequences obiiter.IBioSequence) obiiter.IBioSequence {
    // Récupération des options via les getters CLI*()
    options := make([]obichunk.WithOption, 0, 30)
    
    options = append(options,
        obichunk.OptionBatchCount(CLINumberOfChunks()),
    )
    
    if CLIUniqueInMemory() {
        options = append(options, obichunk.OptionSortOnMemory())
    } else {
        options = append(options, obichunk.OptionSortOnDisk())
    }
    
    // Appel de la fonction de traitement réelle
    iUnique, err := obichunk.IUniqueSequence(sequences, options...)
    
    if err != nil {
        log.Fatal(err)
    }
    
    return iUnique
}
```

**Autres exemples d'implémentation :**

- **obimicrosat/microsat.go** : Contient `MakeMicrosatWorker()` et `CLIAnnotateMicrosat()`
- **obisummary/obisummary.go** : Contient `ISummary()` et les structures de données

#### c) Fichiers utilitaires (optionnel)

Certaines commandes ont des fichiers additionnels pour des fonctionnalités spécifiques.

**Exemple (obipairing/options.go) :**

```go
// Fonction spéciale pour créer un itérateur de séquences pairées
func CLIPairedSequence() (obiiter.IBioSequence, error) {
    forward, err := obiconvert.CLIReadBioSequences(_ForwardFile)
    if err != nil {
        return obiiter.NilIBioSequence, err
    }
    
    reverse, err := obiconvert.CLIReadBioSequences(_ReverseFile)
    if err != nil {
        return obiiter.NilIBioSequence, err
    }
    
    paired := forward.PairTo(reverse)
    return paired, nil
}
```

### 2. Package `obiconvert` - La base commune

Le package `obiconvert` est spécial car il fournit les fonctionnalités de base utilisées par toutes les autres commandes :

#### Fonctionnalités fournies :

1. **Lecture de séquences** (`sequence_reader.go`)
   - `CLIReadBioSequences()` : lecture depuis fichiers ou stdin
   - Support de multiples formats (FASTA, FASTQ, EMBL, GenBank, etc.)
   - Gestion des fichiers multiples
   - Barre de progression optionnelle

2. **Écriture de séquences** (`sequence_writer.go`)
   - `CLIWriteBioSequences()` : écriture vers fichiers ou stdout
   - Support de multiples formats
   - Gestion des lectures pairées
   - Compression optionnelle

3. **Options communes** (`options.go`)
   - Options d'entrée (format, skip, etc.)
   - Options de sortie (format, fichier, compression)
   - Options de mode (barre de progression, etc.)

#### Utilisation par les autres commandes :

Toutes les commandes incluent les options de `obiconvert` via :

```go
func OptionSet(options *getoptions.GetOpt) {
    obiconvert.OptionSet(false)(options)  // false = pas de fichiers pairés
    MaCommandeOptionSet(options)          // Options spécifiques
}
```

### 3. Exécutable `cmd/obitools/<commande>/main.go`

Le fichier `main.go` de chaque commande est volontairement **minimaliste** et suit toujours le même pattern :

```go
package main

import (
    "os"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/macommande"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
    // 1. Configuration optionnelle de paramètres par défaut
    obidefault.SetBatchSize(10)
    
    // 2. Génération du parser d'options
    optionParser := obioptions.GenerateOptionParser(
        "macommande",                    // Nom de la commande
        "description de la commande",    // Description
        macommande.OptionSet)            // Fonction de configuration des options
    
    // 3. Parsing des arguments
    _, args := optionParser(os.Args)
    
    // 4. Lecture des séquences d'entrée
    sequences, err := obiconvert.CLIReadBioSequences(args...)
    obiconvert.OpenSequenceDataErrorMessage(args, err)
    
    // 5. Traitement spécifique de la commande
    resultat := macommande.CLITraitement(sequences)
    
    // 6. Écriture des résultats
    obiconvert.CLIWriteBioSequences(resultat, true)
    
    // 7. Attente de la fin du pipeline
    obiutils.WaitForLastPipe()
}
```

## Patterns architecturaux

### Pattern 1 : Pipeline de traitement de séquences

La plupart des commandes suivent ce pattern :

```
Lecture → Traitement → Écriture
```

**Exemples :**
- **obiconvert** : Lecture → Écriture (conversion de format)
- **obiuniq** : Lecture → Déréplication → Écriture
- **obimicrosat** : Lecture → Annotation → Filtrage → Écriture

### Pattern 2 : Traitement avec entrées multiples

Certaines commandes acceptent plusieurs fichiers d'entrée :

**obipairing** :
```
Lecture Forward + Lecture Reverse → Pairing → Assemblage → Écriture
```

### Pattern 3 : Traitement sans écriture de séquences

**obisummary** : produit un résumé JSON/YAML au lieu de séquences

```go
func main() {
    // ... parsing options et lecture ...
    
    summary := obisummary.ISummary(fs, obisummary.CLIMapSummary())
    
    // Formatage et affichage direct
    if obisummary.CLIOutFormat() == "json" {
        output, _ := json.MarshalIndent(summary, "", "  ")
        fmt.Print(string(output))
    } else {
        output, _ := yaml.Marshal(summary)
        fmt.Print(string(output))
    }
}
```

### Pattern 4 : Utilisation de Workers

Les commandes qui transforment des séquences utilisent souvent le pattern Worker :

```go
// Création d'un worker
worker := MakeMicrosatWorker(
    CLIMinUnitLength(),
    CLIMaxUnitLength(),
    // ... autres paramètres
)

// Application du worker sur l'itérateur
newIter = iterator.MakeIWorker(
    worker, 
    false,                              // merge results
    obidefault.ParallelWorkers()        // parallélisation
)
```

## Étapes d'implémentation d'une nouvelle commande

### Étape 1 : Créer le package dans `pkg/obitools/`

```bash
mkdir -p pkg/obitools/macommande
```

### Étape 2 : Créer `options.go`

```go
package macommande

import (
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
    "github.com/DavidGamba/go-getoptions"
)

// Variables privées pour les options
var _MonOption = "valeur_par_defaut"

// Configuration des options spécifiques
func MaCommandeOptionSet(options *getoptions.GetOpt) {
    options.StringVar(&_MonOption, "mon-option", _MonOption,
        options.Alias("o"),
        options.Description("Description de l'option"))
}

// OptionSet combine options de base + spécifiques
func OptionSet(options *getoptions.GetOpt) {
    obiconvert.OptionSet(false)(options)  // false si pas de fichiers pairés
    MaCommandeOptionSet(options)
}

// Getters
func CLIMonOption() string {
    return _MonOption
}

// Setters
func SetMonOption(value string) {
    _MonOption = value
}
```

### Étape 3 : Créer le fichier d'implémentation

Créer `macommande.go` (ou un nom plus descriptif) :

```go
package macommande

import (
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// Fonction de traitement principale
func CLIMaCommande(sequences obiiter.IBioSequence) obiiter.IBioSequence {
    // Récupération des options
    option := CLIMonOption()
    
    // Implémentation du traitement
    // ...
    
    return resultat
}
```

### Étape 4 : Créer l'exécutable dans `cmd/obitools/`

```bash
mkdir -p cmd/obitools/macommande
```

Créer `main.go` :

```go
package main

import (
    "os"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/macommande"
    "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
    // Parser d'options
    optionParser := obioptions.GenerateOptionParser(
        "macommande",
        "Description courte de ma commande",
        macommande.OptionSet)
    
    _, args := optionParser(os.Args)
    
    // Lecture
    sequences, err := obiconvert.CLIReadBioSequences(args...)
    obiconvert.OpenSequenceDataErrorMessage(args, err)
    
    // Traitement
    resultat := macommande.CLIMaCommande(sequences)
    
    // Écriture
    obiconvert.CLIWriteBioSequences(resultat, true)
    
    // Attente
    obiutils.WaitForLastPipe()
}
```

### Étape 5 : Configurations optionnelles

Dans `main.go`, avant le parsing des options, on peut configurer :

```go
// Taille des batchs de séquences
obidefault.SetBatchSize(10)

// Nombre de workers en lecture (strict)
obidefault.SetStrictReadWorker(2)

// Nombre de workers en écriture
obidefault.SetStrictWriteWorker(2)

// Désactiver la lecture des qualités
obidefault.SetReadQualities(false)
```

### Étape 6 : Gestion des erreurs

Utiliser les fonctions utilitaires pour les messages d'erreur cohérents :

```go
// Pour les erreurs d'ouverture de fichiers
obiconvert.OpenSequenceDataErrorMessage(args, err)

// Pour les erreurs générales
if err != nil {
    log.Errorf("Message d'erreur: %v", err)
    os.Exit(1)
}
```

### Étape 7 : Tests et debugging (optionnel)

Des commentaires dans le code montrent comment activer le profiling :

```go
// go tool pprof -http=":8000" ./macommande ./cpu.pprof
// f, err := os.Create("cpu.pprof")
// if err != nil {
//     log.Fatal(err)
// }
// pprof.StartCPUProfile(f)
// defer pprof.StopCPUProfile()

// go tool trace cpu.trace
// ftrace, err := os.Create("cpu.trace")
// if err != nil {
//     log.Fatal(err)
// }
// trace.Start(ftrace)
// defer trace.Stop()
```

## Bonnes pratiques observées

### 1. Séparation des responsabilités

- **`main.go`** : orchestration minimale
- **`options.go`** : définition et gestion des options
- **Fichiers d'implémentation** : logique métier

### 2. Convention de nommage cohérente

- Variables d'options : `_NomOption`
- Getters CLI : `CLINomOption()`
- Setters : `SetNomOption()`
- Fonctions de traitement CLI : `CLITraitement()`

### 3. Réutilisation du code

- Toutes les commandes réutilisent `obiconvert` pour l'I/O
- Les options communes sont partagées
- Les fonctions utilitaires sont centralisées

### 4. Configuration par défaut

Les valeurs par défaut sont :
- Définies lors de l'initialisation des variables
- Modifiables via les options CLI
- Modifiables programmatiquement via les setters

### 5. Gestion des formats

Support automatique de multiples formats :
- FASTA / FASTQ (avec compression gzip)
- EMBL / GenBank
- ecoPCR
- CSV
- JSON (avec différents formats d'en-têtes)

### 6. Parallélisation

Les commandes utilisent les workers parallèles via :
- `obidefault.ParallelWorkers()`
- `obidefault.SetStrictReadWorker(n)`
- `obidefault.SetStrictWriteWorker(n)`

### 7. Logging cohérent

Utilisation de `logrus` pour tous les logs :
```go
log.Printf("Message informatif")
log.Errorf("Message d'erreur: %v", err)
log.Fatal(err)  // Arrêt du programme
```

## Dépendances principales

### Packages internes OBITools

- `pkg/obidefault` : valeurs par défaut et configuration globale
- `pkg/obioptions` : génération du parser d'options
- `pkg/obiiter` : itérateurs de séquences biologiques
- `pkg/obiseq` : structures et fonctions pour séquences biologiques
- `pkg/obiformats` : lecture/écriture de différents formats
- `pkg/obiutils` : fonctions utilitaires diverses
- `pkg/obichunk` : traitement par chunks (pour dereplication, etc.)

### Packages externes

- `github.com/DavidGamba/go-getoptions` : parsing des options CLI
- `github.com/sirupsen/logrus` : logging structuré
- `gopkg.in/yaml.v3` : encodage/décodage YAML
- `github.com/dlclark/regexp2` : expressions régulières avancées

## Cas spéciaux

### Commande avec fichiers pairés (obipairing)

```go
func OptionSet(options *getoptions.GetOpt) {
    obiconvert.OutputOptionSet(options)
    obiconvert.InputOptionSet(options)
    PairingOptionSet(options)  // Options spécifiques au pairing
}

func CLIPairedSequence() (obiiter.IBioSequence, error) {
    forward, err := obiconvert.CLIReadBioSequences(_ForwardFile)
    // ...
    reverse, err := obiconvert.CLIReadBioSequences(_ReverseFile)
    // ...
    paired := forward.PairTo(reverse)
    return paired, nil
}
```

Dans `main.go` :
```go
pairs, err := obipairing.CLIPairedSequence()  // Lecture spéciale
if err != nil {
    log.Errorf("Cannot open file (%v)", err)
    os.Exit(1)
}

paired := obipairing.IAssemblePESequencesBatch(
    pairs,
    obipairing.CLIGapPenality(),
    // ... autres paramètres
)
```

### Commande sans sortie de séquences (obisummary)

Au lieu de `obiconvert.CLIWriteBioSequences()`, affichage direct :

```go
summary := obisummary.ISummary(fs, obisummary.CLIMapSummary())

if obisummary.CLIOutFormat() == "json" {
    output, _ := json.MarshalIndent(summary, "", "  ")
    fmt.Print(string(output))
} else {
    output, _ := yaml.Marshal(summary)
    fmt.Print(string(output))
}
fmt.Printf("\n")
```

### Commande avec Workers personnalisés (obimicrosat)

```go
func CLIAnnotateMicrosat(iterator obiiter.IBioSequence) obiiter.IBioSequence {
    // Création du worker
    worker := MakeMicrosatWorker(
        CLIMinUnitLength(),
        CLIMaxUnitLength(),
        CLIMinUnitCount(),
        CLIMinLength(),
        CLIMinFlankLength(),
        CLIReoriented(),
    )
    
    // Application du worker
    newIter := iterator.MakeIWorker(
        worker, 
        false,                           // pas de merge
        obidefault.ParallelWorkers(),    // parallélisation
    )
    
    return newIter.FilterEmpty()  // Filtrage des résultats vides
}
```

## Diagramme de flux d'exécution

```
┌─────────────────────────────────────────────────────────────┐
│                      cmd/obitools/macommande/main.go        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  1. Génération du parser d'options                          │
│     obioptions.GenerateOptionParser(                        │
│         "macommande",                                       │
│         "description",                                      │
│         macommande.OptionSet)                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  pkg/obitools/macommande/options.go                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ func OptionSet(options *getoptions.GetOpt)          │   │
│  │   obiconvert.OptionSet(false)(options) ───────────┐ │   │
│  │   MaCommandeOptionSet(options)                    │ │   │
│  └───────────────────────────────────────────────────┼─┘   │
└────────────────────────────────────────────────────────┼─────┘
                              │                         │
                              │                         │
                ┌─────────────┘                         │
                │                                       │
                ▼                                       ▼
┌─────────────────────────────────┐  ┌───────────────────────────────┐
│ 2. Parsing des arguments        │  │ pkg/obitools/obiconvert/      │
│    _, args := optionParser(...) │  │    options.go                 │
└─────────────────────────────────┘  │  - InputOptionSet()           │
                │                     │  - OutputOptionSet()          │
                ▼                     │  - PairedFilesOptionSet()     │
┌─────────────────────────────────┐  └───────────────────────────────┘
│ 3. Lecture des séquences        │
│    CLIReadBioSequences(args)    │
└─────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ pkg/obitools/obiconvert/sequence_reader.go                  │
│  - ExpandListOfFiles()                                      │
│  - ReadSequencesFromFile() / ReadSequencesFromStdin()       │
│  - Support: FASTA, FASTQ, EMBL, GenBank, ecoPCR, CSV        │
└─────────────────────────────────────────────────────────────┘
                │
                ▼ obiiter.IBioSequence
┌─────────────────────────────────────────────────────────────┐
│ 4. Traitement spécifique                                    │
│    macommande.CLITraitement(sequences)                      │
└─────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ pkg/obitools/macommande/<implementation>.go                 │
│  - Récupération des options via CLI*() getters             │
│  - Application de la logique métier                         │
│  - Retour d'un nouvel iterator                              │
└─────────────────────────────────────────────────────────────┘
                │
                ▼ obiiter.IBioSequence
┌─────────────────────────────────────────────────────────────┐
│ 5. Écriture des résultats                                   │
│    CLIWriteBioSequences(resultat, true)                     │
└─────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ pkg/obitools/obiconvert/sequence_writer.go                  │
│  - WriteSequencesToFile() / WriteSequencesToStdout()        │
│  - Support: FASTA, FASTQ, JSON                              │
│  - Gestion des lectures pairées                             │
│  - Compression optionnelle                                  │
└─────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Attente de fin du pipeline                               │
│    obiutils.WaitForLastPipe()                               │
└─────────────────────────────────────────────────────────────┘
```

## Conclusion

L'architecture des commandes OBITools est conçue pour :

1. **Maximiser la réutilisation** : `obiconvert` fournit les fonctionnalités communes
2. **Simplifier l'ajout de nouvelles commandes** : pattern standardisé et minimaliste
3. **Faciliter la maintenance** : séparation claire des responsabilités
4. **Garantir la cohérence** : conventions de nommage et structure uniforme
5. **Optimiser les performances** : parallélisation intégrée et traitement par batch

Cette architecture modulaire permet de créer rapidement de nouvelles commandes tout en maintenant une qualité et une cohérence élevées dans toute la suite OBITools.
