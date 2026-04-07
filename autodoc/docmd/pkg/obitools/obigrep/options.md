# `obigrep`: Sequence Filtering Module for OBITools4

The `obigrep` package provides a rich set of command-line-driven filtering capabilities for biological sequence records (e.g., FASTA/FASTQ), built on top of the OBITools4 framework. It enables users to select or exclude sequences based on diverse criteria, including sequence content, metadata attributes, taxonomy, abundance, and pattern matching (exact or approximate).

## Core Functionalities

- **Sequence Length & Abundance Filtering**:  
  Select sequences by minimum/maximum length (`--min-length`, `--max-length`) and count (abundance; `--min-count`, `--max-count`).

- **Pattern Matching**:  
  Supports exact regex matching against:
  - Sequence (`--sequence`, `-s`)
  - Definition line (`--definition`, `-D`)
  - Identifier (`--identifier`, `-I`)  
  Case-insensitive by default.

- **Approximate Pattern Matching**:  
  Allows fuzzy matching with configurable error tolerance (`--pattern-error`), indels (`--allows-indels`), and strand orientation (`--only-forward`). Uses `obigrep`’s approximate pattern engine.

- **Taxonomic Filtering**:  
  - Restrict to specific taxa (`--restrict-to-taxon`, `-r`)
  - Exclude taxa (`--ignore-taxon`, `-i`)  
  Validate taxonomy presence/consistency (`--valid-taxid`)
  - Require specific taxonomic ranks (`--require-rank`)

- **Attribute-Based Selection**:  
  Filter by presence of attributes (`--has-attribute`, `-A`) or match attribute values with regex (`--attribute=key=pattern`, `-a`).

- **Identifier List Filtering**:  
  Load identifiers from a file (`--id-list`) to select only those records.

- **Custom Predicate Expressions**:  
  Evaluate arbitrary boolean expressions per sequence (`--predicate`, `-p`), with access to attributes and the `sequence` object.

- **Paired-End Read Handling**:  
  Controls how conditions apply to read pairs via `--paired-mode` (e.g., `"forward"`, `"and"`, `"xor"`).

- **Output Control**:  
  Save rejected sequences to file (`--save-discarded`) or invert selection globally (`--inverse-match`, `-v`).

## Architecture

- All options are parsed via `go-getoptions`.
- Filtering logic is composed into a single predicate (`CLISequenceSelectionPredicate`) using logical AND/OR composition.
- Taxonomy-aware predicates leverage `obitax`, sequence operations use `obiseq`, and utilities (e.g., file I/O) rely on `obiutils`.
- Integration with conversion pipelines via `obiconvert.OptionSet`.

This module serves as the backbone for selective data extraction in metagenomic and metabarcoding workflows.
