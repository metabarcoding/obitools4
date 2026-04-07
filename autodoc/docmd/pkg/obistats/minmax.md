# `obistats` Package — Core Statistical Functions

The `obistats` package provides generic, type-safe implementations of fundamental descriptive statistics for numeric types in Go.

## Key Functions

- **`Max[T]()`**  
  Returns the maximum value in a slice of numeric types (`int`, `int8`–`64`, `float32/64`).  
  *Implementation*: Iterates once, tracking the largest element.

- **`Min[T]()`**  
  Returns the minimum value in a slice of numeric types (including unsigned integers: `uint`, `uint8`–`64`).  
  *Implementation*: Single-pass scan, comparing each element to the current minimum.

- **`Mode[T]()`**  
  Computes the *most frequent* value (mode) for signed integer types only (`int`, `int8`–`64`).  
  *Implementation*: Builds a frequency map, then selects the value with highest count.

## Design Notes

- **Generics**: All functions use Go type parameters (`[T ...]`) for compile-time safety and performance.
- **Type Scope**:
  - `Max` supports signed integers + floats (no unsigned).
  - `Min` includes all integer variants.
  - `Mode` is restricted to signed integers (due to map key constraints and semantics).
- **Assumptions**: Input slices are non-empty; no explicit error handling for edge cases (e.g., empty input).
- **Use Case**: Lightweight, reusable utility functions suitable for statistical pipelines or exploratory data analysis.

> ⚠️ *Note*: No mean, median, variance, or standard deviation functions are provided in this excerpt.
