# Memory-Limited Biosequence Iterator

This Go function extends an `IBioSequence` iterator with memory-aware throttling to prevent excessive heap allocation during data processing.

## Core Functionality

- **`LimitMemory(fraction float64)`**  
  Returns a new iterator that respects an upper bound on heap usage relative to total system memory.

- **Memory Monitoring**  
  Uses `runtime.ReadMemStats()` and `github.com/pbnjay/memory.TotalMemory()` to compute the current heap fraction (`Alloc / TotalMemory`) dynamically.

- **Backpressure Mechanism**  
  While the memory fraction exceeds `fraction`, the producer goroutine yields control (`runtime.Gosched()`) until sufficient memory becomes available.

- **Logging**  
  Warns via `obilog.Warnf` when:
  - Memory pressure persists (every ~1000 yields),
  - Or wait duration becomes unusually long (>10,000 yielding cycles).

- **Concurrency Model**  
  - A producer goroutine consumes from the original iterator and pushes items to `newIter`, pausing as needed.
  - A dedicated consumer goroutine calls `WaitAndClose()` to ensure graceful termination and resource cleanup.

## Semantic Behavior

- **Non-blocking consumer**: Downstream consumers are not stalled; they read from an internal buffered channel (`newIter`).
- **Adaptive rate control**: The iterator automatically slows down when memory pressure rises, avoiding OOM conditions.
- **Predictable resource use**: Ensures heap usage stays below the specified `fraction` (e.g., 0.5 → ≤ 50% of total RAM).
