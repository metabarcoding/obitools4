# Semantic Description of `ReadSequencesBatchFromFiles`

This function implements **concurrent, batched streaming** of biological sequences from multiple input files.

## Core Functionality

- **Input**: A slice of file paths (`[]string`), an optional batch reader interface, and a concurrency level.
- **Default behavior**: Uses `ReadSequencesFromFile` if no custom reader is provided.

## Concurrency Model

- Launches `concurrent_readers` goroutines to process files in parallel.
- Files are distributed via a shared channel (`filenameChan`) — ensuring fair load balancing.

## Streaming Interface

- Returns an `obiiter.IBioSequence`, a streaming iterator over batches of biological sequences.
- Internally uses an atomic counter (`nextCounter`) to assign unique, ordered IDs to sequence batches (via `Reorder`), preserving global order despite parallelism.

## Error Handling & Logging

- Panics on file-open failure (via `log.Panicf`).
- Logs start/end of reading per file using structured logging (`log.Printf`, `log.Println`).

## Resource Management

- Uses a barrier pattern: each reader goroutine calls `batchiter.Done()` upon completion.
- A finalizer goroutine waits for all readers (`WaitAndClose`) and logs termination.

## Design Intent

- Enables scalable, memory-efficient ingestion of large NGS datasets.
- Decouples *reading logic* (via `IBatchReader`) from orchestration — supporting pluggable formats.
- Prioritizes throughput and deterministic ordering over strict FIFO per-file semantics.

## Key Abstractions

| Type/Interface | Role |
|----------------|------|
| `IBatchReader` | Reader factory: `(filename, options...) → SequenceIterator` |
| `obiiter.IBioSequence` | Thread-safe batch iterator (push model) |
| `AtomicCounter` | Ensures globally unique, sequential batch IDs across goroutines |

