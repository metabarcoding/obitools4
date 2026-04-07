# JSON Output Module for Biological Sequences (`obiformats`)

This Go package provides utilities to serialize biological sequence data (from `obiseq`) into structured JSON format, supporting batch processing and parallel I/O.

- **`JSONRecord(sequence)`**: Converts a single `BioSequence` into an indented JSON object containing:
  - `"id"`: Sequence identifier.
  - `"sequence"` (optional): Nucleotide/protein sequence string if present.
  - `"qualities"` (optional): Quality scores as a string if available.
  - `"annotations"` (optional): Metadata annotations map.

- **`FormatJSONBatch(batch)`**: Formats a batch of sequences as JSON array elements, returning a `*bytes.Buffer`. Handles comma separation and indentation.

- **`WriteJSON(iterator, file)`**: Writes a stream of sequences to an `io.Writer`, supporting:
  - Parallel workers (configurable via options).
  - Automatic compression (`gzip`/`bgzip`) if enabled.
  - Proper JSON array wrapping: `[`, chunked batches, and final `]`.
  - Atomic ordering to preserve sequence integrity across parallel writes.

- **`WriteJSONToStdout()` / `WriteJSONToFile()`**: Convenience wrappers:
  - Outputs to stdout or a file (with append/truncate control).
  - Supports paired-end data: writes both forward and reverse reads to separate files when configured.

- **Internal helpers**:
  - `_UnescapeUnicodeCharactersInJSON()`: Fixes double-escaped Unicode in JSON output (e.g., `\\u00E9` → `\u00E9`).
  - Uses chunked concurrency with `FileChunk`, ordered by batch number to ensure valid JSON structure.

Designed for high-throughput NGS data pipelines, it ensures correctness and performance while integrating with `obitools4`'s iterator-based processing model.
