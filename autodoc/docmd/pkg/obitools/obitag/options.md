# `obitag` Package Overview

The `obitag` package provides command-line interface (CLI) utilities and core logic for assigning taxonomic tags to biological sequences using a reference database. It is part of the OBITools4 ecosystem, designed for high-throughput sequence analysis in metabarcoding workflows.

## Key Functionalities

- **Reference Database Handling**:  
  - Loads a reference database from file via `CLIRefDB()`, returning a slice of biological sequences (`BioSequenceSlice`).  
  - Supports saving the loaded (and potentially processed) reference DB to disk with `CLISaveRefetenceDB()`, including optional compression and parallel I/O.

- **CLI Option Parsing**:  
  - `TagOptionSet()` defines required and optional flags:
    - `-R/--reference-db`: Input reference database file (mandatory).
    - `--save-db`: Optional output path to persist the processed DB.
    - `-G/--geometric`: Enables an *experimental* geometric similarity heuristic for faster matching.

- **Integration with OBITools4 Components**:  
  - Leverages `obiconvert`, `obiiter`, `obiseq`, and `obiformats` for sequence I/O, iteration batching, parallelization, and format handling (FASTA/FASTQ/JSON/OBI).
  - Inherits standard conversion options via `obiconvert.OptionSet(false)`.

- **Runtime Configuration Helpers**:  
  - Accessors like `CLIGeometricMode()`, `CLIRefDBName()`, and `CLIRunExact()` expose internal state for downstream processing modules.

- **Performance Optimizations**:  
  - Uses batched iteration (`IBatchOver`) and configurable parallel workers (scaled from total pool).
  - Supports output compression based on global defaults.

## Design Notes

- Heuristic mode (`--geometric`) trades accuracy for speed; exact matching is currently commented out but can be re-enabled.
- The package assumes a pre-built reference DB (e.g., curated barcode library) and focuses on *tagging* rather than alignment or assembly.
- Error handling is strict: panics on DB read failure, fatal logs on write errors.

