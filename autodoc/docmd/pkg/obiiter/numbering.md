# `NumberSequences` Function — Semantic Description

The `NumberSequences` method assigns a unique sequential identifier (`seq_number`) to each biological sequence in an `IBioSequence` iterator, preserving consistency for paired-end reads.

## Core Functionality

- **Sequential numbering**: Assigns integers (starting from `start`, defaulting to 0 or user-defined) incrementally across sequences.
- **Thread-safe**: Uses `sync.Mutex` and `atomic.Int64` to safely manage the global counter during concurrent processing.
- **Paired-read support**: When input is paired (`IsPaired()`), both reads in a pair receive the *same* `seq_number`, ensuring alignment between mates.

## Parallelization Strategy

- **Default mode**: Uses multiple workers (`ParallelWorkers()`) for performance; batches are processed concurrently.
- **Reordering mode**: If `forceReordering` is true:
  - Input iterator is batch-sorted (`SortBatches()`).
  - Parallelism disabled (1 worker) to ensure deterministic numbering order.

## Implementation Details

- Each goroutine processes its own split of the input iterator.
- A shared `next_first` counter tracks the next available sequence number globally.
- Locking ensures atomic increment and assignment, preventing race conditions.

## Output

Returns a new `IBioSequence` iterator:
- Contains the same sequence batches (possibly reordered if sorted).
- Each `BioSequence` object now carries a `"seq_number"` attribute.
- Paired sequences are co-numbered and marked accordingly.

## Use Cases

- Preparing data for downstream tools requiring unique sequence IDs.
- Maintaining cross-read identity in paired-end workflows (e.g., assembly, mapping).
- Reproducible numbering across pipeline stages or restarts.
