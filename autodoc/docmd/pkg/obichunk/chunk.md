# `ISequenceChunk` Function — Semantic Description

The `ISequenceChunk` function provides a unified interface for processing biological sequence data in chunks, supporting two execution modes: **in-memory** and **on-disk**, depending on resource constraints or performance needs.

- It accepts an iterator over biological sequences (`obiiter.IBioSequence`) and a sequence classifier (`obiseq.BioSequenceClassifier`), used to annotate or categorize sequences.
- A boolean flag `onMemory` determines whether processing occurs in RAM (`ISequenceChunkOnMemory`) or on disk (`ISequenceChunkOnDisk`), enabling scalability for large datasets.
- Optional parameters allow fine-tuning:
  - `dereplicate`: enables deduplication of identical sequences.
  - `na`: specifies how missing or ambiguous values are handled (e.g., `"?"`, `"N"`, etc.).
  - `statsOn`: configures what metadata (e.g., description fields) are tracked for statistics.
  - `uniqueClassifier`: an optional secondary classifier used to assign unique identifiers or labels.

The function abstracts the underlying implementation, ensuring consistent behavior regardless of storage strategy. It returns an iterator over processed sequences (`obiiter.IBioSequence`) or an error, supporting streaming workflows and compatibility with downstream pipeline stages.

This design promotes flexibility, memory efficiency, and modularity in high-throughput sequence analysis pipelines (e.g., metabarcoding).
