# `obiutils` Package: Semantic Overview

The `obiutils` package provides generic and reflection-based utilities for computing minima and maxima across multiple data structures in Go.

## Core Features

- **Generic `MinMax` / `Min/MaxSlice`**:  
  - `MinMax[T constraints.Ordered]`: Returns the ordered pair `(min, max)` of two values.  
  - `MinMaxSlice[T constraints.Ordered]`: Finds min and max in a slice of ordered types (panics on empty input).

- **Map-based Min/Max**:  
  - `MinMap` / `MaxMap`: Returns the key and value of the smallest/largest *value* in a map (errors on empty maps).

- **Unified `Min` / `Max` Functions**:  
  - Accepts *any* Go value: single scalar, slice/array/map.  
  - Uses reflection to dispatch logic based on runtime type (`reflect.Kind`).  
  - Supports ordered kinds: integers, floats, strings (signed/unsigned ints via `constraints.Ordered` subset).  
  - Returns an error for unsupported or empty containers.

- **Helper Reflection Functions**:  
  - `minFromIterable` / `maxFromIterable`: Scan slices/arrays.  
  - `minFromMap` / `maxFromMap`: Iterate over map values (ignores keys in comparisons).  
  - `isOrderedKind`, `less`, `greater`: Internal comparison logic for reflection-based ordering.

## Design Highlights

- **Type Safety & Generics**: Leverages Go 1.18+ generics for compile-time type constraints where possible.
- **Flexibility**: The `Min(data interface{})` / `Max(...)` functions allow a *single API* for heterogeneous inputs.
- **Error Handling**: Explicit errors (e.g., `"empty slice"`, `"unsupported type"`), no panics for user-facing APIs.
- **Fallback Support**: Checks if the input has a `Min()`/`Max()` method (via reflection) before falling back to generic logic.

## Limitations

- Reflection-based paths are slower than direct generics.
- No support for custom types without ordering defined (e.g., structs unless they satisfy `constraints.Ordered`).
- Maps compare only *values*; keys are irrelevant for min/max selection.
