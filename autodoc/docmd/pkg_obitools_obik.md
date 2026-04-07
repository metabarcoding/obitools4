# `obik`: K-mer Index Management Toolkit for Biological Sequences

`obik` is a CLI tool from the OBITools4 ecosystem designed for building, inspecting, filtering, and manipulating **k-mer indices**—compact data structures encoding k-mer occurrences from biological sequences (e.g., FASTA/FASTQ). It enables scalable, parallelized processing of large-scale sequencing data for applications such as taxonomic profiling, contamination screening, and metagenomic analysis.

All documented features are **public APIs**, accessible via subcommands. Internal implementation details (e.g., low-level k-mer engines) are omitted.

---

## Core Subcommands

### `obik index`
Builds or extends a k-mer set group from raw sequences:
- Configurable `k` (2–31) and optional minimizer size (`m`) for space-efficient hashing.
- Filters by k-mer frequency: `--minocc`, `--maxocc`.
- Entropy-based low-complexity filtering (`--entropy-threshold`, `--entropy-size`).
- Supports metadata tagging at group, set, and per-set levels (`--set-tag`, `--index-id`).
- Optionally saves top *N* frequent k-mers (`--save-freq-kmers`) for downstream analysis.
- Parallel sequence processing with atomic counters and thread-safe batching.

### `obik ls`
Lists metadata of k-mer sets in an index:
- Accepts glob-like `--set PATTERN`s to filter target sets.
- Outputs structured metadata: set index, ID, and k-mer count (`count`).
- Supports multiple formats: CSV (default), JSON, YAML.
- No k-mers themselves are printed—only set-level summaries.

### `obik summary`
Aggregates and reports comprehensive statistics:
- Structural info: k, m, partitions, total sets/unique kmers.
- Per-set stats: ID, count, disk usage (computed recursively).
- Optional pairwise **Jaccard distance matrix** for similarity analysis.
- Multi-format export (JSON/YAML/CSV) with full metadata preservation.

### `obik cp`
Copies selected or all k-mer sets from a source index to a new destination:
- Requires `<source_index>` and `<dest_dir>`.
- Pattern-based selection via `--set PATTERN` (glob-style); fails if no match.
- Prevents overwrites unless `--force`.
- Uses atomic copy operations via `CopySetsByIDTo`, preserving original structure.

### `obik mv`
Safely moves sets between indices:
- Copy-first, then delete strategy ensures atomicity.
- Supports `--set PATTERN` for selective moves; fails if no sets match patterns.
- Removes source sets in reverse order to avoid index renumbering issues.
- Logs progress and final counts for observability.

### `obik rm`
Removes k-mer sets from an index:
- Requires at least one glob-like `--set PATTERN`.
- Validates existence and match success before deletion.
- Deletes sets in reverse order to preserve indices during bulk removals.
- Fails fast on errors, leaving index consistent.

### `obik spectrum`
Exports k-mer frequency spectra per set:
- Computes histogram: how many distinct kmers occur *exactly N times*.
- Outputs sparse CSV (only non-zero frequencies), with per-set columns.
- Enables comparative analysis of redundancy/complexity across samples.

### `obik filter`
Filters k-mers from an index using configurable criteria:
- Currently supports entropy-based filtering (`--entropy-threshold`, `--entropy-size`).
- Runs in parallel across partitions (per-worker filter instantiation for stateful filters).
- Preserves partitioning structure and `spectrum.bin` files.
- Logs per-set statistics (kept %, total processed).

### `obik match`
Annotates query sequences with reference matches:
- Loads a k-mer index and selects target sets via patterns.
- Reads sequences (FASTA/FASTQ), prepares queries in parallel, and merges batches incrementally.
- Matches k-mers against reference sets using `MatchBatch`, attaches match positions as attributes (e.g., `"kmer_matched_ref_genome"`).
- Streams annotated output with paired-end integrity preserved.

### `obik lowmask`
Masks or extracts low-complexity regions in sequences:
- Uses multi-scale entropy analysis (window sizes 1–`level_max`) on canonical k-mers.
- Three modes: **mask** (replace with `.` or custom char), **split**, and **extract low-complexity fragments**.
- Preserves metadata (e.g., entropy values) on output sequences.

### `obik super`
Extracts *super k-mers* from overlapping reads:
- Merges contiguously overlapped kmers sharing a minimizer into longer, non-overlapping super-k-mers.
- Configurable `k` and `m`; parallelized via worker pipeline.
- Optimized for alignment-free analysis, read correction, and compression.

---

## Shared Capabilities

### Set Selection
- Glob-style pattern matching (`--set PATTERN`, repeatable).
- Resolves to exact set IDs using `MatchSetIDs`.

### Output Formatting
- Structured output: CSV, JSON (`--json-output`), YAML (`--yaml-output`) across multiple commands.

### Metadata Handling
- Group-, set-, and per-kmer metadata support (`--set-tag`, `metadata.toml`).
- Preserved during copy/move/filter operations.

### Safety & Observability
- Structured logging (Logrus), progress bars (`progressbar`).
- Context-aware cancellation and timeout support.
- Detailed error wrapping with `%w`.

### Parallelism
- Multi-worker pipelines (e.g., `nworkers` from system defaults).
- Thread-safe accumulation and atomic counters where needed.

---

> **Note**: All commands assume a valid `KmerSetGroup` index structure (`.kdi`, `.toml`). No k-mer sequences themselves are printed—only metadata, counts, or match annotations.
