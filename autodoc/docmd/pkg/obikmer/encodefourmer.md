# Semantic Description of `obikmer` Package

The `obikmer` package provides efficient k-mer encoding and comparison utilities for biological sequences, optimized for DNA analysis.

## Core Functionalities

1. **Nucleotide Encoding**  
   - `EncodeNucleotide(b byte)`: Maps DNA bases (A, C, G, T/U) to 2-bit values:  
     `0→A`, `1→C`, `2→G`, `3→T/U`.  
     Ambiguous or non-standard characters (e.g., N, R, Y) default to `A` (`0`).  
     Uses a lookup table for O(1) performance.

2. **4-mer Encoding**  
   - `Encode4mer(seq, buffer)`: Converts a biological sequence into overlapping 4-mers.  
     Each k-mer is encoded as an unsigned byte (0–255), where each nucleotide contributes 2 bits.  
     Supports optional buffer reuse for memory efficiency.

3. **4-mer Indexing**  
   - `Index4mer(seq, index, buffer)`: Builds an inverted index mapping each 4-mer code (0–255) to all its occurrence positions in the sequence.  
     Returns `[][]int`, where rows correspond to k-mer codes and columns list positions.

4. **Fast Sequence Comparison**  
   - `FastShiftFourMer(...)`: Compares two sequences using a FASTA-like shift-scoring algorithm.  
     - Uses precomputed 4-mer index of a reference sequence and encodes the query.  
     - Counts co-occurring 4-mers across all possible shifts (`refpos − queryPos`).  
     - Computes raw and relative scores (normalized by alignment length).  
     - Returns optimal shift, count of matching 4-mers, and maximum score (raw or relative).

## Design Highlights

- **Memory-aware**: Supports buffer reuse to minimize allocations.  
- **Robustness**: Non-canonical bases handled gracefully (defaulting to A).  
- **Performance-oriented**: O(n) encoding and indexing; efficient hash-based shift counting.  

Intended for rapid alignment-free sequence comparison in metabarcoding or metagenomic workflows.
