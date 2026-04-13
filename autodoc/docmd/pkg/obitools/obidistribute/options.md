# `obidistribute` Package Overview

The `obidistribute` package provides command-line interface (CLI) utilities for splitting biological sequence data into multiple output files or directories based on user-defined criteria.

## Core Functionality

- **File Distribution Logic**: Sequences are dispatched to output files or directories using one of three strategies:
  - `--classifier` / `-c`: Distribute by annotation tag (e.g., sample ID, taxonomic assignment).
  - `--directory` / `-d`: Optional companion to classifier — organizes output into subdirectories.
  - `--batches` / `-n`: Split input evenly across *N* batches (round-robin assignment).
  - `--hash` / `-H`: Hash-based distribution into up to *N* batches for reproducible sharding.

- **Flexible Output Naming**: The `--pattern` / `-p` option defines output filenames via a format string (e.g., `"toto_%s.fastq"`), where `%s` is substituted with the classifier value or batch index.

- **Handling Missing Annotations**: The `--na-value` option specifies a fallback label (default `"NA"`) for sequences lacking the classifier annotation.

- **Append Mode**: With `--append` / `-A`, existing files are appended to instead of overwritten.

## Integration

- Leverages `obiconvert` for input/output handling (e.g., FASTQ/FASTA parsing/writing).
- Uses `obiseq.BioSequenceClassifier` internally to abstract distribution logic.
- Built on top of the `obitools4` ecosystem for NGS data processing.

## CLI Design

- Options are registered via `getoptions`, supporting short/long aliases and required checks.
- Validation ensures exactly one distribution mode (`classifier`, `batches`, or `hash`) is selected.
- Filename pattern correctness is verified at runtime to prevent malformed output paths.

## Semantic Summary

This module enables flexible, annotation- or hash-based splitting of sequencing datasets — essential for sample demultiplexing, batch processing, and scalable data management in metabarcoding workflows.
