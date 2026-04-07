# Semantic Description of `obialign` Package

The `obialign` package provides high-performance, memory-efficient tools for **pairwise alignment of paired-end biological sequences**, optimized specifically for Next-Generation Sequencing (NGS) data.

## Core Functionalities

### 1. **Memory Arena Management**
- `PEAlignArena` is a reusable memory buffer to avoid repeated allocations during multiple alignments.
- Preallocates matrices (`scoreMatrix`, `pathMatrix`), alignment buffers, and auxiliary structures based on expected max sequence lengths.

### 2. **Dynamic Programming Alignment Functions**
Implements three specialized global alignment variants using Needleman–Wunsch with affine gap penalties (scaled per mismatch):

- **`PELeftAlign`**: Free gaps at the *start* of `seqB` and end of `seqA`. Ideal for aligning overlapping reads where the first read starts before or within the second.
- **`PERightAlign`**: Free gaps at start of `seqA` and end of `seqB`. Suited when the second read extends beyond the first.
- **`PECenterAlign`**: Free gaps at both ends of *both* sequences; requires `seqA ≥ seqB`. Designed for full overlap scenarios (e.g., merging paired-end reads).

All use column-major matrix storage and efficient index arithmetic via helper functions `_GetMatrix`, `_SetMatrices`, etc.

### 3. **Scoring & Quality Integration**
- Pairwise base/quality scores computed by `_PairingScorePeAlign`, combining:
  - Nucleotide compatibility (via precomputed `_NucPartMatch`)
  - Phred quality scores (`_NucScorePartMatchMatch`, `_NucScorePartMatchMismatch`)
  - A user-defined `scale` factor to modulate mismatch penalties.

### 4. **Fast Heuristic Pre-Alignment**
The main `PEAlign` function integrates a kmer-based fast pre-screening:
- Uses 4-mer indexing (`obikmer.Index4mer`) and shift estimation via `FastShiftFourMer`.
- If overlap is significant (`fastCount + 3 < over`), performs localized DP only on the predicted overlapping region (using `PELeftAlign` or `PERightAlign`) to save time.
- Otherwise, computes full alignment over entire sequences (both left and right variants), selecting the best score.

### 5. **Backtracking & Path Output**
- `_Backtracking` reconstructs the optimal alignment path from `pathMatrix`.
- Paths encoded as alternating `(offset, length)` pairs for aligned segments (diagonal = 0), with gaps encoded as `-1`/`+1`.

### Use Case
Designed for **paired-end read merging**, overlap detection, and consensus building in metagenomic pipelines (e.g., OBITOOLS4 ecosystem). Efficient, scalable for large batch processing via arena reuse.
