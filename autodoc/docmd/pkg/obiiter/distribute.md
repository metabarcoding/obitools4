# `IDistribute`: Semantic Description of Biosequence Distribution Functionality

The `IDistribute` type implements a thread-safe mechanism for distributing biosequences into classified, batched outputs.

- **Core Purpose**: Enables concurrent processing of sequences by routing them to dedicated output channels based on classification keys.

- **Key Fields**:
  - `outputs`: A map from integer class codes to output streams (`IBioSequence`).
  - `news`: An unbuffered channel emitting class codes when new output streams are created.
  - `classifier`: A pointer to a sequence classifier used to assign sequences to keys during distribution.

- **Thread Safety**: All access to shared state (`outputs`, `slices`) is synchronized via a mutex.

- **Batching Strategy**:
  - Sequences are accumulated per class key until either `BatchSizeMax()` sequences or `BatchMem()` bytes (per key) are reached.
  - Batches are flushed automatically and on finalization.

- **Asynchronous Processing**:
  - The `Distribute()` method launches a goroutine that consumes the input iterator, classifies each sequence, and feeds batches to per-key outputs.
  - Outputs are closed only after all sequences have been processed.

- **Notifications**:
  - The `News()` channel allows consumers to be notified of newly created output streams (i.e., when a new class key appears).

- **Error Handling**:
  - `Outputs(key)` returns an error if the requested key has no associated output.

- **Integration**:
  - Leverages `obidefault.BatchSizeMax()` and `BatchMem()` for configurable batch limits.
  - Uses `SortBatches()` on the input iterator to ensure ordered processing.

In summary, `IDistribute` provides a scalable, concurrent pipeline for classifying and batching biosequences based on user-defined classification logic.
