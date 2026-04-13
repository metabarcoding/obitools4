# SKM File Format Specification

This Go package implements a binary format for storing *super-kmers*—compact representations of DNA sequences used in bioinformatics. The tests validate reading/writing, padding behavior, and file size correctness.

## Core Functionalities

- **SuperKmer Structure**: Each super-kmer stores a DNA sequence (as bytes), likely padded to 4-base boundaries for efficient storage.
- **SkmWriter**: Serializes super-kmers into a binary file. Each entry writes:
  - A 2-byte little-endian length (number of bases),
  - Then `ceil(length/4)` bytes encoding nucleotides in 2 bits each (A=0, C=1, G=2, T=3).
- **SkmReader**: Parses the binary format back into memory. Returns `(SuperKmer, bool)` via `Next()`, with EOF signaled by `ok = false`.
- **Case Handling**: Writes preserve original case; reads normalize to lowercase (via `| 0x20` in tests), ensuring robust comparison.

## Test Coverage

- **Round-trip integrity**: Verifies exact sequence recovery after write/read.
- **Empty file handling**: Confirms reader returns `ok = false` immediately on empty files.
- **Variable-length padding**: Validates correct encoding/decoding for sequences of length 1–5.
- **Size validation**: Confirms file size = `2 + ceil(L/4)` bytes for a sequence of length *L*.

## Use Case

Efficient, lossless storage and retrieval of super-kmers for downstream genomic analysis (e.g., assembly or alignment acceleration).
