# KDI File Format and Writer

The `obikmer` package implements a compact, sorted sequence storage format for 64-bit k-mers using delta encoding and sparse indexing.

## Core Format (`.kdi`)

- **Magic header**: `KDI\x01` (`4 bytes`) identifies the file type.
- **Count field**: `uint64 LE`, total number of k-mers (patched at close).
- **First value**: `uint64 LE`, the initial k-mer stored as an absolute integer.
- **Deltas**: Subsequent values encoded via *delta-varint* (difference from previous k-mer), enabling high compression for sorted sequences.

## Writer (`KdiWriter`)

- **Strict ordering**: K-mers must be written in *strictly increasing order*.
- Efficient buffering via `bufio.Writer` (64 KB buffer).
- Internally tracks:
  - Current k-mer count,
  - Previous value (for delta computation),
  - Bytes written in data section.
- **Sparse indexing**: Every `defaultKdxStride` k-mers, an entry is recorded in memory for random access.

## Companion Index (`.kdx`)

- Written automatically on `Close()` if indexing entries exist.
- Stores `(kmer, file_offset)` pairs for fast seek-to-position lookups (e.g., binary search on k-mer range).
- Enables efficient random access without full file scan.

## Usage Pattern

```go
w, _ := obikmer.NewKdiWriter("data.kdi")
for _, kmer := range sortedKMers {
    w.Write(kmer)
}
w.Close()  // finalizes header, writes .kdx index
```

The format is optimized for memory-efficient storage and fast retrieval of sorted uint64 k-mers in genomic or sequence analysis pipelines.
