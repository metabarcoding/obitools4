# `obiutils` Package Functional Overview

The `obiutils` package provides two core utility functions for low-level and numerical operations in Go:

- **`InPlaceToLower([]byte) []byte`**  
  Converts all ASCII uppercase letters in a byte slice to lowercase *in-place*, returning the modified slice.  
  - Non-alphabetic bytes remain unchanged.  
  - Memory-efficient: modifies input directly (no allocation of new slice).  

- **`Make2DNumericArray[T any](rows, cols int, zeroed bool) Matrix[T]`**  
  Generates a generic 2D numeric array (`Matrix`) of type `T`, supporting any comparable/numeric Go type.  
  - Parameters: number of rows, columns, and whether to initialize with zero values (`true`) or default `T` (e.g., 0 for int).  
  - Uses Go generics (`[T any]`) for type safety and flexibility.  

Both functions are thoroughly unit-tested in `*_test.go`, covering edge cases:
- Empty/nil inputs (`InPlaceToLower`)
- Various dimensions and zero-initialization modes (`Make2DNumericArray`)

Tests use `reflect.DeepEqual` for structural comparison and subtests via `t.Run`.  
The package assumes a custom type alias: `type Matrix[T any] [][]T`.
