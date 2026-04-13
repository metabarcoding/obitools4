# `obifp.Uint128` Package — Semantic Feature Overview

This Go package provides a 128-bit unsigned integer type (`Uint128`) with comprehensive arithmetic, comparison, and bitwise operations. Internally represented as two `uint64` limbs (`w1`: high, `w0`: low), it supports:

- **Arithmetic Operations**  
  - `Add`, `Sub`, `Mul` (128×128), and `Mul64` (scalar multiplication)  
  - Division: `Div`, `Mod`, and combined quotient/remainder via `QuoRem` (and their 64-bit variants)  
- **Comparison & Equality**  
  - `Cmp`, `Equals`, `LessThan`/`GreaterThan`, and their inclusive variants (`≤`, `≥`)  
  - Support for comparing against both `Uint128` and native `uint64` values  
- **Bitwise Operations**  
  - Logical AND (`And`), OR (`Or`), XOR (`Xor`) between two `Uint128`s  
  - Bitwise NOT (`Not`) — inverts all bits of the value  
- **Conversion & Utility**  
  - `AsUint64()` safely truncates to lower 64 bits (assumes upper limb is zero)  

All operations handle overflow/underflow correctly, including carry propagation in addition and borrow handling in subtraction. Tests cover edge cases: zero values, max `uint64` boundaries (e.g., wrapping in addition/subtraction), and large multiplications. Designed for cryptographic or high-precision numeric use where native integer types are insufficient.
