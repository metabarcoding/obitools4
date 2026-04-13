## Uint128 Type in `obifp`: Semantic Overview

This Go package defines a custom 128-bit unsigned integer type (`Uint128`) composed of two `uint64` limbs (high and low). It provides comprehensive arithmetic, comparison, bitwise operations, and type conversions.

- **Basic Constructors**: `Zero()`, `MaxValue()` initialize the smallest/largest possible values.
- **State Checks**: `IsZero()`, and equality/comparison methods (`Equals`, `Cmp`, `<`, `>`, etc.) enable conditional logic.
- **Type Casting**: Safe conversions to/from smaller (`Uint64`, `uint64`) and larger (`Uint256`) integer types, with overflow warnings where applicable.
- **Arithmetic**: Full support for addition (`Add`, `Add64`), subtraction (`Sub`), multiplication (`Mul`, `Mul64`) — with panic on overflow.
- **Division & Modulo**: Integer division (`Div`, `Div64`) and remainder (`Mod`, `Mod64`), implemented via optimized quotient-remainder pairs (`QuoRem`, `QuoRem64`) using hardware-assisted 64-bit operations.
- **Bit Manipulation**: Left/right shifts (`LeftShift`, `RightShift`), and bitwise logic: AND, OR, XOR, NOT.
- **Utility**: Direct access to low limb via `AsUint64()`.

All operations preserve 128-bit precision, with strict overflow checking for correctness in high-precision contexts (e.g., bioinformatics counting).
