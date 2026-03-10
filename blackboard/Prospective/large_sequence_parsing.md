# Optimisation du parsing des grandes séquences

## Contexte

OBITools4 doit pouvoir traiter des séquences de taille chromosomique (plusieurs Gbp), notamment
issues de fichiers GenBank/EMBL (assemblages de génomes) ou de fichiers FASTA convertis depuis
ces formats.

## Architecture actuelle

### Pipeline de lecture (`pkg/obiformats/`)

```
ReadFileChunk (goroutine)
    → ChannelFileChunk
    → N × _ParseGenbankFile / _ParseFastaFile (goroutines)
    → IBioSequence
```

`ReadFileChunk` (`file_chunk_read.go`) lit le fichier par morceaux via une chaîne de
`PieceOfChunk` (rope). Chaque nœud fait `fileChunkSize` bytes :

- GenBank/EMBL : 128 MB (`1024*1024*128`)
- FASTA/FASTQ  : 1 MB (`1024*1024`)

La chaîne est accumulée jusqu'à trouver la fin du dernier enregistrement complet (splitter),
puis `Pack()` est appelé pour fusionner tous les nœuds en un seul buffer contigu. Ce buffer
est transmis au parseur via `FileChunk.Raw *bytes.Buffer`.

### Parseur GenBank (`genbank_read.go`)

`GenbankChunkParser` reçoit un `io.Reader` sur le buffer packé, lit ligne par ligne via
`bufio.NewReader` (buffer 4096 bytes), et pour chaque ligne de la section `ORIGIN` :

```go
line = string(bline)                        // allocation par ligne
cleanline := strings.TrimSpace(line)        // allocation
parts := strings.SplitN(cleanline, " ", 7) // allocation []string + substrings
for i := 1; i < lparts; i++ {
    seqBytes.WriteString(parts[i])
}
```

Point positif : `seqBytes` est pré-alloué grâce à `lseq` extrait de la ligne `LOCUS`.

### Parseur FASTA (`fastaseq_read.go`)

`FastaChunkParser` lit **octet par octet** via `scanner.ReadByte()`. Pour 3 Gbp :
3 milliards d'appels. `seqBytes` est un `bytes.Buffer{}` sans pré-allocation.

## Problème principal

Pour une séquence de plusieurs Gbp, `Pack()` fusionne une chaîne de ~N nœuds de 128 MB en
un seul buffer contigu. C'est une allocation de N × 128 MB suivie d'une copie de toutes les
données. Bien que l'implémentation de `Pack()` soit efficace (libère les nœuds au fur et à
mesure via `slices.Grow`), la copie est inévitable avec l'architecture actuelle.

De plus, le parseur GenBank produit des dizaines de millions d'allocations temporaires pour
parser la section `ORIGIN` (une par ligne).

## Invariant clé découvert

**Si la rope a plus d'un nœud, le premier nœud seul ne se termine pas sur une frontière
d'enregistrement** (pas de `//\n` en fin de `piece1`).

Preuve par construction dans `ReadFileChunk` :
- `splitter` est appelé dès le premier nœud (ligne 157)
- Si `end >= 0` → frontière trouvée dans 128 MB → boucle interne sautée → rope à 1 nœud
- Si `end < 0` → boucle interne ajoute des nœuds → rope à ≥ 2 nœuds

Corollaire : si rope à 1 nœud, `Pack()` ne fait rien (aucun nœud suivant).

**Attention** : rope à ≥ 2 nœuds ne signifie pas qu'il n'y a qu'une seule séquence dans
la rope. La rope packée peut contenir plusieurs enregistrements complets. Exemple : records
de 80 MB → `nextpieces` (48 MB de reste) + nouveau nœud (128 MB) = rope à 2 nœuds
contenant 2 records complets + début d'un troisième.

L'invariant dit seulement que `piece1` seul est incomplet — pas que la rope entière
ne contient qu'un seul record.

**Invariant : le dernier FileChunk envoyé finit sur une frontière d'enregistrement.**

Deux chemins dans `ReadFileChunk` :

1. **Chemin normal** (`end >= 0` via `splitter`) : le buffer est explicitement tronqué à
   `end` (ligne 200 : `pieces.data = pieces.data[:end]`). Frontière garantie par construction
   pour tous les formats. ✓

