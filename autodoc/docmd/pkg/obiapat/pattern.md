# Apat Package: Pattern Matching for Biological Sequences

The `obiapat` Go package provides high-performance pattern matching over biological sequences using the **Apat algorithm**, a C-based implementation wrapped in Go. It supports fuzzy matching (with mismatches and indels), reverse-complement patterns, memory-safe resource management via finalizers, and efficient filtering of non-overlapping matches.

## Core Types

- `ApatPattern`: Represents a compiled pattern (up to 64 bp), supporting IUPAC ambiguity codes (`W`, `[AT]`), negated bases (`!A`), and fixed positions (`#`).  
- `ApatSequence`: Wraps a biological sequence (from `obiseq.BioSequence`) for fast matching, with optional circular topology support and memory recycling.

## Key Functions & Methods

- `MakeApatPattern(pattern string, errormax int, allowsIndel bool)`: Compiles a pattern with max error tolerance and optional indels.  
- `ReverseComplement()`: Returns the reverse-complemented pattern (useful for DNA strand symmetry).  
- `FindAllIndex(...)`: Returns all matches as `[start, end, errors]`, supporting partial sequence searches.  
- `IsMatching(...)`: Boolean check for presence of at least one match in a region.  
- `BestMatch(...)`: Finds the *best* (lowest-error) match, with local realignment for indel-containing patterns.  
- `FilterBestMatch(...)`: Returns *non-overlapping* matches, prioritizing lower-error occurrences.  
- `AllMatches(...)`: Filters and refines all valid matches (including indel-aware alignment).  
- `Free()`, `Len()`: Explicit memory cleanup and length queries.

## Implementation Notes

Internally, the package uses `cgo` to interface with C structures (`Pattern`, `Seq`) allocated via custom memory management. Finalizers ensure safe deallocation, while unsafe pointer arithmetic avoids data copying during search (e.g., `unsafe.SliceData`). Logging is integrated via Logrus.

This package enables scalable, low-level pattern mining in NGS data preprocessing pipelines (e.g., primer detection, adapter trimming).
