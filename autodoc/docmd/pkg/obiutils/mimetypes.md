# OBIMimeUtils: Semantic Description of Features

The `obiutils` Go package provides utilities for detecting and handling biological data file formats, primarily via MIME type inference.

## Core Functionalities

- **BOM Detection (`HasBOM`)**  
  Identifies Byte Order Marks (BOMs) for UTF-8, UTF-16 BE/LE, and UTF-32 BE/LE encodings. Logs detected types for transparency.

- **Last-Line Trimming (`DropLastLine`)**  
  Removes the final newline-delimited line from a byte slice — useful for sanitizing incomplete or truncated files.

- **MIME Type Registration (`RegisterOBIMimeType`)**  
  Extends generic MIME types (e.g., `text/plain`, `application/octet-stream`) with format-specific detectors for:
  - **CSV**: Validates structured comma-separated data (≥2 fields, ≥2 lines).
  - **FASTA/FASTQ**: Regex-based detection of sequence headers (`>` or `@`).
  - **ecoPCR2**: Detects files starting with the magic header `#@ecopcr-v2`.
  - **GenBank/EMBL**: Checks for standard sequence record prefixes (`LOCUS`, `ID`).

- **Format-Specific Extensions**  
  Registers custom MIME subtypes (e.g., `text/fasta`, `.fasta`) and associates them with appropriate file extensions.

- **Idempotent Registration**  
  Ensures MIME detectors are registered only once using a guard flag.

## Design Goals

- Robust, lightweight format inference without full parsing.
- Extensible architecture for future bioinformatics formats.
- Logging-friendly (via `logrus`) to aid debugging and observability.

This package enables accurate, context-aware MIME detection in pipelines handling heterogeneous biological data.
