# `obidemerge` Package Documentation

The **`obidemerge`** package enables *demerging* of biological sequences—i.e., splitting aggregated or merged sequence records into discrete, count-annotated variants based on metadata statistics. It supports both programmatic and CLI workflows for downstream processing in metabarcoding or amplicon-based pipelines.

## Core Functionalities

### 1. `MakeDemergeWorker(key string) SeqWorker`
- **Purpose**: Constructs a sequence processor that splits sequences by statistical metadata.
- **Behavior**:
  - Scans the input sequence for a statistics map under attribute `key`.
    - *Example*: If `"sample"` → `{ "S1": 5, "S2": 3 }`, two new sequences are generated.
  - For each `(stat_key, count)` pair:
    - Copies the original sequence data,
    - Adds a new attribute: `key = stat_key`,
    - Sets `.Count` to the corresponding integer value.
  - Removes original statistics from the input sequence after splitting.
- **Fallback**: If no stats are found for `key`, returns a single-element slice containing the unchanged sequence.

### 2. `CLIDemergeSequences(iterator, slot string) SeqIterator`
- **Purpose**: CLI wrapper for batch demerging.
- **Behavior**:
  - Applies `MakeDemergeWorker(slot)` to each sequence in the input iterator.
  - Supports parallel processing (implementation-dependent).
- **Integration**:
  - Designed to be used with the `--demerge` CLI flag (see below).

### 3. CLI Integration via OptionSet
- **Flag**: `--demerge` (`-d`)
  - Specifies the metadata slot to demerge (default: `"sample"`).
- **APIs**:
  - `DemergeOptionSet(options *getoptions.Options)`: Registers the `-d/--demerge` flag.
  - `CLIDemergeSlot() string`: Returns the selected slot name (e.g., `"sample"`), used by downstream workers.
- **Inheritance**:
  - Extends `obiconvert.OptionSet`, inheriting standard conversion options (I/O formats, filters, etc.).

## Semantic Workflow

1. **Input**: Sequences with embedded statistical metadata (e.g., sample abundances, OTU counts).
2. **Demerge Operation**: Splits each sequence into multiple copies—each tagged with a unique metadata key and abundance.
3. **Output**: A new set of sequences where each variant is independently annotated, enabling:
   - Accurate abundance-aware filtering,
   - Per-variant downstream analysis (e.g., taxonomic assignment, diversity metrics).

## Key Concept: *Demerging*
- **Definition**: Reversal of prior merging steps (e.g., OTU clustering, read pairing).
- **Purpose**: Restores granularity for statistical or ecological interpretation while preserving original sequence data.

## Use Cases
- Post-clustering demerging of OTU/ASV tables.
- Splitting merged paired-end reads by sample or condition metadata.
- Preparing data for tools expecting discrete, count-labeled sequences.

> **Note**: Only *public* APIs are documented. Internal helpers (e.g., slot validation, worker state) remain unspecified.
