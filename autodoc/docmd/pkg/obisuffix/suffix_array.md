# Suffix Array Implementation for Biological Sequences

This Go package (`obisuffix`) provides a suffix array data structure tailored for biological sequence analysis. It supports efficient lexicographic sorting and common-prefix computation over all suffixes of a set of sequences.

## Core Types

- **`Suffix`**: Represents one suffix by storing the sequence index (`Idx`) and starting position (`Pos`).  
- **`SuffixArray`**: Holds a collection of `Suffix`, the original sequences (`Sequences`), and cached common-prefix lengths (`Common`).  

## Key Functions

- **`BuildSuffixArray(data)`**: Constructs a suffix array by enumerating *all* suffixes from all input sequences, then sorts them lexicographically using a custom comparator (`SuffixLess`).  
- **`CommonSuffix()`**: Computes the length of shared prefix between each adjacent pair in the sorted suffix array (i.e., `LCP`-like values), caching results for reuse.  
- **`String()`**: Returns a human-readable table with columns: `Common`, sequence index, position, and suffix string.  

## Semantic Features

- **Lexicographic ordering**: Suffixes are sorted by their nucleotide/amino-acid content; ties break first by shorter length, then lower index, finally earlier position.  
- **Efficiency**: Avoids redundant comparisons via memoization of `Common` values and stable sorting.  
- **Biological relevance**: Designed for use with `obiseq.BioSequenceSlice`, supporting DNA, RNA, or protein sequences.  
- **Transparency**: The `String()` method enables quick inspection of suffix relationships and overlaps.

This structure is foundational for tasks like repeat detection, alignment-free comparison, or pattern mining in multi-sequence datasets.
