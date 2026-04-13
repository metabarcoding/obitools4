# `obiutils` Package: Semantic Overview

This Go package (`obiutils`) provides generic utilities for numerical and matrix operations, leveraging generics (Go 1.18+). It defines foundational types and helper functions for working with multidimensional data structures.

- **Type Interfaces**  
  - `Integer`: Constraint covering signed integer types (`int`, `int8`–`int64`).  
  - `Float`: Constraint for floating-point types (`float32`, `float64`).  
  - `Numeric`: Union of both, enabling generic numeric functions.

- **Data Structures**  
  - `Vector[T]`: A slice-based vector (`[]T`).  
  - `Matrix[T]`: A row-major representation of a 2D matrix (`[][]T`), backed by contiguous memory for performance.

- **Core Functions**  
  - `Make2DArray[T]`: Allocates a zero-initialized, contiguous-row-major matrix of arbitrary type `T`.  
  - `Make2DNumericArray[T]`: Same as above, but restricted to numeric types; optionally pre-fills with zeros if `zeroed=true`.

- **Matrix Methods**  
  - `.Column(i int)`: Extracts column `i` as a slice (not row-wise access).  
  - `.Rows(i ...int)`: Returns a new matrix containing only the specified row indices.  
  - `.Dim() (int, int)`: Returns `(rows, cols)` safely handling `nil` or empty matrices.

The design prioritizes memory efficiency (via contiguous backing arrays), type safety through generics, and ergonomic access patterns for linear algebra-like workflows.
