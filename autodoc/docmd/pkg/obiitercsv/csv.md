# `obiitercsv`: CSV Record Iterator for Streaming and Batch Processing

This Go package provides a thread-safe, channel-based iterator (`ICSVRecord`) for streaming and processing CSV records in batches. It supports ordered batch handling, concurrent access via mutexes, and dynamic header management.

## Core Types

- **`CSVHeader`**: A slice of strings representing column names.
- **`CSVRecord`**: A map from field name to value (`map[string]interface{}`).
- **`CSVRecordBatch`**: A batch of records with metadata: `source`, `order`, and the actual data slice.

## Key Features

- **Streaming via Channels**: Records are consumed as `CSVRecordBatch` items through a channel, enabling asynchronous producers/consumers.
- **Ordered Processing**: Batches include an `order` field, used by `SortBatches()` to reconstruct sequential order even when received out-of-order.
- **Thread Safety**: Uses `sync.RWMutex`, atomic operations (`batch_size`), and `abool.AtomicBool` for flags like `finished`.
- **Iterator Protocol**: Implements standard methods:  
  - `Next()` to advance,  
  - `Get()` to retrieve current batch,  
  - `PushBack()` for re-queuing the last record.
- **Batch Management**:  
  - `SetHeader()` / `AppendField()`: dynamic header updates.  
  - `Split()`: creates a new iterator sharing the same channel but with independent locking.
- **Lifecycle Control**:  
  - `Add()` / `Done()`: track active goroutines (via `sync.WaitGroup`).  
  - `WaitAndClose()` ensures all data is flushed before closing the channel.

## Utility Methods

- **`NotEmpty()`, `IsNil()`**: Check batch validity.
- **`Consume()`**: Drains the iterator (e.g., for side-effect processing).
- **`SortBatches()`**: Reorders batches by `order`, buffering out-of-sequence ones.

Designed for bioinformatics pipelines (e.g., OBITools4), it enables scalable, memory-efficient CSV processing with strict ordering guarantees.
