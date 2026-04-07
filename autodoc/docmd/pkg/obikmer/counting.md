# Semantic Description of `obikmer` Package

This Go package provides utilities for **k-mer (specifically 4-mer) counting and comparison** of biological sequences.

## Core Functionalities

1. **`Count4Mer(seq, buffer, counts)`**  
   Counts occurrences of all possible 16-mer (4-nucleotide) subsequences in a `BioSequence`.  
   - Encodes each 4-mer into an integer (0–255) using `Encode4mer`.  
   - Populates a fixed-size `[256]uint16` table (`Table4mer`) with counts.  
   - Reuses or allocates the `counts` buffer as needed.

2. **`Common4Mer(count1, count2)`**  
   Computes the *intersection* of two 4-mer frequency profiles: sum over all k-mers of `min(count1[k], count2[k])`.  
   Used to measure shared content between sequences.

3. **`Sum4Mer(count)`**  
   Returns the total number of 4-mers in a profile (i.e., sum over all entries).

## Distance & Similarity Bounds

4. **`LCS4MerBounds(count1, count2)`**  
   Estimates bounds for the *Longest Common Subsequence* (LCS) length between two sequences based on 4-mer profiles:  
   - **Lower bound**: `common_kmers + (3 if common > 0 else 0)`  
   - **Upper bound**: `min(total1, total2) + 3 − ceil((min_total – common)/4)`  
   Leverages the fact that overlapping k-mers constrain possible alignments.

5. **`Error4MerBounds(count1, count2)`**  
   Estimates bounds for *alignment errors* (e.g., mismatches + indels):  
   - **Upper bound**: `max_total − common_kmers + 2 * floor((common_kmers + 5)/8)`  
   - **Lower bound**: `ceil(upper_bound / 4)`  
   Provides fast, approximate error estimates without full alignment.

## Use Case

Designed for **high-performance comparison of NGS reads** (e.g., in metabarcoding), where exact alignment is too costly, and k-mer-based heuristics enable scalable similarity estimation.
