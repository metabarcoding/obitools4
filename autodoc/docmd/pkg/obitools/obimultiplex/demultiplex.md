# `obimultiplex.IExtractBarcode` — Semantic Description

The function `IExtractBarcode` performs **demultiplexing** of high-throughput sequencing data by extracting and assigning molecular barcodes (e.g., sample indices) to biological sequences.

- **Input**: An iterator over `BioSequence` objects (`obiiter.IBioSequence`) representing raw sequencing reads.
- **Core Logic**: Uses the `obingslibrary` package to configure and instantiate a *multi-barcode extraction worker*.
- **Configuration Options**:
  - `AllowedMismatches`: Tolerates up to *N* mismatches in barcode matching (via `CLIAllowedMismatch()`).
  - `AllowedIndel`: Permits insertions/deletions in barcode alignment (via `CLIAllowsIndel()`).
  - `Unidentified`: If specified, writes unassigned reads to a file (via `CLIUnidentifiedFileName()`).
  - `DiscardErrors`: Controls whether reads failing barcode matching are retained or filtered (via `CLIConservedErrors()`).
  - Parallelization: Uses configurable worker threads and batch sizes (from `obidefault`).

- **Processing Flow**:
  - Applies barcode extraction via `.MakeISliceWorker(...)`, enabling parallel processing.
  - If error conservation is disabled, filters out sequences with the `"obimultiplex_error"` attribute (i.e., unassigned reads).
  - Optionally spawns a goroutine to persist unidentified sequences to disk using `obiconvert.CLIWriteBioSequences`.

- **Output**: Returns an iterator over *assigned* and barcode-extracted sequences, ready for downstream analysis (e.g., merging with primers or taxonomic assignment).

- **Logging**: Provides runtime feedback on worker count, discarded/retained behavior, and output file usage.

This function implements robust, configurable demultiplexing suitable for large-scale NGS pipelines.
