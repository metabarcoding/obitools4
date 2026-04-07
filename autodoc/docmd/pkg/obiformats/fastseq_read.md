# FastSeq Reader Module — Semantic Description

This Go package (`obiformats`) provides high-performance parsing of FASTA/FASTQ files using a C-backed library (`fastseq_read.h`). It enables streaming, batched reading of biological sequences with optional quality scores.

## Core Features

- **C-based FASTX parsing**: Leverages `kseq.h` via Go's cgo for efficient, low-level file/stream parsing.
- **Batched iteration**: Sequences are grouped into configurable batches (`batch_size`) for memory-efficient processing.
- **Quality score handling**: Supports FASTQ; decodes Phred quality scores using a configurable shift offset (`obidefault.ReadQualitiesShift()`).
- **Source tracking**: Each sequence carries its origin (filename or `"stdin"`), aiding provenance.
- **Header parsing hook**: Optional custom header parser (`ParseFastSeqHeader`) allows metadata extraction or transformation.
- **Full-file batching mode**: When enabled, yields a single batch containing the entire file (useful for small files or global operations).
- **Stdin & File I/O**: Two entry points:
  - `ReadFastSeqFromFile(filename, ...)` for regular files.
  - `ReadFastSeqFromStdin(...)` to process piped input (e.g., from upstream tools).
- **Error resilience**: Gracefully handles missing files, with logging (via `logrus`) for debugging.
- **Async streaming**: Uses goroutines to decouple reading from consumption, enabling concurrent pipelines.

## Integration

Built on top of `obitools4`’s core abstractions:
- `obiiter.IBioSequence`: Iterator interface for biological sequences.
- `obiseq.BioSequence`: Data model holding name, sequence bytes, comment, and quality.
- `obiutils`, `obidefault`: Utilities for path handling and defaults.

Designed for scalability in high-throughput metabarcoding pipelines.
