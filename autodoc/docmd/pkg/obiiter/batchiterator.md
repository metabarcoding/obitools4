# `obiiter`: Stream-Based Biosequence Iterator Library

This Go package provides a concurrent, batch-oriented iterator for processing large collections of biological sequences (`BioSequence`), designed for high-throughput NGS data pipelines.

## Core Functionality

- **Batched Streaming**: Reads sequences in configurable batches (`BioSequenceBatch`) via a channel-based iterator.
- **Thread Safety**: Uses `sync.WaitGroup`, RWMutex, and atomic flags for safe concurrent access.
- **Lazy Evaluation**: Iteration is on-demand via `Next()`/`Get()`, supporting memory-efficient processing.

## Iterator Management

- **Construction**: `MakeIBioSequence()` initializes a new iterator with default settings.
- **Lifecycle Control**:
  - `Add(n)`, `Done()`: Track active workers (like goroutines).
  - `Lock/RLock` and `Unlock/RUnlock`: Explicit synchronization.
  - `Wait()` / `Close()`, `WaitAndClose()`: Graceful shutdown.

## Batch Transformation & Reorganization

- **`Rebatch(size)`**: Redistributes sequences into fixed-size batches (requires sorting).
- **`RebatchBySize(maxBytes, maxCount)`**: Dynamic batching respecting memory and count limits.
- **`SortBatches()`**: Ensures batches are emitted in strict order (by `order` field).
- **Concatenation & Pooling**:
  - `Concat(...)`: Sequentially merges multiple iterators.
  - `Pool(...)`: Interleaves batches from several sources (preserves order via renumbering).

## Filtering & Predicate-Based Processing

- **`FilterOn(pred, size)`**: Applies a sequence predicate in parallel (configurable workers), recycling discarded sequences.
- **`FilterAnd(pred, size)`**: Same as `FilterOn`, but also checks paired-end consistency.
- **`DivideOn(pred, size)`**: Splits input into two iterators (`true`, `false`) based on predicate.

## Utility & Analysis

- **`Load()`**: Collects all sequences into a single slice (for small datasets).
- **`Count(recycle)`**: Returns `(variants, reads, nucleotides)`.
- **`Consume()` / `Recycle()`**: Drains iterator, optionally triggering sequence recycling.
- **`CompleteFileIterator()`**: Reads entire remaining file as one batch.

## Additional Features

- Supports **paired-end data** via `MarkAsPaired()` / `IsPaired()`.
- Batch ordering preserved for downstream reproducibility.
- Integrates with OBITools4’s `obidefault`, `obiutils` for config and resource management.

> Designed for scalability, low memory footprint, and composability in bioinformatics workflows.
