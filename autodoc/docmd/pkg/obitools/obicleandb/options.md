# `obicleandb` Package Overview

The `obicleandb` package provides a modular command-line interface for filtering and converting biological sequence data using taxonomic criteria. It integrates core utilities from the OBITools4 suite to support reproducible, taxonomy-aware data curation.

## Core Functionalities

- **Taxonomy Loading**: Enables loading of reference taxonomies (e.g., NCBI, SILVA) via `obioptions.LoadTaxonomyOptionSet`, supporting hierarchical filtering and taxonomic assignment.
- **Input Handling**: Leverages `obiconvert.InputOptionSet` to accept diverse input formats (FASTA, FASTQ, etc.), with automatic format detection and streaming support.
- **Output Generation**: Uses `obiconvert.OutputOptionSet` to produce standardized outputs (e.g., FASTA/FASTQ), with configurable compression and splitting options.
- **Taxonomic Filtering**: Applies `obigrep.TaxonomySelectionOptionSet` to include/exclude sequences based on taxonomic lineage (e.g., `--include-family "Lactobacillaceae"`), enabling precise biological subset extraction.

## Design Principles

- **Composability**: Options are modular and reusable across tools via shared option sets.
- **Extensibility**: New input/output formats or filters can be added without modifying core logic.
- **CLI Consistency**: Aligns with standard `getoptions` conventions for intuitive usage.

This package serves as a foundational building block for clean, taxonomically curated amplicon or metagenomic datasets.
