# `obiutils` Package: Pipe Synchronization Utilities

This Go package provides lightweight synchronization primitives for managing concurrent pipeline execution, particularly useful in CLI or batch-processing applications.

## Core Components

- **`globalLocker`:** A `sync.WaitGroup` tracking active pipeline goroutines.
- **`globalLockerCounter`:** An integer counter for logging/debugging the number of active pipes.

## Public Functions

### `RegisterAPipe()`
- Increments both the WaitGroup and counter.
- Logs current count at debug level (`log.Debugln`).
- Typically called when starting a new pipeline stage or goroutine.

### `UnregisterPipe()`
- Decrements the WaitGroup and counter.
- Logs updated count at debug level.
- Should be invoked when a pipeline finishes (e.g., `defer UnregisterPipe()`).

### `WaitForLastPipe()`
- Blocks until all registered pipes complete (`globalLocker.Wait()`).
- Intended to be called at the end of `main()`, ensuring graceful shutdown.

## Semantic Use Case

Enables safe, concurrent execution of multiple independent pipelines (e.g., data processing stages), ensuring the program waits for all to finish before exiting — without explicit channel or mutex management.

## Design Notes

- **Thread-safe** via `sync.WaitGroup`.
- **Minimalist**: No error handling; assumes correct usage.
- **Logging-focused** for observability in development/debug builds.

> ⚠️ Not production-ready without additional safeguards (e.g., panic recovery, timeout support).
