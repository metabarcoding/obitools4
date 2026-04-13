# `obik filter`: K-mer Index Filtering Subcommand

The `runFilter` function implements the `obik filter` CLI command, enabling users to apply configurable filters to an existing k-mer index and generate a new filtered version.

## Core Functionality

- **Input**: Reads an existing k-mer index (`<source_index>`) via `obikmer.OpenKmerSetGroup`.
- **Output**: Writes a new index (`--out <dest_index>`) containing only k-mers that pass all specified filters.
- **Parallelism**: Filters partitions in parallel using goroutines; each worker instantiates its own filter (to support stateful filters like entropy-based ones).

## Supported Filters

- **Entropy Filter** (`--entropy-filter`):
  - Removes low-complexity k-mers using a sliding-window entropy metric.
  - Configurable via `--entropy-threshold` and `--entropy-size`.
  - Implemented by wrapping `obikmer.NewKmerEntropyFilter`.

## Filtering Architecture

- Uses a **factory pattern** (`KmerFilterFactory`) to generate per-worker filter instances.
- `chainFilterFactories` composes multiple filters with logical AND semantics (all must accept a k-mer).
- Filters are applied per-partition (`filterPartition`) using `KdiReader`/`KdiWriter`.

## Set & Partition Handling

- Supports selection of specific sets via `--set-patterns`.
- Processes all partitions (`P`) per set, preserving original partitioning structure.
- Preserves `spectrum.bin` files (if present) in the output.

## Metadata & Reporting

- Copies group-level metadata and records applied filters (e.g., entropy threshold).
- Logs per-set statistics: total processed, kept k-mers, and removal percentage.
- Uses `progressbar` for interactive progress feedback (when enabled).
