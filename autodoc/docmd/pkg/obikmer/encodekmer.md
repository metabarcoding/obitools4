# Semantic Description of `obikmer` Package

The `obikmer` package provides high-performance, zero-allocation utilities for **k-mer manipulation** in DNA sequences (A/C/G/T/U), targeting bioinformatics applications like genome indexing, assembly, and error correction.

## Core Encoding & Decoding

- **`EncodeKmer`, `DecodeKmer`**: Convert between DNA sequences and compact 62-bit uint64 representations (2 bits/base), preserving top 2 bits for optional error markers.
- **`EncodeCanonicalKmer`, `CanonicalKmer`**: Encode or normalize k-mers to their *biological canonical form* — the lexicographically smaller of a k-mer and its reverse complement.

## Iterators (Memory-Efficient Streaming)

- **`IterKmers`, `IterCanonicalKmers`**: Stream all overlapping k-mers from a sequence without allocating intermediate slices — ideal for large-scale processing (e.g., inserting into Roaring Bitmaps).
- **`IterCanonicalKmersWithErrors`**: Same as above, but detects ambiguous bases (N/R/Y/W/S/K/M/B/D/H/V) and encodes their count in the top 2 bits (error code: 0–3). Only valid for **odd k ≤ 31**.

## Error Handling & Markers

- `SetKmerError`, `GetKmerError`, and `ClearKmerError` manipulate the top 2 bits of a uint64 to store error metadata (e.g., ambiguous base count), enabling downstream filtering or correction.

## Reverse Complement & Circular Normalization

- **`ReverseComplement`, `CanonicalKmer`**: Compute biological reverse complement and canonical form.
- **`NormalizeCircular`, `EncodeCircularCanonicalKmer`**: Compute *circular canonical form* — the lexicographically smallest rotation (used for low-complexity masking).
- Distinction: `CanonicalKmer` uses **reverse complement**, while `NormalizeCircular` uses **rotation**.

## Counting & Math Utilities

- **`CanonicalCircularKmerCount`, `necklaceCount`, etc.**: Compute exact counts of unique circular k-mer equivalence classes using **Moreau’s necklace formula**, with Euler's totient function and divisor enumeration.

## Performance & Safety

- All functions avoid heap allocations where possible (reusing buffers).
- Panics on invalid `k` or length mismatches for correctness.
- Supports case-insensitive input (A/a, T/t…), and ambiguous bases via `__single_base_code_err__`.

## Use Cases

- K-mer counting in assemblers (e.g., with Bloom filters or bitmaps)
- Error-aware k-mer filtering in sequencing pipelines
- Low-complexity region detection via circular entropy normalization
