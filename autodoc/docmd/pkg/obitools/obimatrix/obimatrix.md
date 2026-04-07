# MatrixData Module Overview

The `obimatrix` package provides a structured way to build, manipulate, and export biological sequence data matrices (e.g., OTU/ASV tables) in Go.

- **Core Type**: `MatrixData` stores a sparse matrix (`map[row] → map[column]interface{}`), per-row attributes, and metadata (e.g., NA placeholder).
- **Construction**: `MakeMatrixData()` / `NewMatrixData()` initialize the structure with configurable NA value and fixed attribute columns (e.g., `"id"`, `"count"`).
- **Transpose**: `TransposeMatrixData()` flips rows/columns, preserving column IDs under a new `"id"` attribute.
- **Merging**: `MergeMatrixData()` combines two matrices (panics on duplicate row keys).
- **Updating**: `Update(seq, mapkey)` populates a matrix from an `obiseq.BioSequence`, extracting stats (e.g., per-taxon counts) or arbitrary map attributes.
- **Parallel Construction**: `IMatrix()` builds a full matrix from an iterator using parallel workers, auto-detecting extra columns if enabled.
- **Export**:
  - `CLIWriteCSVToStdout()`: writes a wide CSV (rows = sequences, columns = attributes + samples).
  - `CLIWriteThreeColumnsToStdout()`: writes a long-format CSV (`id`, attribute name, value).
- **Flexibility**: Supports customizable attributes (via CLI flags), quality strings (Phred+33/64-aware ASCII encoding), taxonomic labels, and strict mode for missing attributes.
- **Error Handling**: Uses `logrus` to panic on duplicates, type mismatches, or uncastable values.

This module is designed for high-performance processing of metabarcoding datasets in the OBITools4 ecosystem.
