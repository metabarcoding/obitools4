# KDX Index Format and Functionality

The `obikmer` package provides a sparse indexing mechanism for `.kdi` files (likely storing sorted k-mers with delta encoding). The **`.kdx` file** serves as a fast lookup table to accelerate k-mer searches.

## Core Concepts

- **Magic bytes**: `KDX\x01` validates file integrity.
- **Stride-based sparsity**: One index entry every *N* k-mers (default: 4096), balancing memory vs. search speed.
- **Entry structure**: Each entry stores:
  - `kmer`: the k-mer value at that index position.
  - `offset`: absolute byte offset in the corresponding `.kdi` file.

## Key Operations

- **Loading**: `LoadKdxIndex()` reads and validates a `.kdx` file; returns `(nil, nil)` if missing (graceful degradation).
- **Searching**: `FindOffset(target uint64)` performs binary search over index entries to find the *best jump point*:
  - Returns `offset`, `skipCount` (k-mer count already passed), and a boolean success flag.
  - Enables efficient seeking: after `offset`, only up to *stride* k-mers need linear scanning.
- **Writing**: `WriteKdxIndex()` serializes an in-memory index to disk (for building indexes).
- **Helper**: `KdxPathForKdi()` derives the `.kdx` path from a given `.kdi` file.

## Performance

- Search complexity: **O(log M)** for the binary search (where *M* = #index entries), plus ≤ stride linear steps.
- Memory footprint: Linear in index size (16 bytes per entry), highly scalable for large k-mer sets.

## Design Philosophy

Minimalist, binary-safe format with explicit endianness (little-endian), no external dependencies beyond `encoding/binary`, and robust error handling.
