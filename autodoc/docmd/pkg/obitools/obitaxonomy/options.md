# Taxonomy Processing CLI Module (`obitaxonomy`)

This Go package provides a command-line interface (CLI) for interacting with taxonomic data, built on top of the OBItools4 framework. It supports flexible querying, filtering, and export of taxonomic trees.

## Core Functionalities

- **Taxonomy Loading & Management**: Integrates with `obitax` to load and manage taxonomic databases (e.g., NCBI).
- **Taxon Filtering**: Allows restricting output to specific clades via `--restrict-to-taxon` (`-r`) using taxon IDs or names.
- **Rank-Based Filtering**: Restricts output to a specific rank (e.g., `species`, `genus`) with `-rank` (`--rank`).
- **Tree Navigation**:
  - `--parents` (`-p`) displays the full lineage (path) for a given taxon ID.
  - `--sons` (`-s`) lists all direct children of a given taxon ID.
  - `--dump` (`-D`) exports the entire subtree rooted at a given taxon.
- **Output Formatting**:
  - Columns can be toggled: scientific name (`--without-scientific-name`), taxonomic rank (`-R`, `--without-rank`), parent ID (via `-W`, implied via `--without-parent`).
  - Full taxonomic path (`-P`) and matching query source (`--with-query`) can be included.
  - Supports Newick tree output (`-N`, `--newick-output`) with optional leaf labels and root trimming.
- **Data Acquisition**:
  - `--download-ncbi`: Fetches and installs the latest NCBI taxonomy dump.
  - `--extract-taxonomy`: Extracts taxonomic labels from sequence files (e.g., FASTA/FASTQ).
- **Pattern Matching**:
  - `--fixed` (`-F`) enables literal (non-regexp) taxon name matching.
  - `--rank-list` (`-l`) prints all available ranks in the loaded taxonomy.

## Utility Functions

Helper functions (e.g., `CLIRankRestriction()`, `CLIWithScientificName()`), expose parsed CLI flags for downstream processing modules.

## Integration

Designed to be composed with `obiconvert` (output formatting) and standard OBItools4 option parsing (`getoptions`). Fully modular, extensible for taxonomic workflows in metagenomics and biodiversity informatics.
