# `obimatrix` Package: Semantic Overview

The `obimatrix` package enables high-performance construction, manipulation, and export of biological sequence count matrices (e.g., OTU/ASV tables) in the OBITools4 ecosystem. Built around a sparse matrix representation, it supports flexible attribute handling, parallelized input processing, and multiple output formats—ideal for downstream ecological or bioinformatic analysis.

## Core Functionalities

### Matrix Construction & Management
- **`MakeMatrixData()` / `NewMatrixData(naVal string, fixedCols []string)`**:  
  Initializes a new `MatrixData` instance with configurable NA placeholder and fixed column headers (e.g., `"id"`, `"count"`).
- **`Update(seq obiseq.BioSequence, mapKey string)`**:  
  Populates the matrix using a biological sequence’s annotations. Extracts per-taxon counts or arbitrary map attributes (e.g., sample IDs), inserting them into the sparse matrix under `row = seq.ID`, with dynamic column detection.
- **`TransposeMatrixData(md *MatrixData)`**:  
  Flips rows/columns: original columns become new `"id"` attributes; preserves metadata and NA handling.

### Merging & Parallelization
- **`MergeMatrixData(a, b *MatrixData)`**:  
  Combines two matrices row-wise; panics on duplicate sequence IDs to prevent silent overwrites.
- **`IMatrix(iter obiseq.Iterator, mapKey string)`**:  
  Builds a full matrix in parallel from an iterator of sequences. Auto-detects extra sample columns if enabled (via `--auto-cols`), supporting dynamic batch processing.

### Export & CLI Integration
- **`CLIWriteCSVToStdout(md *MatrixData)`**:  
  Outputs a wide-format CSV: rows = sequences, columns = fixed attributes + detected samples. Handles Phred encoding (ASCII 33/64) for quality strings and supports transpose via `--transpose`.
- **`CLIWriteThreeColumnsToStdout(md *MatrixData)`**:  
  Outputs a long-format CSV with columns: `sample`, sequence ID, and value—suited for tools expecting tidy data.
- **CLI Option Aggregation**:  
  Integrates with `getoptions` to expose flags like:
    - `-m, --map-attribute`: grouping key (default: `"merged_sample"`)
    - `--value-name`, `--sample-name`: column headers (defaults: `"count"`, `"sample"`)
    - `-t, --transpose`: toggle row/column orientation
    - `--allow-empty`, `--strict-attributes`: control handling of missing annotations

### Robustness & Flexibility
- **NA Handling**: Replaces absent mapping attributes with a configurable placeholder (default: `"0"`).
- **Strict Mode**: Panics on type mismatches or uncastable values (e.g., non-numeric counts in numeric context).
- **Attribute Extensibility**: Supports arbitrary metadata (taxonomic labels, quality strings) via dynamic column inference.

## Design Philosophy

Focused on **speed**, **type safety**, and **reproducibility** for amplicon sequencing workflows. The package avoids implicit defaults beyond core conventions, favoring explicit CLI configuration and clear error signaling for data integrity.
