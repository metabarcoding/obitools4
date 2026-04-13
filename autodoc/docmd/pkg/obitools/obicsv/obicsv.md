# `obicsv` Package: CSV Export Functionality for Biological Sequences

This Go package provides utilities to serialize biological sequence data (e.g., from NGS pipelines) into CSV format.

## Core Functions

- **`CLIWriteSequenceCSV()`**  
  Converts an iterator of `IBioSequence` objects into a CSV-compatible stream. It configures parallelism, batching, and compression using default settings (e.g., `obidefault.ParallelWorkers()`), then applies CLI-driven column mappings via helper functions (`CLIPrintId()`, `CLIPrintSequence()`, etc.). Returns an `ICSVRecord` iterator.

- **`CLICSVWriter()`**  
  Writes the CSV data either to a file (if `obiconvert.CLIOutPutFileName()` ≠ `"-"`) or to standard output. Handles errors with fatal logging and supports optional terminal consumption of the iterator.

## Key Features

- **Flexible column selection**: Controlled by CLI options (e.g., `CSVTaxon`, `CSVKeys`), allowing selective export of metadata, sequences, quality scores.
- **Compression support**: Output can be gzip-compressed per `obidefault.CompressOutput()`.
- **Parallel processing**: Uses ~¼ of configured workers (min 2) for throughput optimization.
- **CLI integration**: Leverages existing `obiconvert` and CLI abstractions for seamless pipeline usage.
- **Error resilience**: Fails fast on I/O issues with descriptive logs.

## Design Notes

Functions follow a functional-iterator pattern, enabling lazy evaluation and streaming. The `terminalAction` flag determines whether the iterator is consumed immediately (e.g., for final output) or returned for further processing.
