# `CSVTaxaIterator` Function — Semantic Description

The function `CSVTaxaIterator`, part of the `obiformats` package, converts a taxonomic iterator (`*obitax.ITaxon`) into an **incremental CSV record generator** via `obiitercsv.ICSVRecord`. It enables streaming, batched export of taxonomic data to CSV format with configurable fields.

### Core Functionality:
- **Input**: A pointer-based taxonomic iterator (`*obitax.ITaxon`) and optional configuration via `WithOption`.
- **Output**: An asynchronous CSV record iterator (`*obiitercsv.ICSVRecord`) that yields batches of records.

### Configurable Output Fields (via options):
- `query`: Taxon-associated query identifier, if enabled (`WithPattern`).
- `taxid`: Either raw node ID (e.g., string pointer) or formatted taxon path (`WithRawTaxid` toggle).
- `parent`: Parent taxonomic ID or string representation, if enabled (`WithParent`).
- `taxonomic_rank`: Taxon rank (e.g., "species", "genus").
- `scientific_name`: Full scientific name of the taxon.
- Custom metadata fields: Specified via `WithMetadata`, extracted from taxon metadata store.
- `path`: Full lineage path (e.g., "k__Bacteria; p__; c__..."), if enabled (`WithPath`).

### Implementation Highlights:
- Uses **goroutines** for non-blocking push of batches and clean shutdown (`WaitAndClose`, `Done`).
- Supports **batching** (configurable via `BatchSize`) to optimize I/O.
- Dynamically builds CSV headers based on selected options before processing begins.

### Use Case:
Efficient, memory-light conversion of large taxonomic datasets (e.g., from classification pipelines) into structured CSV for downstream analysis or reporting.
