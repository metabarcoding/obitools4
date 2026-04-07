# `obialign` Package: Semantic Overview

The `obialign` package delivers high-performance, memory-efficient utilities for biological sequence alignment within the OBITools4 ecosystem. It targets amplicon and metagenomic data processing, emphasizing speed, numerical stability, and scalability.

---

## Core Functionalities

### 1. **Sequence Encoding & Decoding**
- `Encode4bits`: Converts IUPAC nucleotides (including ambiguous codes like R, Y, N) into compact 4-bit representations.  
- Supports bitwise operations for rapid comparison (e.g., via `_FourBitsBaseCode`).  
- Handles gaps (`.`/`-`) and invalid characters as `0b0000`.

### 2. **Alignment Scoring & Probability Models**
- `_MatchRatio`, `_NucPartMatch`: Compute match likelihoods using bitwise overlap of encoded bases.  
- Log-space helpers (`_Logaddexp`, `_Logdiffexp`) ensure numerical stability in probabilistic scoring.  
- Quality-aware scores via precomputed matrices (`_NucScorePartMatch{Match,Mismatch}`), incorporating Phred scores and base composition priors.

### 3. **Dynamic Programming (DP) Backtracking**
- `_Backtracking`: Reconstructs optimal alignment paths from precomputed matrices.  
  - Encodes diagonal runs and gap segments as alternating `(offset, length)` pairs.
- Optimized for batch reuse of path buffers and minimal allocations.

### 4. **Longest Common Subsequence (LCS) with Error Tolerance**
- `FastLCSEGFScore`, `FastLCSScore`: Compute LCS under bounded error (`maxError`) and optional end-gap-free mode.  
  - Uses diagonal banding for efficiency.
- Designed for rapid similarity filtering (e.g., UMI/OTU clustering).

### 5. **Single-Edit Distance Detection**
- `D1Or0`: Determines if two sequences are identical or differ by exactly one edit (substitution/indel).  
  - Early termination on length mismatch or multiple divergences.  
  - Critical for error correction and dereplication.

### 6. **Local Pattern Matching**
- `LocatePattern`: Finds optimal approximate match of a short query (e.g., primer) in longer sequence.  
  - End-gap-free alignment, using DP with mismatch/gap penalty `-1`.  
  - Returns start/end positions and error count.

### 7. **Paired-End Read Alignment**
- `PEAlign`, `_FillMatrixPeLeftAlign`, etc.: Global alignment of paired-end reads with affine gap penalties.  
  - Supports three modes: `PELeftAlign`, `PERightAlign`, and `PECenterAlign` (for overlaps).  
  - Integrates k-mer pre-screening (`obikmer.Index4mer`) for fast overlap estimation.  
  - Quality-aware scoring via `_PairingScorePeAlign`.

### 8. **Consensus & Alignment Reconstruction**
- `BuildAlignment`, `_BuildAlignment`: Reconstruct aligned sequences from DP path, reusing buffers.  
- `BuildQualityConsensus`: Generates consensus with quality-aware base selection:  
  - Mismatches resolved by higher-quality call or IUPAC ambiguity.  
  - Optional mismatch statistics recording.

### 9. **Memory & Performance Optimization**
- `PEAlignArena`: Reusable memory arena for matrices, paths, and buffers.  
  - Reduces GC pressure in high-throughput pipelines.
- Compact `uint64` encoding for scores, path lengths, and flags (`encodeValues`, `_incscore`).  
  - Enables fast comparisons during DP.

---

## Design Principles

- **IUPAC-aware**: Handles ambiguous nucleotides via `obiseq.SameIUPACNuc`.  
- **Thread-safe initialization**: `_InitDNAScoreMatrix` uses mutex guards.  
- **No allocations in hot paths**: Buffers reused across calls (arena pattern).  
- **End-gap flexibility**: Critical for read merging and primer trimming.

---

## Use Cases

| Functionality | Application |
|---------------|-------------|
| `FastLCSEGFScore`, `D1Or0` | OTU/ASV clustering, UMI deduplication |
| `LocatePattern`, `PEAlign` | Primer trimming, read merging in metabarcoding |
| `BuildQualityConsensus`, `_Backtracking` | Consensus generation post-merge |

Designed for integration into large-scale NGS pipelines—especially where speed, memory footprint, and numerical robustness are critical.
