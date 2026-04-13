# Semantic Description of `obiiter` Package Functionalities

This Go package (`obiiter`) provides utilities for applying functional transformations to biological sequence iterators, supporting parallel execution and modular piping.

- **`MakeIWorker(worker, breakOnError bool, sizes ...int)`**:  
  Applies a `SeqWorker` (sequence-to-sequence transformation) to each sequence in the iterator. Supports configurable parallelism (`nworkers`) and optional channel buffering via `sizes`. Uses internal conversion to slice-based workers.

- **`MakeIConditionalWorker(predicate, worker, breakOnError bool, sizes ...int)`**:  
  Applies a `SeqWorker` only to sequences satisfying a given boolean `predicate`. Enables conditional, parallelized processing while preserving iterator semantics.

- **`MakeISliceWorker(worker, breakOnError bool, sizes ...int)`**:  
  Core method applying a `SeqSliceWorker` (batch-level transformation) across slices of sequences. Implements multi-goroutine parallelism using `nworkers`. Handles errors optionally via fatal logging (`breakOnError`). Preserves paired-end metadata.

- **`WorkerPipe(worker, breakOnError bool, sizes ...int)`**:  
  Returns a `Pipeable` closure wrapping `MakeIWorker`, enabling composition in pipeline chains (e.g., for CLI or DSL-style workflows).

- **`SliceWorkerPipe(worker, breakOnError bool, sizes ...int)`**:  
  Similar to `WorkerPipe`, but for slice-level workers (`SeqSliceWorker`). Facilitates modular, reusable pipeline stages.

All methods support optional size arguments to override default parallelism (from `obidefault`). Internally, they rely on Go concurrency primitives (`go`, channels) and structured batch processing via `IBioSequence` interface.
