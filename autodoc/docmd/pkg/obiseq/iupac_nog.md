# Semantic Description of `obiseq` Package

The `obiseq` package provides utilities for handling **IUPAC nucleotide ambiguity codes** in biological sequences.

## Core Components

- `_iupac`: A lookup table mapping lowercase ASCII letters (`a`–`z`) to numeric IUPAC nucleotide codes:
  - `A=1`, `C=2`, `G=4`, `T/U=8` (standard bases)
  - Ambiguous codes are bitwise OR combinations:  
    e.g., `R = A|G = 1+4=5`, `Y = C|T = 2+8=10`, etc.
- Invalid or non-nucleotide characters map to `0`.

## Key Functionality

### `SameIUPACNuc(a, b byte) bool`
Performs **case-insensitive comparison** of two nucleotide symbols using IUPAC ambiguity rules.

- Converts uppercase letters to lowercase via bitwise OR (`|= 32`).
- For valid nucleotides, checks if their IUPAC codes have **non-zero bitwise AND**:
  - Returns `true` only if the symbols share at least one possible base.
    *Example*: `'R' & 'A' → (5 & 1) = 1 > 0 ⇒ true`  
    `'Y' & 'G' → (10 & 4) = 0 ⇒ false`
- For non-IUPAC or invalid characters, falls back to exact equality (`a == b`).

## Use Case

Enables robust comparison of DNA/RNA sequences where ambiguity codes (e.g., `N`, `R`, `W`) are used—critical for alignment, variant calling, or primer design tools.
