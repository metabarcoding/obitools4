# Semantic Description of `obijoin` Package

The `obijoin` package enables efficient, declarative sequence joins in biological data pipelines. Built on OBITools4’s streaming architecture, it supports left-outer joins between sequence datasets using user-defined attribute keys — ideal for merging paired-end reads, annotating amplicons with metadata, or enriching references.

## Core Components & Functionalities

### `IndexedSequenceSlice`
A composite structure combining a biological sequence slice (`BioSequenceSlice`) with precomputed indices. Each index maps attribute values (e.g., `"sample=S1"`, `"barcode=ATGC"`) to sets of matching sequence indices. Enables sublinear-time filtering via key-based intersection.

### `Get(keys...)`
Performs multi-key *intersection* queries across indexes: returns sequences satisfying **all** provided attribute constraints (e.g., `Get("sample=S1", "barcode=ATGC")`). Keys must match *exactly*; supports arbitrary string attributes via `GetStringAttribute()`.

### `BuildIndexedSequenceSlice()`
Constructs the index structure in **O(*n*)** time by scanning sequences once and grouping them per attribute. Accepts a `BioSequenceSlice` and returns an `IndexedSequenceSlice`. Handles any annotation attribute supported by the sequence system.

### `MakeJoinWorker()`
Returns a functional `SeqWorker` implementing join logic:
- For each input sequence, extracts join keys (e.g., `sample`, `barcode`) from annotations.
- Uses the index to find matching partner sequences (`join_with`).
- Outputs one sequence per match, copying original data and enriching it with partner annotations.
- Optionally updates ID/sequence/quality fields *only if* corresponding flags (`--update-id`, etc.) are enabled.

### `CLIJoinSequences()`
Top-level CLI entry point:
- Reads primary input (stdin or file).
- Loads secondary dataset (`--join-with`), builds index via `BuildIndexedSequenceSlice()`.
- Applies join using worker from `MakeJoinWorker()` with flags (`--by`, `-i/-s/-q`).
- Integrates seamlessly into OBITools4’s streaming iterator model.

## Join Semantics

| Feature | Behavior |
|--------|----------|
| **Join type** | Left outer join (primary dataset fully preserved) |
| **Key matching** | Exact string equality; no regex/fuzzy logic implied |
| **Updates** | Controlled by flags: `-i/--update-id`, `-s/--update-sequence`, `-q/--update-quality` |
| **Metadata handling** | Partner annotations are appended unless fields are overwritten |

## CLI Options

- `-j/--join-with` *(required)*: Path to secondary sequence file (FASTA/FASTQ/TAB).
- `-b/--by`: Join key mapping, e.g. `"id=id"` or `"sample=well"`. Defaults to `["id"]`.
- `-i/--update-id`: Replace sequence identifiers with partner values.
- `-s/--update-sequence`: Overwrite nucleotide/amino acid sequences from partners.
- `-q/--update-quality`: Replace quality scores (FASTQ only).

## Usage Example

```bash
obijoin -i input.fastq \
        --join-with annotations.tsv \
        --by "id=name" \
        -i -s
```
→ Joins `input.fastq` with TSV annotations, matching on `id == name`; updates IDs and sequences.

## Design Principles

- **Efficiency**: Indexing avoids repeated full scans; uses optimized `obiutils.Set[int]` for fast intersection.
- **Extensibility**: Works with any annotation attribute supported by `BioSequence`.
- **Modularity**: CLI logic is configuration-only — no I/O or core algorithms embedded.
- **Composability**: Extends `obiconvert.OptionSet()`; inherits standard format options (`-f`, `-o`) and follows OBITools4 CLI conventions.
