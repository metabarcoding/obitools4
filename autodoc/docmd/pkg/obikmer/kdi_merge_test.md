# K-Way Merge Functionality in `obikmer`

This Go package provides utilities for merging sorted k-mer streams stored in `.kdi` files. Its core component is the `KWayMerge`, which performs a k-way merge of multiple sorted input streams, aggregating duplicate k-mers by counting their occurrences.

## Key Features

- **Sorted K-Mer Input**: Reads k-mers from `.kdi` files via `KdiReader`, assuming each file contains *sorted* 64-bit unsigned integers (`uint64`).
- **K-Way Merge**: Merges multiple sorted streams into a single globally sorted stream using an efficient priority queue (min-heap) internally.
- **Count Aggregation**: When identical k-mers appear across multiple streams, the merge counts how many times each unique k-mer occurs.
- **Memory-Efficient Streaming**: Processes data incrementally, avoiding full loading of all streams into memory.
- **Robust Test Coverage**: Includes unit tests for:
  - Basic merging with overlapping and non-overlapping values.
  - Single-stream input (degenerate case).
  - Empty streams handling.
  - All identical k-mers across inputs.

## API Highlights

- `NewKdiReader(path)` — opens a `.kdi` file for reading.
- `writeKdi(...)` (test helper) — writes sorted k-mers to a `.kdi` file.
- `NewKWayMerge([]*KdiReader)` — constructs the merger from multiple readers.
- `.Next()` → `(kmer uint64, count int, ok bool)` — yields next merged k-mer and its frequency; `ok=false` signals end-of-stream.
- `.Close()` — cleans up resources.

## Use Case

Ideal for aggregating k-mer counts across multiple sequencing samples (e.g., in bioinformatics), where each sample’s k-mers are pre-sorted and persisted, enabling scalable distributed counting without full in-memory deduplication.
