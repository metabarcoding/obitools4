# Obifp Package: Generic Fixed-Point Unsigned Integer Operations

This Go package (`obifp`) provides a generic, type-safe interface for fixed-point unsigned integer arithmetic over three size variants: `Uint64`, `Uint128`, and `Uint256`.

## Core Interface: `FPUint[T]`

The interface defines a unified API for unsigned integer types, supporting:

- **Initialization & Conversion**:  
  - `Zero()`, `Set64(v)`: Create zero or set from a `uint64`.  
  - `AsUint64()`: Downcast to standard `uint64`.

- **Logical Operations**:  
  - Bitwise: `And`, `Or`, `Xor`, `Not`.  
  - Shifts: `LeftShift(n)`, `RightShift(n)`.

- **Arithmetic**:  
  - Addition (`Add`), subtraction (`Sub`), multiplication (`Mul`). Division is commented out—likely reserved for future implementation.

- **Comparison**:  
  - Full ordering: `<`, `<=`, `>`, `>=`.

- **Utility Predicates**:  
  - `IsZero()` for zero-checking.

## Helper Functions

- `ZeroUint[T]`: Returns the neutral element (zero) for type `T`.  
- `OneUint[T]`: Constructs value 1 via `Set64(1)`.  
- `From64[T]`: Converts a standard Go `uint64` into the generic type.

All operations are **method-chaining friendly** (return `T`, not pointers), enabling fluent syntax. The design promotes correctness and performance in cryptographic or financial contexts where large, fixed-size integers are required.
