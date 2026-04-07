# Semantic Description of `obiseq` Package Functionalities

The `obiseq` package provides composable, higher-order worker functions for processing biological sequence data in Go. It defines three core functional types:

- `SeqAnnotator`: In-place annotation of a single sequence (e.g., adding metadata).
- `SeqWorker`: Processes one sequence and returns zero or more output sequences (1→N transformation).
- `SeqSliceWorker`: Processes a slice of sequences and returns another slice (bulk pipeline stage).

Key utilities include:

- **`NilSeqWorker`**: Identity worker—returns the input sequence unchanged.
- **`AnnotatorToSeqWorker`**: Converts an in-place annotator into a `SeqWorker`, preserving compatibility with pipeline interfaces.
- **`SeqToSliceWorker`**: Lifts a `SeqWorker` to operate on slices, with configurable error handling (`breakOnError`). Supports dynamic slice growth and logging via `obilog`.
- **`SeqToSliceFilterOnWorker`**: Filters sequences in a slice using a `SequencePredicate`, preserving order and avoiding unnecessary allocations.
- **`SeqToSliceConditionalWorker`**: Applies a `SeqWorker` only to sequences satisfying a predicate; others pass through unchanged.
- **`.ChainWorkers()`**: Method on `SeqWorker` to compose two workers sequentially (pipeline chaining), enabling modular, reusable workflows.

All functions emphasize safety: errors are either propagated (`breakOnError = true`) or logged with warnings, ensuring robustness in large-scale sequence processing pipelines.
