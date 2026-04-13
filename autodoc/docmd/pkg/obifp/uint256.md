# Uint256 Type and Operations — Semantic Overview

The `obifp` package provides a custom 256-bit unsigned integer type (`Uint256`) implemented in Go, composed of four 64-bit limbs (`w0` to `w3`). It supports arithmetic, comparison, bitwise operations, and safe casting with overflow detection.

- **Core Representation**: `Uint256` stores values as four 64-bit words, enabling arbitrary-precision unsigned integers up to $2^{256} - 1$.

- **Utility Methods**:
  - `Zero()` / `MaxValue()`: Return the neutral and maximum values.
  - `IsZero()`, `Equals(v)`, comparison methods (`LessThan`, etc.): Enable logical and ordering checks.

- **Casting & Conversion**:
  - `Uint64()`, `Uint128()` downcast with warnings on overflow.
  - `Set64(v)`: Initializes from a standard `uint64`.
  - `AsUint64()`: Direct access to least-significant limb.

- **Bitwise Operations**:
  - `And`, `Or`, `Xor`, `Not`: Standard bitwise logic per limb.

- **Shifts**:
  - `LeftShift(n)` / `RightShift(n)`: Multi-limb shifts with carry propagation.

- **Arithmetic**:
  - `Add(v)`, `Sub(v)` / `Mul(v)`: Use Go’s `math/bits` for carry-aware operations; panic on overflow.
  - `Div(v)`: Implements long division via repeated subtraction of shifted multiples; panics on zero divisor.

- **Safety & Logging**:
  - Warnings via `obilog.Warnf` for silent overflows during narrowing casts.
  - Panics on arithmetic overflow or division-by-zero using `log.Panicf`.

This type is suitable for cryptographic, genomic (OBITools), or high-precision counting use cases requiring precise control over large unsigned integers.
