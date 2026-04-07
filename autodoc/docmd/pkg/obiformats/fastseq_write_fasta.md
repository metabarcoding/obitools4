# `obiformats` Package Overview

The `obiformats` package provides utilities for formatting and writing biological sequences (e.g., DNA, RNA) in standard formats—primarily **FASTA**. It is designed for high-performance batch processing and supports parallel I/O, compression-aware streaming, and flexible configuration.

## Core Formatting Functions

- **`FormatFasta(seq, formater)`**  
  Converts a single `BioSequence` into a FASTA string: header (`>id description`) followed by sequence lines of up to 60 characters.

- **`FormatFastaBatch(batch, formater, skipEmpty)`**  
  Efficiently formats a batch of sequences into FASTA using pre-allocated buffers and direct byte writes—avoiding intermediate strings. Empty sequences are either skipped (with warning) or cause a fatal error.

## File Writing Functions

- **`WriteFasta(iterator, file, options...)`**  
  Writes a stream of sequences to any `io.WriteCloser`. Supports:
  - Parallel workers (`ParallelWorkers`)
  - Chunked writing via `WriteFileChunk`
  - Optional compression (e.g., gzip)  
  Returns a new iterator mirroring the input for pipeline chaining.

- **`WriteFastaToStdout(iterator, options...)`**  
  Convenience wrapper to output FASTA directly to `stdout`, with file-closing behavior configurable.

- **`WriteFastaToFile(iterator, filename, options...)`**  
  Writes to a named file with:
  - Truncation or append mode (`AppendFile`)
  - Automatic paired-end output if `HaveToSavePaired()` is enabled  
    (writes reverse reads to a secondary file specified via `PairedFileName`)

## Key Design Highlights

- **Memory-efficient**: Uses `bytes.Buffer.Grow()` and avoids unnecessary allocations.
- **Robust error handling**: Panics on nil sequences; logs warnings/errors via `logrus`.
- **Pipeline-friendly**: Integrates with the `obiiter` iterator abstraction for streaming workflows.
