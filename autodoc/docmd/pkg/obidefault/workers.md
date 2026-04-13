# Obidefault: Parallelism Configuration Module

This Go package (`obideault`) provides a centralized, configurable interface for managing parallel execution parameters—particularly useful in I/O- and CPU-bound workloads.

## Core Concepts

- **CPU-aware defaults**: Automatically detects available cores via `runtime.NumCPU()`.
- **Configurable workers per core**:
  - General: `_WorkerPerCore` (default `1.0`)
  - Read-specific: `_ReadWorkerPerCore` (`0.25`, i.e., ~1 reader per 4 cores)
  - Write-specific: `_WriteWorkerPerCore` (`0.25`)
- **Strict overrides**: Allow hardcoding worker counts via `SetStrictReadWorker()`/`Write...`, bypassing per-core scaling.

## Public API

| Function | Purpose |
|---------|--------|
| `ParallelWorkers()` | Total workers = `MaxCPU() × WorkerPerCore` |
| `Read/WriteParallelWorkers()` | Resolves to strict count if set, else per-core calculation (min 1) |
| `ParallelFilesRead()` | Files read in parallel: defaults to `ReadParallelWorkers()`, overridable |
| Getters (`MaxCPU`, `WorkerPerCore`, etc.) | Expose current settings safely |
| Setters (`Set*`) | Dynamically adjust behavior at runtime |

## Configuration Sources

- **Command-line flags**: e.g., `--max-cpu` or `-m`
- **Environment variable**: `OBIMAXCPU`

## Design Highlights

✅ Decouples resource discovery from policy  
✅ Supports both *proportional* (per-core) and *absolute* (strict) worker definitions  
✅ Ensures non-zero defaults for critical paths (`ReadParallelWorkers` ≥ 1)  

⚠️ **Note**: `WriteParallelWorkers()` contains a likely bug—returns `_StrictReadWorker` in the else branch instead of `StrictWriteWorker`.
