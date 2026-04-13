# KDI File Format and API

The `obikmer` package implements a compact, sorted k-mer storage format (`.kdi`) with delta compression for efficient disk persistence and retrieval.

## Core Features

- **Sorted k-mer serialization**: K-mers (as `uint64`) are written in ascending order.
- **Delta encoding**: Consecutive differences (deltas) between k-mers are stored using variable-length integers (`varint`), drastically reducing size for dense sequences.
- **Round-trip integrity**: Full write/read cycles preserve exact k-mer values and counts.

## File Structure

A `.kdi` file contains:
1. **Magic header** (4 bytes): Identifies the format.
2. **Count field** (8 bytes, `uint64`): Number of stored k-mers.
3. **First value** (8 bytes, `uint64`): Initial k-mer.
4. **Delta-encoded tail**: `(n−1)` deltas, each encoded as a `varint`.

## API

- **`NewKdiWriter(path string)`**: Creates a writer; `Write(v uint64)` appends k-mers.
- **`Writer.Count()`**: Returns the number of written items before closing.
- **`NewKdiReader(path string)`**: Opens a reader; `Next() (uint64, bool)` yields k-mers in order.
- **`Reader.Count()`**: Returns total stored count.

## Tests Validate

1. Basic round-trip with diverse values (including large `uint64`s).
2. Empty and single-k-mer files.
3. Exact file size for minimal cases (e.g., 20 bytes for one k-mer).
4. Delta compression efficiency on dense sequences (e.g., 10k even numbers → ~9,999 extra bytes).
5. Real-world usage: extracting canonical k-mers from DNA sequences, sorting/deduplicating, and persisting them.

The format is optimized for memory-mapped access or streaming traversal in bioinformatics pipelines.
