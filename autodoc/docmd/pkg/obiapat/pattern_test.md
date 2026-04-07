# Semantic Description of `obiapat` Package Functionality

The `obiapat` package provides utilities for constructing and representing **approximate sequence patterns**—flexible biological or symbolic string templates supporting mismatches, insertions, and deletions.

## Core Functionality

- **`MakeApatPattern(pattern string, errormax int, allowsIndel bool)`**  
  Parses a pattern specification (e.g., `"A[T]C!GT"`) and returns an internal representation (`*ApatPattern`) suitable for approximate matching.

  - `pattern`: A string where:
    - Standard characters (e.g., `'A'`, `'C'`) denote exact matches.
    - Brackets `[X]` indicate *optional* or *variable positions*, e.g., ambiguity (like IUPAC codes).
    - Exclamation `!` marks positions where **errors** (substitutions) are permitted.
  - `errormax`: Maximum number of allowed errors (mismatches or indels, depending on flags).
  - `allowsIndel`: Boolean flag enabling/disabling insertion/deletion operations.

## Behavior & Semantics

- Returns a compiled pattern object (non-nil) on success; errors may arise from malformed input or invalid parameters.
- Supports three modes:
  - **Exact matching** (`errormax = 0`, `allowsIndel = false`).
  - **Substitution-only approximation** (`errormax > 0`, `allowsIndel = false`).
  - **Full approximate matching with indels** (`errormax > 0`, `allowsIndel = true`).

## Testing Coverage

The provided test suite validates:
- Valid pattern parsing across different configurations.
- Correct handling of `nil` vs. non-nil output pointers.
- Robustness against error conditions (e.g., invalid inputs would trigger expected errors).

In summary, `obiapat` enables efficient definition and handling of *approximate regular expressions* tailored for sequence analysis in bioinformatics or pattern recognition contexts.
