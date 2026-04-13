# Semantic Description of `obiformats.WriterDispatcher`

The package `obiformats` provides utilities for writing biosequences (e.g., DNA/RNA/protein reads) to files in a structured, parallelized manner. Its core component is the `WriterDispatcher` function.

- **Purpose**: Enables concurrent, classifier-guided writing of biosequence batches to multiple output files based on dynamic dispatching logic.
- **Input**: Takes a prototype filename template (`prototypename`), an `IDistribute` dispatcher (which partitions and routes sequences by classification keys), a formatting/writing function (`formater` of type `SequenceBatchWriterToFile`), and optional configuration.
- **Concurrency**: Launches one goroutine per classification category (via `dispatcher.News()`), ensuring scalable parallel writes.
- **Classification Handling**: Supports simple and composite keys (e.g., dual annotations like sample + region), parsing JSON-encoded classifier values when needed.
- **File Naming & Organization**: Substitutes keys into the prototype name, appends `.gz` if compression is enabled, and creates subdirectories (e.g., for sample groups) as required.
- **Error Handling**: Uses `log.Fatalf` to abort on unrecoverable errors (e.g., failed key parsing, directory creation issues).
- **Resource Management**: Ensures all goroutines complete before returning via `sync.WaitGroup`.
- **Extensibility**: The generic `SequenceBatchWriterToFile` type allows plugging in different output formats (e.g., FASTA, JSON) without modifying the dispatcher logic.

In summary: `WriterDispatcher` is a high-level orchestrator for parallel, classifier-based batch writing of biological sequences to organized file outputs.
