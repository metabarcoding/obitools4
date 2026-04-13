# Semantic Description of `obijoin` Package

The `obijoin` package implements a flexible sequence-join mechanism for biological sequencing data within the OBITools4 framework. It supports efficient lookup and merging of metadata/sequences based on user-defined keys.

- **`IndexedSequenceSlice`**: A data structure combining a slice of biological sequences (`BioSequenceSlice`) with precomputed indices for fast filtering. Each index maps attribute values (e.g., sample IDs, barcodes) to sets of matching sequence indices.

- **`Get(keys...)` method**: Performs multi-key intersection queries across pre-built indexes to retrieve sequences matching *all* specified attribute values — enabling complex filtering (e.g., `sample="S1" AND barcode="ATGC"`).

- **`BuildIndexedSequenceSlice()`**: Constructs the index structure in linear time by scanning sequences and grouping them per attribute key. Supports arbitrary string attributes retrieved via `GetStringAttribute()`.

- **`MakeJoinWorker()`**: Returns a functional worker (`SeqWorker`) that, for each input sequence:
  - Extracts join keys (e.g., `sample`, `barcode`) from its annotations.
  - Uses the index to find matching partner sequences (`join_with`).
  - Produces one output sequence per match, copying the original and enriching it with annotations from partners.
  - Optionally updates ID, sequence, or quality scores based on partner data.

- **`CLIJoinSequences()`**: Top-level CLI entry point: reads a reference dataset (via `--join-with`), builds the index, and applies join logic using command-line flags (`--by`, `--update-id`, etc.). Integrates with OBITools4’s streaming iterator model.

- **Use Cases**: Merging paired-end reads, annotating amplicons with sample metadata, or combining reference databases — all via declarative key-based joins.

- **Efficiency**: Indexing avoids repeated scanning; intersection logic is optimized via `obiutils.Set[int]`.
- **Extensibility**: Works with any attribute supported by the sequence annotation system.
