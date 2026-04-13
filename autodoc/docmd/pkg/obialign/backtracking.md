# Semantic Description of `obialign` Backtracking Module

The `_Backtracking` function implements a **traceback algorithm** for sequence alignment, reconstructing the optimal path through an alignment matrix.

## Core Functionality

- **Input**:  
  - `pathMatrix`: Encodes alignment decisions (match/mismatch/gap) as integers.  
  - `lseqA`, `lseqB`: Lengths of sequences A and B.  
  - `path`: Pre-allocated slice to store the traceback path.

- **Output**: A compact representation of alignment steps, alternating between:
  - Diagonal moves (`ldiag`): Matches/mismatches (one step in both sequences).
  - Horizontal/vertical moves (`lleft` or `lup`): Gaps in sequence B (horizontal) or A (vertical).

## Algorithm Highlights

- **Reverse traversal** from `(lseqA−1, lseqB−1)` to origin.
- **Batching logic**: Consecutive gaps in same direction are aggregated (e.g., `lleft += step`) to compress run-length encoding.
- **Path reconstruction**: Steps are pushed *backwards* into the `path` slice using a moving pointer `p`.
- **Memory efficiency**: Uses `slices.Grow()` to preallocate space and logs resizing for debugging.

## Encoded Path Semantics

Each pair in the returned slice encodes:
- `[diag_count, move_type]`, where `move_type` is either a gap length (`lleft > 0`: horizontal, or `lup < 0`: vertical) or zero (end of diagonal run).

## Use Case

Enables efficient reconstruction and serialization of alignment paths—ideal for tools requiring low-level control over dynamic programming backtracking (e.g., pairwise aligners, edit-distance decompositions).
