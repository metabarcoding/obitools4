# CSV Export Functionality Overview

This Go package (`obicsv`) provides command-line interface options and utilities for exporting biological sequence data to CSV format. It integrates with the OBITools4 framework, supporting flexible attribute selection and formatting.

## Core Export Options
- **`--ids/-i`**: Outputs sequence identifiers.
- **`--sequence/-s`**: Includes raw nucleotide/amino acid sequences.
- **`--quality/-q`**: Adds per-base quality scores (e.g., Phred values).
- **`--definition/-d`**: Prints sequence headers or definitions.
- **`--count`**: Includes abundance/observation counts per sequence.

## Taxonomic & Pairing Data
- **`--taxon`**: Exports NCBI taxid and corresponding scientific name.
- **`--obipairing`**: Includes metadata added by `obipairing`, such as alignment mode, score, and mismatch count.

## Attribute Filtering
- **`--keep/-k KEY`**: Restricts output to specified attributes (multiple `-k` allowed).
- **`--auto`**: Inspects first records to auto-detect and suggest relevant attributes.
  
## Configuration
- **`--na-value NAVALUE`**: Sets placeholder string (default `"NA"`) for missing fields.

## Integration
- Extends `obiconvert` input/output and taxonomy-loading options.
- Provides CLI accessor functions (e.g., `CLIPrintSequence()`, `CLIHasToBeKeptAttributes()`).
- Supports soft attribute groups (e.g., `"obipairing"` expands to 8 specific fields).

Designed for high-throughput sequence analysis pipelines, enabling customizable tabular output compatible with downstream tools.
