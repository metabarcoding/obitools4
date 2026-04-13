# `obiannotate` Package: Semantic Description of Features

The `obiannotate` package provides a rich set of command-line options for annotating, transforming, and filtering biological sequence records (e.g., FASTA/FASTQ). It integrates with `obiconvert` and `obigrep`, extending functionality via structured metadata manipulation.

## Core Annotation Features
- **Metadata Clearing**: `--clear` removes all existing attributes.
- **Sequence Metadata Injection**:
  - `--length`: Adds a `seq_length` attribute.
  - `--number`: Assigns an ordinal index (`seq_number`) starting at 1.
- **Taxonomic Annotation**:
  - `--taxonomic-path`: Adds full lineage path (e.g., "cellular organisms; Bacteria; ...").
  - `--taxonomic-rank`: Adds taxonomic rank (e.g., "species", "genus").
  - `--scientific-name`: Adds the scientific name (e.g., *Homo sapiens*).
  - `--with-taxon-at-rank RANK`: Extracts and adds taxon at a specific rank (e.g., `--with-taxon-at-rank species`).
  - `--add-lca-in SLOT`: Computes and injects the Lowest Common Ancestor (LCA) taxid into a named slot, with tolerance via `--lca-error`.

## Pattern & Sequence Manipulation
- **Pattern Matching** (`--pattern`, `--aho-corasick`):
  - Simple regex-like pattern matching with error reporting (`pattern_match`, `pattern_error` slots).
  - Efficient multi-pattern search using Aho-Corasick automaton (file-based input).
- **Sequence Editing**:
  - `--cut start:end`: Trims sequence to specified positions (1-based; supports open-ended via empty bounds).
  - `--set-identifier EXPRESSION`: Dynamically assigns new IDs using Python-like expressions.

## Attribute Management
- **Rename/Delete/Keep**:
  - `--rename-tag NEW=OLD`: Renames attributes (skips records if OLD is missing).
  - `--delete-tag KEY`: Removes specified attribute(s) (skips if absent).
  - `--keep KEY` (`-k`): Retains only specified attributes.
- **Dynamic Attribute Creation**:
  - `--set-tag KEY=EXPRESSION` (`-S`): Computes new attributes from expressions (e.g., `tag1=seq_length > 200`).

## Utility & Validation
- Helper functions expose internal state (e.g., `CLIHasPattern()`, `CLICut()`).
- Robust parsing with logging and error handling (e.g., invalid cut format triggers fatal exit).

This package enables flexible, scriptable annotation workflows for high-throughput sequencing data in the OBITools4 ecosystem.
