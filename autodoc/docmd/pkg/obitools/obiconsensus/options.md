# `obiconsensus` Package Functional Overview

The `obiconsensus` package provides command-line options and configuration helpers for sequence clustering, consensus building, and denoising within the OBITools4 framework.

## Core Features

- **Sequence Clustering Mode**: Activated via `--cluster` (`-C`) flag; enables graph-based clustering of related sequences.
- **Denoising with Distance Threshold**: Controlled by `--distance` (`-d`, default: 1), sets the maximum Hamming distance between sequences in a cluster.
- **K-mer Size Control**: `--kmer-size` (`SIZE`, default: -1 = auto-selected) tunes the k-mer size used during consensus construction.
- **Sample Attribute Handling**: `--sample` (`-s`, default: `"sample"`) specifies the metadata field used to group sequences by sample origin.
- **Singleton Filtering**: `--no-singleton` discards unique (non-repeated) sequences if enabled.
- **Low-Coverage Filtering**: Sequences with sample coverage below `--low-coverage` (default: 0.0) are excluded.
- **Dereplication Output**: `--unique` (`-U`) enables output deduplication (equivalent to `obiuniq`).
- **Graph & Ratio Export**: Optional debug outputs:
  - `--save-graph DIR`: Saves DAG structures in GraphML format.
  - `--save-ratio FILE`: Exports edge abundance ratios as CSV.

## Integration

- Integrates with `obiconvert` via input/output option sets (`InputOptionSet`, `OutputOptionSet`) for format handling.
- Uses the `go-getoptions` library to define and parse CLI arguments.

## Getter Functions

All configuration values are exposed via typed accessor functions (e.g., `CLIDistStepMax()`, `CLIKmerSize()`), enabling clean separation of option parsing and logic execution.

