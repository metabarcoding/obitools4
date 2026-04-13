# Bioinformatics Sequence Processing Pipeline — Public API Overview

The `obiiter` package provides a high-performance, concurrent framework for processing biological sequence data (e.g., FASTQ/FASTA) in batched, streaming fashion. Built around the `IBioSequence` iterator interface and value-type batches (`BioSequenceBatch`), it supports scalable, traceable workflows with built-in memory control, threading safety, and functional composition.

## Core Abstractions

- **`IBioSequence`**: A concurrent iterator over `BioSequenceBatch`, enabling lazy, batched consumption.
- **`BioSequenceBatch`**: An immutable-friendly container holding ordered sequences with metadata (`source`, `order`). Supports FIFO popping, slicing, and pairing.
- **`Pipeable`**: A function type `func(IBioSequence) IBioSequence`, enabling composable transformations.

## Batch & Iterator Management

- `MakeIBioSequence(...)`: Constructs a new iterator (e.g., from files or slices).
- `Concat(...IBioSequence)`: Sequentially merges multiple iterators.
- `Pool(...)`: Interleaves batches from several sources, preserving global order via renumbering.
- `Rebatch(size)` / `RebatchBySize(maxBytes, maxCount)`: Dynamically regroups sequences into fixed or memory-bound batches.
- `SortBatches()`: Ensures strict ordering by batch metadata (`order` field).
- `CompleteFileIterator()`: Reads remaining file content as a single batch.

## Functional Transformations

- `MakeIWorker(...)`, `WorkerPipe(...)`: Applies per-sequence workers in parallel.
- `MakeISliceWorker(...)`, `SliceWorkerPipe(...)`: Applies batch-level (`SeqSliceWorker`) transformations.
- `MakeIConditionalWorker(...)`: Conditional worker application based on a predicate.

## Filtering & Splitting

- `FilterOn(pred, size)`: Parallel filtering with sequence recycling.
- `DivideOn(pred, size)`: Splits input into two independent iterators (`true`/`false` branches).
- `FilterAnd(pred, size)`: Same as above but enforces paired-end consistency.

## Memory & Performance Control

- `LimitMemory(fraction)`: Enforces heap usage ≤ fraction × total RAM via backpressure (uses `runtime.ReadMemStats()`).
- Parallel workers (`nworkers`) and batch sizes are configurable via defaults or variadic args.

## Paired-End Data Handling

All operations preserve pairing semantics:
- `IsPaired()`, `MarkAsPaired()` on iterators and batches.
- `PairTo(other)`: Synchronizes two batch/iterator pairs (same order required).
- `PairedWith()`, `UnPair()` for mate extraction and unpairing.

## Sequence Numbering & Annotation

- `NumberSequences(start, forceReordering)`: Assigns unique sequential IDs to sequences (same ID for mates in paired mode). Supports parallel or deterministic ordering.
- `MakeSetAttributeWorker(rank)`: Returns a worker that annotates each sequence with taxon at specified rank (e.g., `"species"`).

## Taxonomic Profiling

- `ExtractTaxonomy(iterator, seqAsTaxa)`: Aggregates taxonomy across all sequences via `.Slice().ExtractTaxonomy()` calls. Implements map-reduce semantics for scalable taxonomic summarization.

## Fragmentation

- `IFragments(minsize, length, overlap)`: Splits long sequences into overlapping fragments (fusion mode for remainder), with parallel workers and memory-efficient recycling.

## Utility & Analysis

- `Load()`: Collects all sequences into a slice (for small data).
- `Count(recycle)`: Returns `(variants, reads, nucleotides)` counts.
- `Consume()` / `Recycle()`: Drains iterator and optionally triggers sequence recycling.

## Pipeline & Teeing

- `Pipeline(start, parts...)`: Composes a chain of `Pipeable` transformations.
- `(IBioSequence).Pipe(...)`: Fluent method chaining for pipelines.
- `Teeable` / `(IBioSequence).CopyTee()`: Duplicates stream into two independent, concurrently readable iterators (preserves pairing).

## Progress & Logging

- `Speed()`, `SpeedPipe()`: Adds a non-intrusive progress bar (stderr only, terminal-aware). Updates per batch and respects `--no-progressbar` flag.

## Distribution & Routing

- **`IDistribute(classifier, batchSize)`**: Routes sequences to classified outputs based on a classifier function. Batches per class key are flushed when size or memory thresholds are reached.
  - `News()` channel notifies on new output streams (i.e., newly seen class keys).
  - Thread-safe, async distribution via goroutines.

All public APIs assume interoperability with `obiseq`, `obitax`, and OBITools4’s config modules (`obidefault`, `obilog`). Design emphasizes immutability-by-copy, safe concurrent access (via mutexes/atomics), and composability for reproducible bioinformatics pipelines.
