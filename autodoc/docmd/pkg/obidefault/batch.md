# ObiDefault Package: Batch Configuration Module

This Go module provides centralized configuration for sequence batching in Obitools, supporting both **count-based** and **memory-aware** batch processing.

## Core Features

- `_BatchSize` / `SetBatchSize()`  
  Defines and configures the *minimum* number of sequences per batch (default: `1`).  
  Used internally as `minSeqs` in `RebatchBySize`.

- `_BatchSizeMax()` / `SetBatchSizeMax()`  
  Sets the *maximum* sequences per batch (default: `2000`). Batches are flushed upon reaching this limit, regardless of memory.

- **CLI & Environment Integration**  
  Batch size is determined by `--batch-size` CLI flag and/or the `OBIBATCHSIZE` environment variable (via parsing logic not shown here but implied by comments).

- `_BatchMem()` / `SetBatchMem(n int)`  
  Configures the *maximum memory per batch* (default: `128 MB`). A value of `0` disables memory-based batching, falling back to pure count-based logic.

- `_BatchMemStr()`  
  Stores the *raw CLI string* passed to `--batch-mem` (e.g., `"256M"`, `"1G"`), enabling human-readable input parsing elsewhere.

## Utility Functions

- `BatchSizePtr()`, `BatchMemPtr()`  
  Expose pointers to internal variables for direct modification or inter-process sharing.

- `BatchSizeMaxPtr()`, `BatchMemStrPtr()`  
  Provide read/write access to max-size and raw memory string values.

## Design Intent

- Separates **configuration** (defaults, CLI/env parsing) from **processing logic**, enabling modular and testable batch handling.
- Supports both scalable, large-scale processing (via count limits) and memory-constrained environments (via soft RAM caps).