2. **Chemin EOF** (`end < 0`, `end = pieces.Len()`) : tout le reste du fichier est envoyé.
   - **GenBank/EMBL** : présuppose fichier bien formé (se termine par `//\n`). Le parseur
     lève un `log.Fatalf` sur tout état inattendu — filet de sécurité suffisant. ✓
   - **FASTQ** : présupposé, vérifié par le parseur. ✓
   - **FASTA** : garanti par le format lui-même (fin d'enregistrement = EOF ou `>`). ✓

**Hypothèse de travail adoptée** : les fichiers d'entrée sont bien formés. Dans le pire cas,
le parseur lèvera une erreur explicite. Il n'y a pas de risque de corruption silencieuse.

## Piste d'optimisation : se dispenser de Pack()

### Idée centrale

Au lieu de fusionner la rope avant de la passer au parseur, **parser directement la rope
nœud par nœud**, et **écrire la séquence compactée in-place dans le premier nœud**.

Pourquoi c'est sûr :
- Le header (LOCUS, DEFINITION, SOURCE, FEATURES) est **petit** et traité en premier
- La séquence (ORIGIN) est **à la fin** du record
- Au moment d'écrire la séquence depuis l'offset 0 de `piece1`, le pointeur de lecture
  est profond dans la rope (offset >> 0) → jamais de collision
- La séquence compactée est toujours plus courte que les données brutes

### Pré-allocation

Pour GenBank/EMBL : `lseq` est connu dès la ligne `LOCUS`/`ID` (première ligne, dans
`piece1`). On peut faire `slices.Grow(piece1.data, lseq)` dès ce moment.

Pour FASTA : pas de taille garantie dans le header, mais `rope.Len()` donne un majorant.
On peut utiliser `rope.Len() / 2` comme estimation initiale.

### Gestion des jonctions entre nœuds

Une ligne peut chevaucher deux nœuds (rare avec 128 MB, mais possible). Solution : carry
buffer de ~128 bytes pour les quelques bytes en fin de nœud.

### Cas FASTA/FASTQ multi-séquences

Un FileChunk peut contenir N séquences (notamment FASTA/FASTQ courts). Dans ce cas
l'écriture in-place dans `piece1` n'est pas applicable directement — on écrase des données
nécessaires aux séquences suivantes.

Stratégie par cas :
- **Rope à 1 nœud** (record ≤ 128 MB) : `Pack()` est trivial (no-op), parseur actuel OK
- **Rope à ≥ 2 nœuds** : par l'invariant, `piece1` ne contient pas de record complet →
  une seule grande séquence → in-place applicable

### Format d'une ligne séquence GenBank (Après ORIGIN)

```
/^ *[0-9]+( [nuc]{10}){0,5} [nuc]{1,10}/
```

### Format d'une ligne séquence GenBank (Après SQ)

La ligne SQ contient aussi la taille de la séquence

```
/^ *( [nuc]{10}){0,5} [nuc]{1,10} *[0-9]+/
```

Compactage in-place sur `bline` ([]byte brut, sans conversion `string`) :

```go
w := 0
i := 0
for i < len(bline) && bline[i] == ' '  { i++ }   // skip indentation
for i < len(bline) && bline[i] <= '9'  { i++ }   // skip position number
for ; i < len(bline); i++ {
    if bline[i] != ' ' {
        bline[w] = bline[i]
        w++
    }
}
// écrire bline[:w] directement dans piece1.data[seqOffset:]
```

## Changements nécessaires

1. **`FileChunk`** : exposer la rope `*PieceOfChunk` non-packée en plus (ou à la place)
   de `Raw *bytes.Buffer`
2. **`GenbankChunkParser` / `EmblChunkParser`** : accepter `*PieceOfChunk`, parser la
   rope séquentiellement avec carry buffer pour les jonctions
3. **`FastaChunkParser`** : idem, avec in-place conditionnel selon taille de la rope
4. **`ReadFileChunk`** : ne pas appeler `Pack()` avant envoi sur le channel (ou version
   alternative `ReadFileChunkRope`)

## Fichiers concernés

- `pkg/obiformats/file_chunk_read.go` — structure rope, `ReadFileChunk`
- `pkg/obiformats/genbank_read.go` — `GenbankChunkParser`, `_ParseGenbankFile`
- `pkg/obiformats/embl_read.go` — `EmblChunkParser`, `ReadEMBL`
- `pkg/obiformats/fastaseq_read.go` — `FastaChunkParser`, `_ParseFastaFile`
- `pkg/obiformats/fastqseq_read.go` — parseur FASTQ (même structure)
