# `obiitercsv`: CSV Record Iterator for Streaming and Batch Processing

A Go package providing a thread-safe, channel-based iterator (`ICSVRecord`) for efficient streaming and batch processing of CSV data. Designed with scalability in mind—especially for bioinformatics pipelines like OBITools4—it enables ordered, concurrent handling of large CSV files without loading all records into memory.

## Core Concepts

- **`CSVHeader`**: A `[]string` representing column names; used to define the schema of records.
- **`CSVRecord`**: A `map[string]interface{}` mapping field names to values, supporting flexible typed data.
- **`CSVRecordBatch`**: A structured batch of records (`[]*CSVRecord`) enriched with metadata:
  - `source`: origin identifier (e.g., file or shard name),
  - `order`: sequence index for deterministic reassembly,
  - `data`: the slice of records.

## Iterator Interface (`ICSVRecord`)

Implements a standard iterator protocol over batches via an unbuffered channel:

- **`Next() bool`**: Advances to the next batch; returns `false` when exhausted.
- **`Get() *CSVRecordBatch`**: Retrieves the current batch (nil-safe).
- **`PushBack()`**: Requeues the last retrieved batch for reprocessing—useful in error recovery or conditional branching.
- **`Channel() <-chan *CSVRecordBatch`**: Exposes the internal channel for external consumption.

## Thread-Safe Operations

- All shared state (e.g., batch queue, flags) is guarded by a `sync.RWMutex`.
- Atomic operations (`atomic.Bool`, `int32`) are used for lightweight flags like `finished` and counters such as `batch_size`.
- Methods ensure safe concurrent access across multiple goroutines.

## Header Management

Supports dynamic schema evolution:

- **`SetHeader(header CSVHeader)`**: Sets or replaces the header (must be called before first `Next()`).
- **`AppendField(name string, value interface{}) bool`**: Adds a new field to the current record (returns `false` if no active batch or header mismatch).

## Batch Lifecycle Control

- **`Add()` / `Done()`**: Track active producer/consumer goroutines using a `sync.WaitGroup`.
- **`WaitAndClose()`**: Blocks until all tracked goroutines complete, then closes the output channel—ensuring no data loss.

## Utility & Validation

- **`NotEmpty(batch *CSVRecordBatch) bool`**: Returns `true` if the batch is non-nil and contains ≥1 record.
- **`IsNil(batch *CSVRecordBatch) bool`**: Returns `true` if the batch is nil.
- **`Consume(iterator ICSVRecord, fn func(*CSVRecordBatch))`**: Drains the iterator by applying `fn` to each batch—ideal for side-effect processing (e.g., writing, aggregation).

## Ordering & Recovery

- **`SortBatches(batches []*CSVRecordBatch) [](*CSVRecordBatch)`**: Reorders batches by `order`, buffering out-of-sequence items until missing predecessors arrive—critical for reconstructing global order in distributed or parallel pipelines.

## Splitting & Sharing

- **`Split() ICSVRecord`**: Creates a new iterator instance sharing the same underlying channel but with independent locking—enables fan-out patterns without duplicating data.

## Design Goals

- **Memory efficiency**: Processes records in streaming batches, avoiding full-file loads.
- **Deterministic ordering**: Supports reconstruction of sequential order despite concurrent delivery.
- **Robustness**: Graceful handling of race conditions, nil states, and partial batches.

> *Intended for high-throughput CSV pipelines where correctness, concurrency safety, and low latency are paramount.*
