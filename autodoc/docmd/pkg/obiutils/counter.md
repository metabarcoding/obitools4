# `obiutils.Counter`: Thread-Safe Atomic Counter

A minimal, thread-safe counter implementation in Go.

## Features
- **Atomic increment/decrement**: `Inc()` and `Dec()` modify the internal counter atomically using a mutex.
- **Current value retrieval**: `Value()` safely returns the current count without modifying it.
- **Initial value support**: Constructor accepts an optional initial integer (defaults to `0`).
- **Closure-based API**: Encapsulates state and synchronization behind clean, functional methods.
- **No external dependencies**: Uses only the standard library (`sync`).

## Usage Example
```go
counter := obiutils.NewCounter(10) // start at 10
fmt.Println(counter.Inc())         // → 11
fmt.Println(counter.Dec())         // → 10
fmt.Println(counter.Value())       // → 10 (unchanged)
```

## Thread Safety
All operations are protected by a `sync.Mutex`, ensuring correctness in concurrent environments.

## Design Notes
- Immutable interface: methods return updated values, not pointers.
- No reset method provided—intentionally minimal and focused on core counting semantics.
