# Semantic Description of `obialign` Package

The `obialign` package provides high-performance functions for computing the **Longest Common Subsequence (LCS)** between two biological sequences, with support for error tolerance and end-gap-free alignment.

## Core Algorithm

- Implements a **Needleman-Wunsch** dynamic programming algorithm optimized for speed and memory efficiency.
- Uses bit-packed encoding (`uint64`) to store score, path length, and gap status in a compact form.
- Leverages **diagonal banding** to restrict computation only within the allowed error margin, reducing time and space complexity.

## Scoring Scheme

- **Match**: +1 point  
- **Mismatch or gap (indel)**: 0 points  

## Key Functions

1. `FastLCSEGFScoreByte(bA, bB []byte, maxError int, endgapfree bool, buffer *[]uint64) (int, int, int)`  
   - Computes LCS score and alignment length between raw byte sequences.  
   - If `endgapfree` is true, ignores leading/trailing gaps (useful for read alignment).  
   - Returns `(score, length, end_position)`; `end_position` marks where the LCS ends in sequence A.  
   - Returns `-1, -1, -1` if the actual error count exceeds `maxError`.

2. `FastLCSEGFScore(seqA, seqB *obiseq.BioSequence, maxError int, buffer ...)`  
   - Wrapper for `FastLCSEGFScoreByte` with end-gap-free mode enabled by default.  
   - Designed for standard biosequence inputs.

3. `FastLCSScore(seqA, seqB *obiseq.BioSequence, maxError int, buffer ...)`  
   - Computes standard LCS (including end gaps). Returns `(score, alignment_length)`.

## Features

- **Error-bounded**: Supports `maxError = -1` (unlimited) or a fixed max number of mismatches + gaps.
- **Memory-efficient**: Reuses user-provided or auto-created buffers to avoid allocations during repeated calls.
- **IUPAC-aware**: Uses `obiseq.SameIUPACNuc()` to handle ambiguous nucleotide codes (e.g., `R`, `Y`).
- **Optimized for short reads**: Particularly suited to high-throughput sequencing data alignment tasks (e.g., in OBITools4).

## Use Cases

- Molecular barcode/UMI clustering  
- Read-to-reference alignment in amplicon sequencing  
- Similarity filtering of biological sequences
