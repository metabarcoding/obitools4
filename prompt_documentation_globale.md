# Objectif
Documenter intégralement l'application OBITools (version 4, écrit en Go) en trois phases incrémentales, en utilisant les outils MCP fournis par cclsp.

# Contexte
- Le code source est organisé en packages Go, chaque outil étant un exécutable avec un `main.go`.
- Vous avez accès à un serveur MCP (cclsp) qui expose les outils suivants (liste non exhaustive) :
  - `find_definition(file, line, character)` : retourne la définition d’un symbole.
  - `find_references(file, line, character)` : retourne toutes les références d’un symbole.
  - `get_diagnostics(file)` : retourne les erreurs/warnings.
  - `rename_symbol(file, line, character, new_name)` : permet de renommer (optionnel).
- Vous pouvez également exécuter des commandes shell pour lister les fichiers, lire/écrire des fichiers.
- Tous les fichiers de documentation seront stockés dans un répertoire `docs/` à la racine du projet, avec la structure suivante :
  ```
  docs/
    phase1/
      <package>/
        <fichier>.md
    phase2/
      <package>.md
    phase3/
      <outil>.md
  ```

# Instructions générales
- Avant de commencer, vérifiez que le répertoire `docs/phase1`, `docs/phase2`, `docs/phase3` existe (créez-les si nécessaire).
- Pour chaque phase, respectez le format Markdown demandé.
- N’hésitez pas à utiliser les outils MCP pour obtenir des informations précises sur les symboles (signatures, commentaires, types).
- Si un outil MCP nécessite des coordonnées (ligne, caractère), vous pouvez extraire ces informations en lisant le fichier source.
- Soyez méthodique : traitez un fichier/package/outil à la fois et sauvegardez immédiatement le résultat.

---

## Phase 1 : Documentation par fichier Go (sauf `main.go`)

**Objectif** : Pour chaque fichier `.go` qui n’est pas un `main.go` d’outil, générer un document Markdown décrivant son rôle, ses structures, ses fonctions principales et ses dépendances.

**Étapes** :
1. Listez tous les fichiers `.go` du projet, en excluant ceux nommés `main.go` (ils seront traités en phase 3). Utilisez par exemple `find . -name "*.go" -not -name "main.go"`.
2. Pour chaque fichier :
   a. Lisez son contenu (vous pouvez le faire via shell ou en utilisant l’outil `read_file` si disponible).
   b. Utilisez les outils MCP pour extraire les informations suivantes :
      - Pour chaque fonction publique (commençant par une majuscule) : appelez `find_definition` avec la ligne approximative où se trouve la fonction. Récupérez la signature et les commentaires associés.
      - Pour les types (structs, interfaces) : faites de même.
   c. Générez un document Markdown avec les sections :
      ```markdown
      # Fichier : `chemin/vers/fichier.go`
      ## Rôle
      (une phrase ou deux)

      ## Structures
      - `NomStruct` : description, champs principaux

      ## Fonctions principales
      - `NomFonction(paramètres) (retours)` : description

      ## Dépendances notables
      - packages importés (externes ou internes) significatifs
      ```
   d. Sauvegardez dans `docs/phase1/<package>/<fichier>.md` (où `<package>` est le nom du répertoire parent contenant le fichier, et `<fichier>` le nom sans extension).

**Exemple d’utilisation d’outil MCP** :
```
# Pour obtenir la définition de la fonction `Align` dans `align/align.go`
# On suppose qu’elle se trouve approximativement ligne 120, colonne 1.
find_definition(file="align/align.go", line=120, character=1)
```

**Conseil** : Vous pouvez d’abord lister tous les symboles exportés d’un fichier en utilisant `go list` ou en analysant le code. Mais les outils MCP peuvent aussi vous aider.

---

## Phase 2 : Agrégation par package

**Objectif** : Pour chaque package (répertoire contenant au moins un fichier `.go` sauf `main.go`), créer un document Markdown qui synthétise l’ensemble des fichiers du package.

