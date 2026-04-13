# FASTQ Output Module (`obiformats`)

This Go package provides utilities for formatting and writing biological sequence data in **FASTQ format**. It supports single-end, paired-end, batch processing, and parallelized I/O.

## Core Functionality

- **`FormatFastq(seq, headerFormatter)`**: Formats a single `BioSequence` into FASTQ string.  
- **`FormatFastqBatch(batch, headerFormatter, skipEmpty)`**: Formats a batch of sequences efficiently with dynamic buffer growth and optional skipping/termination on empty reads.

## Header Customization

- Accepts a `FormatHeader` function to inject custom metadata (e.g., read group, sample ID) after the sequence identifier.

## Writing to Streams/Files

- **`WriteFastq(iterator, fileWriter)`**: Writes sequences from an iterator to any `io.WriteCloser`, supporting compression and parallel workers via options.
- **`WriteFastqToStdout(...)`**: Convenience wrapper for stdout output (e.g., piping).
- **`WriteFastqToFile(...)`**: Writes to a file, with support for:
  - Append/truncate modes
  - Paired-end output (splits iterator and writes to two files)
  - Automatic compression via `obiutils.CompressStream`

## Parallelization & Robustness

- Uses goroutines to parallelize formatting/writing across multiple workers.
- Handles empty sequences gracefully: logs warning or fatal error based on `skipEmpty` option.
- Ensures ordered output via batch tracking (`Order()`) and chunked writing.

## Integration

Designed to work seamlessly with the `obitools4` ecosystem:  
- Uses `obiiter.BioSequenceBatch`, `obiseq.BioSequence`, and logging via Logrus.
- Extensible through functional options (`WithOption`) for configuration.

> *Efficient, scalable FASTQ output with support for high-throughput NGS workflows.*
