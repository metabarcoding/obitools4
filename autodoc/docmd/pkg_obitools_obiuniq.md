# `obiuniq` Package: Semantic Feature Overview

The `obiuniq` package enables scalable, metadata-aware deduplication of biological sequence data (e.g., NGS amplicons or UMI-tagged reads), supporting both CLI and programmatic use. It groups identical sequences while preserving provenance, abundance counts, and user-defined metadata distinctions.

## Core Functionalities

### Sequence Dereplication
- **Grouping by user-defined attributes** (`--category-attribute`, `-c`): Sequences are collapsed based on one or more metadata fields (e.g., `sequence`, `umi`, `sample`), enabling stratified deduplication.
- **Singleton filtering** (`--no-singleton`, `-n`): Removes groups with only one member, reducing noise from sequencing errors or low-count artifacts.
- **NA value handling** (`--na-value`, `-N`): Replaces missing classifier tags (e.g., unassigned taxonomy) with a configurable placeholder to ensure consistent grouping.

### Scalable & Configurable Processing
- **Chunked I/O** (`--chunk-count`, `--in-memory`): Processes large datasets efficiently using configurable disk-backed or in-memory chunking via the `obichunk` framework.
- **Sorting strategy** (`--on-disk`, `-d`): Switches between in-memory and external sorting to optimize memory usage for large inputs.
- **Parallelization** (`--parallel-workers`): Uses default worker threads to accelerate sorting and grouping steps.

### Statistics & Metadata Preservation
- **Merge statistics** (`--merge`, `-m`): When enabled, records original sequence IDs per group (stored in `_StatsOn`) for lineage tracing and QC.
- **Flexible subcategorization** (`OptionSubCategory`): Allows grouping by multiple metadata keys (e.g., `umi + sample`) to support complex experimental designs.
- **Batch processing** (`--batch-size`, `OptionsBatchSize`): Controls chunk size for memory/performance tuning.

### Programmatic Control
- **CLI state accessors**: Functions like `CLINAValue()`, `CLIKeys()`, and `CLINoSingleton()` expose runtime configuration.
- **Mutable setters**: Enable dynamic tuning (e.g., `SetNAValue()`, `AddStatsOn()`).
- **Integration with OBItools4**: Inherits generic I/O and options (`obiconvert.OptionSet`) for seamless pipeline compatibility.

## `CLIUnique` Function

Implements the main dereplication logic as a streaming iterator over deduplicated sequences (`obiiter.IBioSequence`). Each output sequence carries:
- A count of original occurrences (abundance),
- Merged metadata from input entries,
- Optional per-group statistics when `--merge` is active.

Internally, it orchestrates chunked reading → sorting (in-memory or disk) → grouping → optional filtering — all guided by CLI-configurable parameters. Errors during initialization are logged via `log.Fatal`; runtime issues propagate through the iterator interface.

Designed for high-performance, reproducible deduplication in UMI-aware or multiplexed NGS workflows.
