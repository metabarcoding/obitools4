# `obiutils.Abs` — Generic Absolute Value Function

This package provides a **type-generic utility function** for computing the absolute value of signed numeric types in Go.

## Function Signature

```go
func Abs[T constraints.Signed](x T) T
```

- **Generic constraint**: `T` must satisfy `constraints.Signed`, i.e., any signed integer type (`int`, `int8`–`int64`) or floating-point type (via future Go versions supporting floats in `constraints.Signed`).
- **Input**: A value of type `T`.
- **Output**: The absolute (non-negative) counterpart, same type as input.

## Semantics

- Returns `x` if `x ≥ 0`.
- Otherwise, returns `-x`, effectively flipping the sign.
- Handles all signed numeric types uniformly — no need for type-specific overloads.

## Example Usage

```go
absInt := obiutils.Abs(-5)        // → 5 (type: int)
absFloat64 := obiutils.Abs(-3.14) // → 3.14 (type: float64)
```

## Design Rationale

- Leverages Go generics for **reusability** and type safety.
- Avoids duplication across `AbsInt`, `AbsFloat64`, etc.
- Follows Go’s standard library conventions (e.g., similar to `math.Abs` but *generic* and not limited to floats).

## Limitations

- Does **not** support unsigned types (by design: `constraints.Signed` excludes them).
- For floating-point special cases (`NaN`, `-0.0`) behavior matches native negation semantics.

## Dependencies

- Requires `golang.org/x/exp/constraints` for the generic type constraint.
