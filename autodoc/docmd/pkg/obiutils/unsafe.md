# `obiutils`: Unsafe String–Byte Conversions in Go

This package provides low-level, zero-copy utilities for converting between `string` and `[]byte` in Go using the `unsafe` package.

## Core Functions

- **`UnsafeBytes(str string) []byte`**  
  Converts a `string` to a mutable byte slice **without copying**, by directly accessing the underlying memory.  
  ⚠️ *Unsafe*: Modifications to the returned slice may corrupt or alter the original string (undefined behavior).  
  Use only when performance is critical and immutability can be guaranteed.

- **`UnsafeString(b []byte) string`**  
  Converts a `[]byte` to an immutable `string`, again **without copying**, by reinterpreting the byte slice’s memory as a string.  
  ⚠️ *Unsafe*: If `b` is later modified, the resulting string may become invalid (memory safety violation).  
  Requires that `b` remains immutable for the lifetime of the returned string.

## Semantic Purpose

These functions enable high-performance interop between strings and byte slices—critical in systems programming, serialization frameworks, or memory-constrained environments where allocation overhead must be avoided.

## Risks & Best Practices

- **Never mutate the returned slice or original input after conversion**.
- Prefer standard conversions (`[]byte(s)`, `string(b)`) unless profiling confirms a measurable bottleneck.
- Ensure inputs are valid and owned (e.g., not shared across goroutines without synchronization).
