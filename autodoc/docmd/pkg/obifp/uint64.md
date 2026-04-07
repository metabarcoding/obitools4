# Uint64 Type Functionalities Overview

The `obifp` package provides a custom `Uint64` type wrapping Go’s native 64-bit unsigned integer (`uint64`) to support arithmetic, bitwise operations, and type conversions in a structured way.

## Core Operations

- **`Zero()` / `MaxValue()`**: Returns the zero and maximum representable values, respectively.
- **`IsZero()` / `Equals(v)`**: Checks if the value is zero or equal to another.
- **`Cmp(v)`, `LessThan(v)`**, etc.: Standard comparison operations returning `-1/0/+1` or boolean results.

## Arithmetic with Overflow Detection

- **Add/Sub/Mul**: Performs 64-bit addition, subtraction, and multiplication.
  - Uses `math/bits` for low-level operations (`bits.Add64`, etc.).
  - Panics on overflow (carry ≠ 0), enforcing strict safety.

## Bitwise Operations

- **`And`, `Or`, `Xor`, `Not()`**: Standard bitwise logic operations.
- **`LeftShift(n)` / `RightShift(n)`**:
  - Shifts bits left/right by *n* positions.
  - Uses internal `LeftShift64`/`RightShift64`, supporting *carry-in* for multi-word arithmetic.

## Extended Precision Conversions

- **`Uint128()` / `Uint256()`**: Casts the 64-bit value into larger unsigned integer types (zero-extended).
- **`Set64(v)`**: Reassigns the internal value from a raw `uint64`.

## Utility & Logging

- **`AsUint64()`**: Extracts the underlying `uint64`.
- **Warning on overflow in shift operations** (e.g., shifts ≥ 128 bits) via `obilog.Warnf`.

> Designed for use in high-precision or cryptographic contexts where explicit overflow handling and type safety are critical.
