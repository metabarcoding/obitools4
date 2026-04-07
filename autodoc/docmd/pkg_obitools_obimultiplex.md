# `obimultiplex`: Semantic Description

The `obimultiplex` package enables **high-throughput demultiplexing of PCR amplicon sequencing data**, assigning reads to samples using molecular barcodes (tags) and primer sequences. It supports flexible matching, configurable error tolerance, parallel processing, and optional output of unassigned reads—making it suitable for scalable NGS preprocessing pipelines.

## Core Functionalities

### 1. **NGSFilter Configuration Parsing**
- Reads experiment definitions from a CSV file (`--tag-list` / `-s`) conforming to the `NGSFilter` schema.
- Each row defines: sample name, forward/reverse primer sequences, and one or more barcode (tag) sequences.
- Supports optional metadata columns for custom annotations.

### 2. **Barcode & Primer Matching Engine**
- Uses `obingslibrary` to instantiate a multi-barcode extraction worker.
- Implements three matching modes:
  - `strict`: exact sequence match only;
  - `hamming`: allows mismatches up to a threshold (`--allowed-mismatches` / `-e`);
  - `indel`: extends hamming to permit insertions/deletions (`--with-indels`).
- Default tolerance: ≤2 mismatches; configurable via CLI or programmatic options.

### 3. **Read Assignment & Annotation**
- Assigns each input read to a sample based on successful tag + primer matching.
- Reads failing assignment are flagged with the `"obimultiplex_error"` attribute (unless retained).
- Optional error annotation preserves mismatch/indel details in output metadata.

### 4. **Unidentified Read Handling**
- If `--unidentified` / `-u` is specified, unassigned reads are written to the given file.
- Uses `obiconvert.CLIWriteBioSequences` in a background goroutine for non-blocking I/O.

### 5. **Parallel & Batched Processing**
- Leverages `obidefault` to configure worker threads and batch sizes.
- Applies `.MakeISliceWorker(...)` for concurrent barcode extraction across reads.

### 6. **Template Generation**
- The `--template` option prints a minimal, commented CSV example to stdout for rapid setup.

## CLI Interface Summary

| Option | Alias | Description |
|--------|-------|-------------|
| `--tag-list` / `-s` |  | Path to NGSFilter CSV config (required) |
| `--allowed-mismatches` / `-e` |  | Max mismatches allowed (default: `2`) |
| `--with-indels` |  | Allow indel errors in matching (default: false) |
| `--unidentified` / `-u` |  | Output file for unassigned reads (optional) |
| `--keep-errors` / `--conserved-error` |  | Retain error info in output (default: false) |
| `--template` |  | Print sample CSV template to stdout |

## Design Principles

- **Composability**: Integrates with `obiconvert.OptionSet()` for modular pipeline building.
- **Extensibility**: Extra CSV columns are preserved as read annotations (key-value pairs).
- **Logging & Feedback**: Reports worker count, error handling mode, and output file usage via `logrus`.
- **Dependencies**: Built on top of `obitools4` (`obiformats`, `obingslibrary`) and standard Go CLI tooling.

> **Note**: Only *public* APIs (e.g., `IExtractBarcode`, CLI options, CSV schema) are documented. Internal helpers and low-level workers remain opaque.
