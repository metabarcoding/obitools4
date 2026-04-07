# Semantic Description of `IMergeSequenceBatch` and `MergePipe`

This code defines two related functions in the `obiiter` package for batch-wise merging of biological sequences during iteration.

- **`IMergeSequenceBatch(na, statsOn, sizes...) IBioSequence → IBioSequence`**  
  - Consumes an input sequence iterator (`IBioSequence`) and returns a new one.
  - Groups incoming sequences into batches (default size: `100`, configurable via variadic argument).
  - For each batch:
    - Collects up to `batchsize` sequences via the input iterator.
    - Applies `.Merge(na, statsOn)` on each sequence group (presumably merging reads based on `na`, e.g., nucleotide alignment or overlap).
    - Wraps merged results into a `BioSequenceBatch` with ordering metadata.
  - Emits batches asynchronously via goroutines; the output iterator is closed when input finishes.

- **`MergePipe(na, statsOn, sizes...) Pipeable → func(IBioSequence) IBioSequence`**  
  - A *pipeline combinator* (higher-order function), enabling functional composition.
  - Returns a `Pipeable` — i.e., a transformation function compatible with iterator pipelines.

**Semantic Purpose**:  
Enables efficient, memory-smoothed merging of biological sequence reads (e.g., paired-end merges) in streaming fashion, with optional statistics tracking (`statsOn`) and configurable batching.
