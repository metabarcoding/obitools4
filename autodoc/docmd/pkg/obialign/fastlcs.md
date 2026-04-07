# Semantic Description of `obialign` Package

The `obialign` package provides low-level utilities for efficiently encoding, decoding, and manipulating alignment-related metrics—specifically **score**, **path length**, and an **out-flag**—within compact 64-bit integers. This design supports high-performance operations in sequence alignment pipelines (e.g., OBITools4).

- **Core Encoding Strategy**:  
  A `uint64` encodes three fields: a *score* (upper bits), an inverted path *length*, and a single-bit flag indicating whether the value represents an "out" (i.e., terminal/invalid) state.

- **`encodeValues(score, length int, out bool)`**:  
  Packs `score`, `-length-1` (to preserve ordering via unsigned comparison), and the `out` flag into one integer. The most significant bit (bit 32) marks out-values.

- **`decodeValues(value uint64)`**:  
  Reverses encoding: extracts score, reconstructs original length via `((value + 1) ^ mask)`, and checks the out-flag.

- **Utility Bitwise Helpers**:
  - `_incpath(value)`: decrements stored length (since it's negated, subtraction increases actual path).
  - `_incscore(value)`: increments score by `1 << wsize`.
  - `_setout(value)`: clears the out-flag, marking value as *not* terminal.

- **Predefined Constants**:
  - `_empty`: neutral state (score=0, length=0).
  - `_out`/`_notavail`: sentinel values for invalid or unavailable paths (high length, score=0).

This compact representation enables fast comparisons and updates during dynamic programming or alignment graph traversal—critical for scalability in large-scale metabarcoding analyses.
