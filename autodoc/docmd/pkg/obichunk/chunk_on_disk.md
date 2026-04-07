# `obichunk` Package: On-Disk Chunking and Dereplication of Biosequences

The `obichunk` package provides functionality to efficiently process large sets of biological sequences by splitting them into manageable, disk-based chunks. Its core feature is the `ISequenceChunkOnDisk` function, which takes a sequence iterator and distributes sequences into temporary files using a classifier. Each file corresponds to one *batch* (e.g., `chunk_*.fastx`), enabling scalable, parallel-friendly workflows.

Key capabilities include:

- **Temporary Directory Management**: Automatically creates and cleans up a system temp directory (`obiseq_chunks_*`) for intermediate storage.
- **File Discovery**: Recursively finds all `.fastx` files generated during chunking via `find`.
- **Asynchronous Streaming**: Returns an iterator (`obiiter.IBioSequence`) that yields batches asynchronously, decoupling chunk creation from consumption.
- **Optional Dereplication**: When enabled (`dereplicate = true`), sequences are deduplicated *per batch* using a composite key (sequence + classification categories). Merged duplicates retain aggregated statistics.
- **Logging & Monitoring**: Logs total batch count and per-batch processing start events for transparency.

Internally, `ISequenceChunkOnDisk` uses:
- `obiiter.MakeIBioSequence()` to build the output iterator,
- `obiformats.WriterDispatcher` for parallel writing of distributed sequences into chunk files,
- and a second goroutine to read, optionally dereplicate (via `BioSequenceClassifier`), and push batches back into the output iterator.

Designed for memory efficiency, it avoids loading all sequences in RAM by streaming and chunking on-disk—ideal for large-scale NGS data preprocessing.
