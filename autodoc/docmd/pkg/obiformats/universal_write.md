# `obiformats` Package: Sequence Writing Utilities

This Go package provides utilities for writing biological sequence data to files or standard output in FASTA/FASTQ formats.

## Core Functionality

- **`WriteSequence()`**:  
  Main dispatcher that detects sequence quality data and writes either FASTQ (if qualities present) or FASTA.  
  - Accepts an `IBioSequence` iterator, a writable stream (`io.WriteCloser`), and optional configuration.  
  - Preserves iterator state via `PushBack()` to allow chaining.

- **`WriteSequencesToStdout()`**:  
  Convenience wrapper writing sequences to `stdout`. Automatically closes the output stream.

- **`WriteSequencesToFile()`**:  
  Writes sequences to a specified file. Supports:
    - File creation/truncation or append mode (`OptionAppendFile()`).
    - Paired-end output: writes mate pairs to a second file if `OptionSavePaired()` is enabled.

## Design Highlights

- **Format-Aware Dispatch**: Automatically selects FASTQ vs. FASTA based on presence of quality scores (`HasQualities()`).
- **Iterator Preservation**: Ensures non-consumed sequences remain available after write operations.
- **Error Handling & Logging**: Uses `logrus` for fatal errors during file I/O; returns structured error codes.
- **Configurable Options**: Extensible via `WithOption` pattern (e.g., append mode, paired-end handling).

## Integration

Designed for use within the OBITools4 ecosystem—works with `obiiter.IBioSequence` iterators to support streaming, memory-efficient processing of large sequencing datasets.
