# `obiuniq` Package: Semantic Feature Overview

The `obiuniq` package provides command-line and programmatic configuration for deduplicating biological sequence data, grouping identical sequences while preserving metadata-rich distinctions.

## Core Functionality

- **Sequence Grouping**: Groups sequences based on user-defined *category attributes* (`--category-attribute` / `-c`) and optional merging criteria.
- **Singleton Filtering**: Optionally excludes sequences occurring only once (`--no-singleton`), reducing noise from rare artifacts.
- **NA Handling**: Allows custom placeholder (`--na-value`) for missing classifier tags (e.g., taxonomy labels).
- **Scalable Processing**: Uses chunked disk/memory storage (`--chunk-count`, `--in-memory`) to handle large datasets efficiently.

## Configuration API

- **CLI Options**: Built via `getoptions`, exposing flags like `-m` (merge stats), `-c` (grouping keys).
- **State Accessors**: Functions like `CLIKeys()`, `CLINAValue()`, and `CLINoSingleton()` expose runtime configuration.
- **Mutable Setters**: Enables programmatic tuning (e.g., `SetNAValue()`, `AddStatsOn()`).

## Statistics & Metadata

- **Merged Attributes**: Tracks original IDs per group via `--merge` (`_StatsOn`) — useful for provenance and QC.
- **Flexible Grouping**: Supports multiple attributes (e.g., `sequence`, `umi`, `sample`) to define *identity* beyond raw sequence.

## Integration

- Extends generic I/O options from `obiconvert.OptionSet`, ensuring compatibility with OBItools4 pipelines.

> Designed for high-performance, metadata-aware deduplication in NGS workflows (e.g., amplicon or UMI-based data).
