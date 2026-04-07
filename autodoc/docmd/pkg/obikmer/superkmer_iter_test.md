# Semantic Description of `obikmer` Package Functionalities

The `obikmer` package provides tools for **super k-mer extraction and minimizer-based sequence analysis** in bioinformatics.

## Core Concepts

A **super k-mer** is a maximal contiguous subsequence of DNA where *all* embedded *k*-mers share the **same minimizer**—a compact representative (typically lexicographically minimal) of *m*-mers, considering both forward and reverse-complement strands.

## Key Functions & Features

- **`IterSuperKmers(seq, k, m)`**:  
  An iterator over all super *k*-mers in input sequence `seq`, parameterized by:
  - `k`: length of embedded *k*-mers,
  - `m`: size of minimizer window (`m ≤ k`).  
  Yields structured objects with:
  - `Sequence`: the super *k*-mer substring,
  - `Start`/`End`: genomic coordinates (0-based half-open),
  - `Minimizer`: canonical hash of the shared minimizer.

- **`ExtractSuperKmers(...)`**:  
  Synchronous counterpart returning a slice of all super *k*-mers.

## Verified Properties (via Tests)

1. **Boundary correctness**: Extracted subsequences match `seq[start:end]`.
2. **Consistency between iterator and slice versions**: Both APIs produce identical results.
3. **Bijection property**:
   - Each unique super *k*-mer sequence maps to exactly one minimizer.
   - All embedded *k*-mers within a super *k-mer* share the same minimizer.

## Implementation Notes

- Minimizers are computed canonically (min of forward and reverse-complement encodings).
- Uses base encoding via `__single_base_code__` (assumed helper mapping A/C/G/T → 0/1/2/3).
- Tests cover simple, homopolymer-rich, and complex genomic patterns.

## Design Rationale

Super *k*-mers enable efficient compression, indexing (e.g., in minimizer spaces), and alignment-free comparisons—crucial for scalable genomic analysis.
