# Définition du super k-mer

## Définition

Un **super k-mer** est une **sous-séquence MAXIMALE** d'une séquence dans laquelle **tous les k-mers consécutifs partagent le même minimiseur**.

### Termes

- **k-mer** : sous-séquence de longueur k
- **minimiseur** : le plus petit m-mer canonique parmi tous les m-mers d'un k-mer
- **k-mers consécutifs** : k-mers aux positions i et i+1 (chevauchement de k-1 nucléotides)
- **MAXIMALE** : ne peut être étendue ni à gauche ni à droite

## RÈGLES ABSOLUES

### RÈGLE 1 : Longueur minimum = k

Un super k-mer contient au minimum k nucléotides.

```
longueur(super-kmer) >= k
```

### RÈGLE 2 : Chevauchement obligatoire = k-1

Deux super-kmers consécutifs se chevauchent d'EXACTEMENT k-1 nucléotides.

```
SK1.End - SK2.Start = k - 1
```

### RÈGLE 3 : Bijection séquence ↔ minimiseur

Une séquence de super k-mer a UN et UN SEUL minimiseur.

```
Même séquence → Même minimiseur (TOUJOURS)
```

**Si vous observez la même séquence avec deux minimiseurs différents, c'est un BUG.**

### RÈGLE 4 : Tous les k-mers partagent le minimiseur

TOUS les k-mers contenus dans un super k-mer ont le même minimiseur.

```
∀ k-mer K dans SK : minimiseur(K) = SK.minimizer
```

### RÈGLE 5 : Maximalité

Un super k-mer ne peut pas être étendu.

- Si on ajoute un nucléotide à gauche : le nouveau k-mer a un minimiseur différent
- Si on ajoute un nucléotide à droite : le nouveau k-mer a un minimiseur différent

## VIOLATIONS INTERDITES

❌ **Super k-mer de longueur < k**
❌ **Chevauchement ≠ k-1 entre consécutifs**
❌ **Même séquence avec minimiseurs différents**
❌ **K-mer dans le super k-mer avec minimiseur différent**
❌ **Super k-mer extensible (non-maximal)**

## CONSÉQUENCES PRATIQUES

### Pour l'extraction

L'algorithme doit :
1. Calculer le minimiseur de chaque k-mer
2. Découper quand le minimiseur change
3. Assigner au super k-mer le minimiseur commun à tous ses k-mers
4. Garantir que chaque super k-mer contient au moins k nucléotides
5. Garantir le chevauchement de k-1 entre consécutifs

### Pour la validation

Si après déduplication (obiuniq) on observe :
```
Séquence: ACGT...
Minimiseurs: {M1, M2}  // plusieurs minimiseurs
```

C'est la PREUVE d'un bug : l'algorithme a produit cette séquence avec des minimiseurs différents, ce qui viole la RÈGLE 3.

## DIAGNOSTIC DU BUG

**Bug observé** : Même séquence avec minimiseurs différents après obiuniq

**Cause possible** : L'algorithme assigne le mauvais minimiseur OU découpe mal les super-kmers

**Ce que le bug NE PEUT PAS être** :
- Un problème d'obiuniq (révèle le bug, ne le crée pas)
- Un problème de chevauchement légitime (k-1 est correct)

**Ce que le bug DOIT être** :
- Minimiseur mal calculé ou mal assigné
- Découpage incorrect (mauvais endPos)
- Copie incorrecte des données
