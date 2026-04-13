# SuperKmer and Minimizer-Based Sliding Window Analysis

This Go package provides functionality for extracting *super k-mers* from DNA sequences using a minimizer-based sliding window approach.

## Core Concepts

- **K-mers**: Substrings of length `k` from a DNA sequence.
- **Minimizer**: The lexicographically smallest canonical *m*-mer (substring of length `m`) among all `(k − m + 1)` overlapping *m*-mers in a given k-mer.
- **Super K-mer**: A maximal contiguous subsequence where *every* consecutive k-mer shares the **same minimizer**.

## Data Structures

### `SuperKmer`
Represents a maximal region with uniform minimizer:
- `Minimizer`: Canonical 64-bit hash of the shared m-mer.
- `Start`, `End`: Slice-style bounds (0-indexed, exclusive end).
- `Sequence`: Raw byte slice of the DNA subsequence.

### `dequeItem`
Used internally to maintain a monotone deque:
- `position`: Index of the m-mer in the sequence.
- `canonical`: Canonical hash value (e.g., lexicographically smallest of forward/reverse-complement).

## Main Function

### `ExtractSuperKmers(seq, k, m, buffer)`
- Extracts all maximal super k-mers from `seq`.
- Parameters validated:  
  - `1 ≤ m < k`,  
  - `2 ≤ k ≤ 31`,  
  - sequence length ≥ `k`.
- Uses an efficient **O(n)** time algorithm via internal iteration.
- Supports optional preallocation (`buffer`) to reduce memory allocations.

## Algorithm Highlights

- Maintains a sliding window of size `k − m + 1` over *m*-mers.
- Tracks the current minimizer using a monotone deque for O(1) updates per step.
- Detects *minimizer transitions* to delimit super k-mer boundaries.

## Complexity

| Aspect        | Bound                         |
|---------------|-------------------------------|
| Time          | **O(n)** (linear in sequence length) |
| Space         | **O(k − m + 1)** for deque + output size |

Useful in genome compression, read clustering, and minimizer-based alignment acceleration.
