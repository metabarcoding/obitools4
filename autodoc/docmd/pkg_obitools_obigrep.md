Here's a concise, semantically structured Markdown description (≤200 lines) of the **public-facing functionalities** provided by the `obigrep` package, based on your input and focusing only on *public* APIs:

```markdown
# `obigrep`: Command-Line Sequence Filtering for OBITools4

`obigrep` delivers a robust, CLI-driven filtering engine for biological sequences (FASTA/FASTQ), enabling precise selection or exclusion of reads using diverse criteria—length, abundance, taxonomy, patterns (exact/fuzzy), metadata attributes—and paired-end logic.

## Core Filtering Capabilities

### Length & Abundance
- `--min-length`, `--max-length`: Filter by sequence length.
- `--min-count`, `--max-count`: Filter based on read abundance (count attribute).

### Pattern Matching
- Exact regex via `--sequence`/`-s`, `--definition`/`-D`, or `--identifier`/`-I`.
  - Case-insensitive by default.
- Approximate matching via `--pattern`, with options:
  - `--pattern-error`: Max edit distance.
  - `--allows-indels`: Allow insertions/deletions (default: mismatches only).
  - `--only-forward`: Restrict to forward strand.

### Taxonomic Filtering
- `--restrict-to-taxon`/`-r`: Keep only sequences matching given taxon(s).
- `--ignore-taxon`/`-i`: Exclude specific taxa.
- `--valid-taxid`: Enforce presence of valid NCBI taxids in records.
- `--require-rank`: Require specific taxonomic rank (e.g., *species*, *genus*).

### Attribute & Metadata Filtering
- `--has-attribute`/`-A`: Retain sequences with a given attribute key.
- `--attribute=key=pattern`/`-a`: Match regex against a specific attribute value.
- `--id-list FILE`: Select sequences whose identifiers appear in the file.

### Custom Logic
- `--predicate`/`-p`: Evaluate arbitrary boolean expressions (e.g., `"attr['quality'] > 30 && len(sequence) < 500"`).

### Paired-End Handling
- `--paired-mode`: Define how filters apply to read pairs:
  - `"forward"`: Only forward read considered.
  - `"and"`, `"or"`, `"xor"`, etc.: Logical combinations of forward/reverse filters.

### Output Control
- `--save-discarded FILE`: Write rejected sequences to file.
- `--inverse-match`/`-v`: Globally invert selection (i.e., output *only* discarded reads).

## Implementation Notes

- Filters are composed into a single predicate using `CLISequenceSelectionPredicate()`.
- Paired-end logic is layered via `PairedPredicat()` when input files are paired (`CLIHasPairedFile()`).
- Filtering is executed via `iterator.FilterOn(...)` (in-place) or `DivideOn(...)` + async write to discarded file.
- Uses structured logging (`logrus`) and graceful error handling for robust CLI operation.

## Semantic Role

`obigrep` acts as the **semantic filter layer** in OBITools4 workflows—translating user CLI flags into type-safe, composable predicates that operate uniformly over `IBioSequence` iterators. It bridges high-level biological intent (e.g., “keep only *Bacillales* with ≥Q30 and no Ns”) to low-level filtering primitives.
