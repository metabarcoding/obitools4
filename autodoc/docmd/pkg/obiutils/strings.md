# `obiutils` Package Overview

The `obiutils` package provides low-level, high-performance utilities for ASCII string and set manipulation in Go.

### Core Components

- **`AsciiSet[256]bool`**: A compact boolean lookup table for ASCII characters (0–127), optimized for membership tests.
- **Predefined Sets**:
  - `AsciiSpaceSet`: Whitespace characters (`\t\n\v\f\r `)
  - `AsciiDigitSet`, `AsciiUpperSet`, `AsciiLowerSet`
  - Derived sets: `Alpha` (letters), `Alnum` (alphanumeric)

### Key Functions

- **Set Operations**:
  - `AsciiSetFromString(s string)`: Build a set from characters in a literal.
  - `.Contains(c byte)` / `.Union()` / `.Intersect()`: Efficient membership and set algebra.

- **String Parsing & Transformation**:
  - `UnsafeStringFromBytes([]byte) string`: Zero-copy conversion (⚠️ unsafe; use only when memory safety is externally guaranteed).
  - `FirstWord(s string)`: Extract first non-whitespace token.
  - `(AsciiSet).FirstWord(...) (string, error)`: Same as above but validates characters against a restriction set.
  - `TrimLeft(s string)` (via method on *AsciiSet): Remove leading whitespace using space-aware logic.
  - `LeftSplitInTwo(s string, sep byte)`: Split at first occurrence of a separator.
  - `RightSplitInTwo(s string, sep byte)`: Split at last occurrence.

### Design Goals

- **Performance**: Avoid allocations where possible (e.g., `unsafe.String`, direct indexing).
- **Simplicity**: Focused on ASCII-only operations for speed and predictability.
- **Safety Trade-offs**: `UnsafeStringFromBytes` trades safety for efficiency; other functions are safe and bounds-checked.

Intended use: embedded systems, parsers, or performance-critical text processing where standard library overhead is undesirable.
