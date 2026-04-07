# Obikmer: Efficient K-mer Encoding and Manipulation in Go

This package provides high-performance utilities for DNA sequence analysis using *k*-mers—contiguous substrings of length `k`. It supports encoding, canonicalization (forward/reverse-complement normalization), minimizer-based super-*k*-mer extraction, and error tagging—all optimized for 64-bit integer arithmetic.

## Core Functionalities

### K-mer Encoding (`EncodeKmers`, `IterKmers`)
Encodes DNA sequences (A/C/G/T/U, case-insensitive) into `uint64` using 2 bits per nucleotide (A=00, C=01, G=10, T/U=11). Supports sliding-window extraction and streaming via an iterator. Handles sequences up to 31-mers (62 bits), with validation for invalid `k` values.

### Reverse Complement (`ReverseComplement`)
Computes the reverse complement of a *k*-mer in constant time using bit manipulation. Preserves error metadata (see below) and satisfies involution: `RC(RC(x)) = x`.

### Canonical K-mers (`CanonicalKmer`, `EncodeCanonicalKmers`)
Returns the lexicographically smaller of a *k*-mer and its reverse complement—enabling strand-agnostic analysis. Supports both single-kmer normalization (`CanonicalKmer`) and full-sequence canonical encoding.

### Super *k*-mers Extraction (`ExtractSuperKmers`)
Groups overlapping *k*-mers sharing the same minimizer (minimal *m*-mer in sliding window) into contiguous regions ("super *k*-mers"). Output includes start/end positions and minimizer values, all canonicalized.

### Error Marking (`SetKmerError`, `GetKmerError`, etc.)
Uses the top 2 bits of a `uint64` to tag error states (e.g., sequencing errors), leaving 62 bits for sequence data. Error operations preserve the underlying *k*-mer and work seamlessly with canonicalization/RC.

## Key Features

- **Memory Efficiency**: Reusable buffers via optional `*[]uint64` or `*[]SuperKmer` parameters.
- **Edge Case Handling**: Gracefully handles empty sequences, `k > len(seq)`, invalid parameters (`m ≥ k`), and max-length constraints.
- **Performance**: Optimized for speed—benchmarks included for all major functions (e.g., `BenchmarkEncodeKmers`, `BenchmarkExtractSuperKmers`).
- **Comprehensive Testing**: Covers basic cases, boundary conditions (e.g., 31-mers), symmetry properties (canonical/RC invariance), and error resilience.

## Use Cases

- Genome assembly &DBG construction  
- Minimizer-based sketching (e.g., *Mash*, *Sourmash*)  
- Error-aware k-mer counting & filtering  
- Strand-unbiased sequence comparison  

All functions operate on `[]byte` DNA sequences and return canonicalized, efficient representations suitable for hashing or indexing.
