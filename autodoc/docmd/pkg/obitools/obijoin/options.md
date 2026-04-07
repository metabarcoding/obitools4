# `obijoin` Package: Semantic Description

The `obijoin` package provides command-line interface (CLI) configuration and semantic logic for joining two biological sequence files, typically used in NGS data processing pipelines (e.g., merging FASTQ/FASTA with annotation tables).

## Core Functionality

- **Join Specification**:  
  Accepts a file to join with via the `-j/--join-with` option (required). Specifies how records from two datasets are matched.

- **Join Keys Definition**:  
  Uses `-b/--by` to define matching criteria (e.g., `"id=id"` or `"sample=well"`). Defaults to `["id"]` on both sides if omitted. Supports asymmetric keys via `"left=right"` syntax.

- **Field Update Control**:  
  Three boolean flags determine which fields are overwritten in the primary dataset during join:
  - `-i/--update-id`: Replace sequence identifiers.
  - `-s/--update-sequence`: Overwrite nucleotide/amino acid sequences.
  - `-q/--update-quality`: Replace quality scores (relevant for FASTQ).

- **Integration with Base Converter**:  
  Extends `obiconvert.OptionSet()` — inherits standard conversion options (e.g., input/output formats) and appends join-specific flags.

## Semantic Behavior

- Performs a **left outer join** (primary dataset preserved; matched records from secondary file appended/updated).
- Keys are compared semantically: exact string match by default (no regex or fuzzy matching implied).
- Updates occur **only if flags are enabled**; otherwise, joined metadata is ignored or appended conditionally.

## Usage Example

```bash
obijoin -i input.fastq \
        --join-with annotations.tsv \
        --by "id=name" \
        -i -s
```
→ Joins `input.fastq` with `annotations.tsv`, matching on `id == name`; updates IDs and sequences.

## Design Notes

- Minimalist CLI: Leverages `go-getoptions` for declarative argument parsing.
- No file I/O logic in this module — purely configuration and option extraction (`CLI*` accessor functions).
- Designed for composability within `obitools4`, following modular CLI patterns.
