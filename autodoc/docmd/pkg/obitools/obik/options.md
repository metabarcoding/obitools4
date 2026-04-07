# ObiTools4 CLI Package: Semantic Description of Features

This Go package (`obik`) defines command-line interface (CLI) options and utilities for the ObiTools4 suite, focused on **k-mer-based analysis of biological sequences**.

## Core Functionalities

### 1. K-mer Indexing (`index` subcommand)
Builds a k-mer index from input sequences, supporting:
- Configurable **k-mer size** (`--kmer-size`, default 31).
- Minimizer-based parallelization via `--minimizer-size`.
- Filtering by occurrence (`--min-occurrence`, `--max-occurrence`).
- Entropy-based low-complexity filtering (`--entropy-filter`, `--entropy-filter-size`).
- Metadata storage in TOML/YAML/JSON (`--metadata-format`).
- Optional export of top *N* frequent k-mers to CSV.

### 2. Low-complexity Masking (`lowmask` subcommand)
Processes sequences to handle low-complexity regions using:
- **Masking mode** (default): replaces with `.` or custom char (`--masking-char`).
- **Split mode** (`--extract-high`): splits into high-complexity fragments.
- **Extract mode** (`--extract-low`): extracts low-complexity regions only.
- Entropy-based detection using word size (`--entropy-size`) and threshold.

### 3. Super-k-mer Extraction (`super` subcommand)
Extracts maximal super-k-mers using shared k-mer/minimizer options.

### 4. Index Matching (`match` subcommand)
Matches query sequences against a pre-built index (requires `--index DIRECTORY`).

### 5. Output Formatting & Set Selection
- Supports structured output: `--json-output`, `--csv-output`, or `--yaml-output`.
- Allows filtering by set ID(s) via glob patterns (`--set PATTERN`, repeatable).
- `--force` flag overrides existing destination sets.

### 6. Metadata & Grouping
Per-set metadata can be attached during indexing (`--set-tag KEY=VALUE`).

## Utility Functions
- `CLIKmerSize()`, `CLIEntropyThreshold()` etc.: typed accessors for CLI flags.
- Validation helpers (e.g., `CLIMaskingChar()` ensures single-character input).
- Default minimizer size auto-computation (`DefaultMinimizerSize`).

All options are registered using `go-getoptions`, enabling consistent, self-documenting CLI interfaces across subcommands.
