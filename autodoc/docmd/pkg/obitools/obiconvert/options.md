# Semantic Description of `obiconvert` Package Functionalities

The `obiconvert` package provides command-line interface (CLI) option parsing and configuration utilities for sequence data conversion within the OBItools4 framework. It supports flexible input/output format handling, filtering options, and metadata annotation standards.

### Input Format Support
- Supports multiple input formats: `FASTA`, `FASTQ`, `EMBL`, `GenBank`, `ecoPCR output`, and `CSV`.
- Allows explicit format specification via CLI flags (e.g., `-fasta`, `-fastq`).
- Auto-detection (`guessed`) is used when no format flag is provided.
- Supports structured header annotations in FASTA/FASTQ via:
  - JSON-style (`--input-json-header`)
  - OBI-compliant format (`--input-OBI-header`)

### Output Format & Options
- Outputs can be forced to specific formats: `fasta`, `fastq`, or `json`.
  - Default behavior is format inference based on presence/absence of quality scores.
- Header annotation style for FASTA/FASTQ output follows:
  - JSON (`--output-json-header`)
  - OBI format (`--output-OBI-header`, alias `-O`).
- Optional gzip compression of output files.
- Progress bar display (disabled when stderr is redirected or stdout pipes to another process).

### Data Filtering & Preprocessing
- Skips empty sequences (`--skip-empty`).
- Optional conversion of Uracil (U) to Thymine (T), useful for RNA-to-DNA normalization (`--u-to-t`).
- Supports skipping first *N* records and processing only next *M* (`--skip`, `--only`; commented out but available for future use).
- Option to treat multiple input files as unordered (`--no-order`).

### File Handling
- Configurable output filename via `-out`, `–o`.
- Support for paired-end reads: specify second file with `--paired-with`.

### Integration
- Integrates taxonomy-loading options (`obioptions.LoadTaxonomyOptionSet`).
- Centralized option setter via `OptionSet(allow_paired bool)` for modular CLI setup.

This package enables robust, standardized conversion between biological sequence formats while preserving metadata semantics and supporting common preprocessing workflows.
