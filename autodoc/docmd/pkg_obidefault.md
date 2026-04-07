# Obidefault Package: Centralized Configuration Module

The `obidefault` package provides a unified, runtime-configurable interface for core application-level settings in the Obitools ecosystem. It centralizes global state related to batching, compression, verbosity, progress reporting, quality handling, taxonomy resolution, and parallelism—enabling consistent behavior across modules without parameter passing or recompilation.

## Batch Configuration

Controls sequence batching for efficient processing:
- `SetBatchSize(n)`, `_BatchSize()` → Minimum sequences per batch (default: 1).
- `SetBatchSizeMax(n)`, `_BatchSizeMax()` → Hard upper limit on batch size (default: 2000).
- `SetBatchMem(n)`, `_BatchMem()` → Memory cap per batch in bytes (default: 128 MB); `0` disables memory-based batching.
- `_BatchMemStr()` stores the raw CLI string (e.g., `"256M"`) for parsing.
- Supports configuration via `--batch-size`/`OBIBATCHSIZE`, and `--batch-mem`.

## Output Compression

Toggles compression of output streams:
- `SetCompressOutput(bool)`, `CompressOutput()` → Enable/disable compression globally.
- Pointer access via `CompressOutputPtr()` for dynamic binding.

## Warning Verbosity

Suppresses warning messages when enabled:
- `SetSilentWarning(bool)`, `SilentWarning()` → Control warning output.
- Pointer access via `SilentWarningPtr()`. When true, all warnings should be suppressed (implementation-dependent).

## Progress Bar Visibility

Enables/disables progress bar rendering:
- `SetNoProgressBar(bool)`, `NoProgressBar()` → Disable/enable bars (default: enabled).
- `ProgressBar()` returns the inverse of `NoProgressBar()`.
- Pointer access via `NoProgressBarPtr()`.

## Quality Score Handling

Configures FASTQ quality score parsing and encoding:
- `SetReadQualitiesShift(byte)`, `ReadQualitiesShift()` → Input offset (default: 33, Phred+33).
- `SetWriteQualitiesShift(byte)`, `WriteQualitiesShift()` → Output offset (default: 33).
- `SetReadQualities(bool)`, `ReadQualities()` → Enable/disable quality parsing (default: true).
- Enables format conversion and performance optimization.

## Taxonomy Configuration

Controls taxonomic identifier handling in OBIDMS workflows:
- `SetSelectedTaxonomy(string)`, `UseRawTaxids()`, etc. → Select taxonomy (e.g., `"NCBI"`), toggle raw/normalized IDs, alternative names.
- `SetFailOnTaxonomy(bool)`, `SetUpdateTaxid(bool)` → Control error behavior and auto-updates.
- Provides getters, setters, and pointer accessors for initialization-time configuration.

## Parallelism Control

Manages worker counts across read/write/ general operations:
- `SetWorkerPerCore(float64)`, `_ReadWorkerPerCore`/`_WriteWorkerPerCore` → Scaling factors (default: 1.0 / 0.25).
- `SetStrictReadWorker(n)`, `_MaxCPU` → Override with absolute worker counts.
- Functions: `ParallelWorkers()`, `Read/WriteParallelWorkers()` → Compute effective worker counts.
- Configurable via CLI flags (`--max-cpu`, `-m`) and `OBIMAXCPU` environment variable.

> **Design Note**: All settings are *not* thread-safe; intended for use during initialization. Public API exposes only getters/setters/pointers—no internal mutation beyond controlled access.
