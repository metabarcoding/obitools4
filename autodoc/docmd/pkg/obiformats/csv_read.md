# CSV Import Module for Biological Sequences (`obiformats`)

This Go package provides functionality to parse biological sequence data from CSV files into structured objects compatible with the OBItools4 framework.

## Core Features

- **CSV Parsing**: Reads CSV data via `io.Reader`, supporting comments (`#`), flexible field counts, and leading-space trimming.
- **Sequence Extraction**: Identifies columns named `sequence`, `id`, or `qualities` by header and maps them to corresponding biological sequence fields.
- **Quality Score Adjustment**: Applies a configurable Phred score shift (default: `33`) to quality strings.
- **Metadata Handling**:
  - Special handling for taxonomic IDs (`taxid`, `*_taxid`).
  - Generic attributes parsed as JSON when possible; fallback to raw string otherwise.
- **Batched Output**: Streams sequences in configurable batches (`batchSize`) via an iterator interface (`obiiter.IBioSequence`).
- **Multiple Entry Points**:
  - `ReadCSV`: From any `io.Reader`.
  - `ReadCSVFromFile`: Loads from a file (with source naming derived from filename).
  - `ReadCSVFromStdin`: Reads from standard input.
- **Error & Edge Handling**:
  - Gracefully handles empty files/streams via `ReadEmptyFile`.
  - Uses structured logging (Logrus) for fatal and informational messages.

## Integration

Designed to integrate with OBItools4’s core types:
- `obiseq.BioSequence`: Holds sequence, ID, qualities, taxid, and arbitrary attributes.
- `obiiter.IBioSequence`: Streaming interface for batched sequence iteration.

## Use Case

Efficient, flexible ingestion of tabular biological data (e.g., from alignment outputs or preprocessed FASTQ/FASTA conversions) into downstream analysis pipelines.
