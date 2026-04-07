# `obichunk`: High-Performance Chunking and Dereplication of Biological Sequences

The `obichunk` package provides scalable, configurable infrastructure for preprocessing large-scale biological sequence data (e.g., FASTA/FASTQ). It enables efficient grouping, sorting, deduplication, and batched streaming of sequences—critical for metabarcoding, metagenomics, or any high-throughput NGS workflow.

---

## Core Functionalities

### `ISequenceChunk`
Unified entry point for sequence chunking, supporting both **in-memory** and **on-disk** execution modes.  
- Accepts an `obiiter.IBioSequence` iterator and a classifier (`obiseq.BioSequenceClassifier`).  
- Mode selection via `onMemory` flag: routes to either `ISequenceChunkOnMemory` or `ISequenceChunkOnDisk`.  
- Optional parameters:
  - `dereplicate`: deduplicate identical sequences per batch.
  - `na`: defines placeholder for missing/ambiguous characters (e.g., `"N"`, `"?"`).
  - `statsOn`: enables metadata tracking (e.g., sample IDs, primer names) for statistics.
  - `uniqueClassifier`: optional secondary classifier to assign unique labels.

Returns an iterator over processed sequences (`obiiter.IBioSequence`), supporting streaming pipelines and downstream integration.

---

### `ISequenceChunkOnDisk`
Efficiently splits sequences into **temporary on-disk batches** (`.fastx` files), ideal for large datasets.  
- Automatically manages a temp directory (`obiseq_chunks_*`) and cleans up post-processing.
- Uses `find` to discover all generated chunk files recursively.
- Asynchronous streaming: batches are yielded via an iterator as they’re written, decoupling production and consumption.
- Optional per-batch dereplication using composite keys (sequence + classification).
- Logs batch count and start events for monitoring.

Internally leverages:
- `obiiter.MakeIBioSequence()` to build output iterator.
- `obiformats.WriterDispatcher` for parallel file writing.
- A dedicated goroutine to read, classify/dereplicate, and emit batches.

---

### `ISequenceChunkOnMemory`
Performs **in-memory parallel chunking** of sequences into classification-based batches.  
- Routes each sequence to a bucket (flux) using the classifier.
- Maintains one `BioSequenceSlice` per classification group in memory (thread-safe via mutex).
- Emits batches **only after full input consumption**, preserving deterministic batch order (0, 1, …).
- Parallel processing: each flux handled in its own goroutine.
- Fails fast on internal errors (e.g., channel issues) via `log.Fatalf`.

Ideal for RAM-sufficient workloads requiring low-latency, ordered batch output.

---

### `Options` System
Configurable pipeline behavior via functional options pattern.  
- Immutable configuration builder: `MakeOptions([]WithOption)` applies setters to internal struct.
- Key options:
  - **Categorization**: `OptionSubCategory(...)` appends sample/marker labels; `PopCategories()` retrieves first.
  - **Missing values**: `OptionNAValue(na)` customizes placeholder (default: `"?"`).
  - **Statistics**: `OptionStatOn(...)` registers fields for metadata tracking.
  - **Batching**:
    - `OptionBatchCount(n)` sets number of batches (e.g., for hashing).
    - `OptionsBatchSize(size)` defines items per batch.
  - **Concurrency**: `OptionsParallelWorkers(n)`.
  - **Sorting strategy**:
    - `OptionSortOnDisk()` enables disk-backed sorting.
    - `OptionSortOnMemory()` (default) uses RAM-based sort.
  - **Singleton filtering**:
    - `OptionsNoSingleton()` excludes singleton reads (count = 1).
    - `OptionsWithSingleton()` allows them.

Defaults drawn from `obidefault`, ensuring reproducibility and ease of use.

---

### `ISequenceSubChunk`
Parallel, class-based sorting and re-batching of sequence batches.  
- Input: iterator over `BioSequenceBatch`, classifier, and worker count.
- For each batch:
  - If size >1: sequences are sorted *in-place* by classification code (via custom `sort.Interface`).
  - Consecutive sequences with same class are regrouped into new batches.
- Uses atomic counters (`nextOrder`) to assign globally increasing order IDs across workers—ensuring deterministic inter-batch ordering.
- Preserves input-order *within* each new batch.

Use case: preparing sorted, class-homogeneous batches for downstream tasks (e.g., consensus calling or alignment).

---

### `IUniqueSequence`
End-to-end **dereplication** pipeline: groups identical sequences, aggregates counts and metadata.  
- Input iterator + optional `Options`.
- Parallelization via configurable workers (falls back to single-threaded if disk sorting enabled).
- **Splitting phase**:
  - Uses `HashClassifier` to partition input deterministically (controlled by `BatchCount`).
- **Storage selection**:
  - In-memory: via `ISequenceChunkOnMemory`.
  - On-disk: uses `ISequenceSubChunk` + external sort (single worker required).
- **Uniqueness logic**:
  - Composite classifier: sequence identity + optional annotations (sample, primer).
  - NA handling for missing annotation fields.
- **Singleton filtering**: optionally excludes reads with count =1 (`NoSingleton()`).
- **Parallel deduplication**:
  - Workers process chunks via `ISequenceSubChunk` + per-group aggregation.
- **Merging**:
  - Aggregates results via `IMergeSequenceBatch`, preserving counts, stats, and ordering.

Scalable from small datasets to terabyte-scale NGS runs.
