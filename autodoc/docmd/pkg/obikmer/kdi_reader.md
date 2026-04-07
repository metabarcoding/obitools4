# KDI Reader: Streaming Delta-Varint Decoding for k-mers

The `obikmer` package provides a high-performance, streaming reader for `.kdi` files—binary containers storing *sorted* k-mers (typically DNA substrings encoded as 64-bit integers). It supports both sequential and indexed access.

## Core Features

- **Streaming decoding**: K-mers are read incrementally using delta-varint compression to minimize I/O and memory footprint.
- **Delta encoding**: After the first absolute `uint64`, subsequent values are stored as *deltas* (difference from previous), encoded via custom `DecodeVarint`.
- **Magic & format validation**: A 4-byte magic header ensures file integrity; Little Endian `uint64` stores total count.
- **Sparse indexing**: When paired with a `.kdx` index, `SeekTo(target)` enables fast forward-only jumps to positions ≥ target k-mer.
- **Graceful fallback**: If `.kdx` is missing or invalid, the reader automatically degrades to sequential mode.

## Key API

- `NewKdiReader(path)` → opens `.kdi` for streaming (no index).
- `NewKdiIndexedReader(path)` → opens with optional `.kdx` for random access.
- `Next()` → returns `(nextKmer, true)` or `(0, false)` when exhausted.
- `SeekTo(target uint64) error` → jumps to first k-mer ≥ target using index (no backward seek).
- `Count()` / `Remaining()` → total and unread k-mers.
- `Close()` → releases file handle.

## Design Highlights

- Uses 64 KB buffer for efficient I/O.
- Index entries record `(kmer, byteOffset)` at fixed strides (e.g., every 1024 k-mers).
- `SeekTo` is idempotent and safe: no-op if target ≤ current position or index unavailable.
- Designed for large-scale genomic k-mer catalogs (e.g., from minimizers or de Bruijn graphs).
