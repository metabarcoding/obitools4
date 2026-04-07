# Paired-End Sequence Handling in `obiiter`

This Go package provides semantic functionality for managing **paired-end biological sequences** within batched iterators.

- `BioSequenceBatch` methods:
  - **`IsPaired()`**: Checks whether the batch contains paired reads.
  - **`PairedWith()`**: Returns a new batch containing only the mate (partner) of each read in the current batch.
  - **`PairTo(*BioSequenceBatch)`**: Synchronizes and pairs reads between two batches *of identical order*; fails if orders differ.
  - **`UnPair()`**: Removes pairing metadata, treating reads as unpaired.

- `IBioSequence` (iterator) methods:
  - **`MarkAsPaired()`**: Marks the iterator as producing paired-end data.
  - **`PairTo(IBioSequence)`**: Combines two iterators into a new paired-end iterator by aligning corresponding batches and calling `PairTo` on each pair.
  - **`PairedWith()`**: Generates a new iterator yielding only the mate reads (i.e., second ends) from an existing paired-end stream.
  - **`IsPaired()`**: Returns whether the iterator was explicitly marked as paired.

All operations preserve batched processing and concurrency via goroutines, ensuring efficient handling of large NGS datasets while maintaining semantic correctness for paired-end workflows.
