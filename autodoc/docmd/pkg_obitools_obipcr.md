# `obipcr`: In-Silico PCR Simulation CLI Package

The `obipcr` package provides a robust, configurable command-line interface for simulating *in silico* PCR amplifications on biological sequences. It enables flexible primer design, mismatch-tolerant binding, amplicon filtering by length and completeness, support for circular genomes, and optimized handling of large input datasets.

## Core Features

### Primer Definition & Matching
- **Forward/Reverse Primers**: Required inputs (`--forward`, `--reverse`) supporting degenerate nucleotide patterns (e.g., IUPAC ambiguity codes) via integration with `obitools4/pkg/obiapat`.
- **Mismatch Tolerance**: Configurable per-primer mismatch budget (`--allowed-mismatches`, `-e`) using pattern-based alignment via `MakeApatPattern`.

### Amplicon Filtering & Constraints
- **Length Bounds**: Enforces minimum (`--min-length`, `-l`) and maximum (`--max-length`, `L`) amplicon sizes (excluding primers).
- **Completeness Check**: Option (`--only-complete-flanking`) restricts output to amplicons where both primer-binding sites are fully contained in the input sequence.

### Topology & Extension Handling
- **Circular DNA Support**: Activated via `--circular` (`-c`) to allow primers binding across sequence termini.
- **Flanking Extension**: Optional inclusion of upstream/downstream regions (`--delta`, `-D`) beyond primer sites for realistic amplicon modeling.

### Scalability & Performance
- **Fragmentation Strategy**: Long sequences (> `max-length × 1000`) are split into overlapping segments (~`max-length × 1000 bp`) to accelerate PCR search (`--fragmented`).
- **Parallel Execution**: Leverages `obidefault.ParallelWorkers()` for concurrent processing.
- **Memory Control**: Limits memory usage to ≤50% of available RAM (`LimitMemory(0.5)`).

## Public API

### CLI Option Registration
- `PCROptionSet()`: Registers all PCR-specific flags with the underlying option parser.
- `OptionSet()`: Extends above by integrating standard conversion options (`obiconvert.OptionSet`).

### Safe Value Accessors
- Getter functions (e.g., `CLIForwardPrimer()`, `CLIMinLength()`) provide typed, validated access to parsed options—including compiled nucleotide patterns and error-checked ranges.

### Main Execution Entry Point
- `CLIPCR(seqIter)`: Performs *in silico* PCR over an input sequence iterator, returning amplified fragments as a new batched output iterator. Configured entirely via CLI options.

## Design Principles

- **Fail-Fast Validation**: All required parameters (e.g., primers) are validated at parse time; missing values trigger immediate fatal errors.
- **Pattern-Centric Matching**: Mismatch-tolerant binding is implemented via robust pattern-matching primitives (`obiapat`), not naive string comparison.
- **Modular Architecture**: Clear separation between CLI parsing, algorithm configuration (`PCRSliceWorker`), and execution orchestration ensures maintainability.

This package is ideal for building scalable amplicon-based metagenomics pipelines with high precision and tunable sensitivity.