**Étapes** :
1. Identifiez tous les packages (répertoires) qui contiennent des fichiers `.go` traités en phase 1.
2. Pour chaque package :
   a. Lisez tous les fichiers `.md` de `docs/phase1/<package>/`.
   b. Utilisez les outils MCP pour obtenir une vue d’ensemble des symboles exportés du package (par exemple, en interrogeant `find_references` sur un symbole clé ou en parcourant les définitions).
   c. Générez un document Markdown avec :
      ```markdown
      # Package : `<nom>`

      ## Présentation
      Description générale du package, son rôle dans OBITools (traitement de séquences, alignement, etc.)

      ## Structure interne
      - Liste des fichiers principaux et leurs responsabilités (liens vers docs phase1)
      - Flux de données ou interactions entre fichiers

      ## API publique
      - **Types** : liste des types exportés avec brève description
      - **Fonctions** : liste des fonctions exportées avec signature et rôle
      - **Constantes/Variables** si pertinentes

      ## Exemple d’utilisation
      (si possible, un extrait de code illustrant comment ce package est utilisé ailleurs)
      ```
   d. Sauvegardez dans `docs/phase2/<package>.md`.

**Conseil** : Pour l’API publique, vous pouvez invoquer `find_definition` sur chaque symbole exporté (en vous basant sur les fichiers sources) pour obtenir les signatures exactes et les commentaires.

---

## Phase 3 : Documentation des outils (exécutables)

**Objectif** : Pour chaque outil (répertoire contenant un `main.go`), générer une documentation utilisateur complète, en utilisant les informations des packages documentés en phase 2.

**Étapes** :
1. Trouvez tous les `main.go` du projet. Pour chacun, identifiez le nom de l’outil (généralement le nom du répertoire parent).
2. Pour chaque outil :
   a. Lisez le `main.go` pour comprendre la logique globale et les options de ligne de commande (flags, cobra, etc.).
   b. Identifiez tous les packages Go importés par l’outil (en analysant les imports du fichier).
   c. Récupérez les documents `docs/phase2/<package>.md` correspondants.
   d. Utilisez les outils MCP pour explorer les fonctions appelées dans le `main.go` et comprendre leur rôle.
   e. Générez un document Markdown :
      ```markdown
      # Outil : `<nom_outil>`

      ## Résumé
      Une ligne de description.

      ## Description
      Explication détaillée de ce que fait l’outil, dans quel contexte l’utiliser, comment il traite les données (séquences, fichiers, etc.).

      ## Options
      | Option | Type | Défaut | Description |
      |--------|------|--------|-------------|
      | `--flag` | string | "" | description extraite du code |

      ## Exemples
      ```bash
      # Exemple 1 : utilisation simple
      outil -i input.fasta -o output.fasta

      # Exemple 2 : avec options avancées
      outil --flag value ...
      ```

      ## Voir aussi
      - [Package `<package1>`](../phase2/<package1>.md)
      - [Package `<package2>`](../phase2/<package2>.md)
      - Autres outils connexes
      ```
   f. Sauvegardez dans `docs/phase3/<outil>.md`.

**Conseil** : Pour extraire les options, examinez le code qui utilise `flag` ou `cobra`. Vous pouvez aussi utiliser `go doc` sur le package principal, mais les outils MCP vous permettront de suivre les références aux symboles.

---

# Validation et finalisation
- Après avoir généré tous les documents, vérifiez que la documentation phase3 inclut bien des liens fonctionnels vers les documents phase2.
- Si certains packages ne sont pas documentés (parce que leurs fichiers n’ont pas été traités en phase1), repérez les oublis et corrigez.
- Vous pouvez générer un index global dans `docs/README.md` listant tous les packages et outils avec leurs descriptions.

---

# Consignes d’exécution
- Travaillez de manière séquentielle : terminez une phase avant de passer à la suivante.
- Pour chaque fichier/package/outil, utilisez les outils MCP de manière parcimonieuse mais exhaustive.
- Si un outil MCP échoue (par exemple, coordonnées incorrectes), ajustez en relisant le fichier source pour trouver la bonne ligne/colonne.
- Enregistrez les résultats immédiatement après génération pour éviter les pertes.

Maintenant, commencez par la phase 1. Listez tous les fichiers `.go` (hors main.go) et générez la documentation pour le premier fichier.


