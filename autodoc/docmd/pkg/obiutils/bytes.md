# `InPlaceToLower` Function — Semantic Description

The `obiutils.InPlaceToLower` function provides a high-performance, memory-efficient utility for converting ASCII uppercase letters to lowercase **in place**, without allocating new data structures.

## Core Functionality
- Takes a `[]byte` slice (`data`) as input.
- Iterates over each byte, identifying uppercase ASCII characters (i.e., `'A'`–`'Z'`, values `65`–`90`).
- Converts each uppercase byte to its lowercase counterpart using a bitwise OR with `32`, leveraging the ASCII encoding property:  
  `lowercase = uppercase | 0b0010_0000` (since `'a' - 'A' = 32`).  
- Returns the **same** `[]byte` slice, now mutated in-place.

## Key Characteristics
- ✅ **Zero-copy**: No new memory is allocated—ideal for performance-critical or low-level contexts (e.g., streaming, embedded systems).
- ✅ **ASCII-safe**: Only modifies bytes in the `'A'`–`'Z'` range; other bytes (e.g., digits, symbols, non-ASCII) remain unchanged.
- ✅ **Idiomatic Go**: Uses idioms like `range` with index/value and bitwise optimization.
- ⚠️ **Destructive**: Input data is permanently modified—callers must clone if preservation is needed.

## Use Cases
- Preprocessing raw HTTP headers or payloads.
- Optimizing case-insensitive comparisons in high-throughput systems.
- Embedded tools where GC pressure or heap allocation must be minimized.

## Example
```go
buf := []byte("HTTP/1.1 200 OK")
InPlaceToLower(buf) // buf is now []byte("http/1.1 200 ok")
```
