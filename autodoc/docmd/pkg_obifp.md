# `obifp`: Semantic Overview of Public API

The `obifp` package provides a family of fixed-size, arbitrary-precision unsigned integer types—`Uint64`, `Uint128`, and `Uint256`—designed for high-precision arithmetic where overflow safety, bitwise control, and type consistency are critical (e.g., cryptography, genomics with OBITools). All types share a unified interface (`FPUint[T]`) and enforce strict correctness via panics on overflow/underflow or division-by-zero.

## Core Principles

- **Explicit precision**: No silent truncation; narrowing casts emit warnings (via `obilog.Warnf`).
- **Panic-on-error semantics**: Arithmetic operations (`Add`, `Sub`, `Mul`, etc.) panic on overflow/underflow.
- **Bit-level fidelity**: Shifts and bitwise operations operate across full bit-width with carry propagation.

---

## Unified Interface: `FPUint[T]`

All three types (`Uint64`, `Uint128`, `Uint256`) implement this generic interface:

- **Construction & Initialization**
  - `Zero() T`: Returns the additive identity.
  - `Set64(v uint64) T`: Initializes from a native 64-bit value (zero-extended).
  - `OneUint[T]`: Helper to construct the value *1*.

- **Downcasting & Utility**
  - `AsUint64() uint64`: Extracts the least-significant limb (assumes higher limbs are zero; warns if not).
  - `IsZero() bool`: Checks for equality with zero.

- **Logical & Bitwise Operations**
  - `And(v T)`, `Or(v T)`, `Xor(v T)` — bitwise logic between two values of same type.
  - `Not() T` — inverts all bits (two’s complement style for unsigned).
  - `LeftShift(n uint) T`, `RightShift(n uint) T` — multi-limb shifts with carry handling; warns if shift ≥ full bit-width.

- **Arithmetic**
  - `Add(v T)`, `Sub(v T)` — with carry/borrow propagation; panics on overflow/underflow.
  - `Mul(v T)` — full-width multiplication (uses hardware-optimized limb-wise ops); panics on overflow.
  - *Division (`Div`, `Mod`) is implemented only for concrete types (see below).*

- **Comparison**
  - `Cmp(v T) int` — returns `-1`, `0`, or `+1`.
  - Overloaded operators: `<`, `<=`, `>`, `>=` (all returning `bool`).

---

## Concrete Types & Specialized Features

### ✅ `Uint64`
- Native Go `uint64` wrapper with strict overflow checking.
- Uses `math/bits.Add64`, `Mul64` internally for correctness.
- Supports conversion to larger types: `Uint128()`, `Uint256()`.

### ✅ `Uint128`
- Internally: two limbs (`w0`, `w1`).
- **Arithmetic**:
  - Full support: `Add(v)`, `Sub(v)`, `Mul(v)` (128×128), and scalar variants: `Add64`, `Mul64`.
  - Division & Modulo:
    - `Div(v)`, `Mod(v)` — integer division with remainder.
    - `QuoRem(v Uint128) (q, r Uint128)` — combined quotient/remainder.
    - `Div64`, `Mod64` for division by 64-bit scalar.
- **Bitwise**: Full support (`And`, `Or`, `Xor`, `Not`), plus shifts.
- **Conversion**:
  - Safe upcast to `Uint256`.
  - Downcast to `uint64` via `AsUint64()` (warns if high limb ≠ 0).

### ✅ `Uint256`
- Internally: four limbs (`w0` to `w3`) — supports values up to $2^{256} - 1$.
- **Arithmetic**:
  - `Add(v)`, `Sub(v)` — limb-wise with carry/borrow.
  - `Mul(v)` — schoolbook multiplication across limbs; panics on overflow.
  - `Div(v)`: Long division implementation (repeated subtraction of shifted multiples); panics on zero divisor.
- **Shifts**: Multi-limb shifts with carry propagation across all limbs.
- **Conversion**:
  - Downcast to `Uint128()` / `AsUint64()`, with overflow warnings.
  - Upcast from smaller types via implicit zero-extension.

---

## Helper Functions (Generic)

- `ZeroUint[T FPUint[T]]() T`: Returns zero for type parameter.
- `From64[T FPUint[T]](v uint64) T`: Converts native 64-bit to typed value.

All operations are **value-returning** (no in-place mutation), enabling fluent chaining and immutability.

> ⚠️ **Design Note**: Division methods are *not* part of the generic `FPUint[T]` interface (commented out), but are implemented concretely for each type. This reflects performance/complexity trade-offs and leaves room to extend later.
