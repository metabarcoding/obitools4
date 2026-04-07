# `obiconvert` Package: Semantic Overview

This Go package provides utilities for writing biological sequence data to files or stdout, supporting multiple formats and parallel processing.

## Core Functionality

- **`BuildPairedFileNames(filename string) (string, string)`**  
  Derives paired-end filenames from a base name by appending `_R1` and `_R2`, preserving directory path and file extension (e.g., `sample.fastq → sample_R1.fastq`, `sample_R2.fastq`).

- **`CLIWriteBioSequences(...)`**  
  Writes `IBioSequence` iterator output to disk or stdout, based on CLI-configured options:
  - **Format support**: FASTQ, FASTA, JSON (default), or generic sequence format.
  - **Header style**: Configurable via `CLIOutputFastHeaderFormat()` — supports `"json"` or `"obi"`.
  - **Parallelism**: Uses `WriteParallelWorkers()` for concurrent I/O.
  - **Batching & compression**: Controlled by batch size and output-compression flags.

## Key Behaviors

- If no filename is given or `"-"` is used, output goes to **stdout**.
- For paired data (`iterator.IsPaired()`), automatically writes R1/R2 to separate files.
- Skips empty sequences if `CLISkipEmpty()` returns true.
- On terminal actions (`terminalAction == true`), recycles resources and returns `nil`.
- Logs critical errors with `log.Fatalf`.

## Integration

Built on top of:
- `obiformats`: Format-specific writers (FASTQ/FASTA/JSON).
- `obiiter`: Sequence iterator abstraction.
- `obidefault`: CLI-default configuration (workers, batch size, compression).
