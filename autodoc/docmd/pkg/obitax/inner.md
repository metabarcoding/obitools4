# `obitax` Package: String Interning with Thread-Safe Storage

This Go package (`obitax`) provides a **thread-safe string interner**—a data structure that deduplicates identical strings by storing only one copy per unique value and returning shared references.

## Core Components

- **`InnerString` struct**  
  Holds:
  - `index`: A map from string values to pointers (ensuring identity via pointer equality).
  - `lock`: An embedded `sync.RWMutex` to guarantee safe concurrent access.

- **Constructor: `NewInnerString()`**  
  Initializes an empty interner with a preallocated map.

- **Method: `Innerize(value string) *string`**  
  - Stores a new unique value (after cloning via `strings.Clone`) if absent.
  - Returns the pointer to either:
    - The newly interned string, or  
    - An existing one (if already present).
  - Ensures **no duplicate string data** is stored for equal values.
  - Fully thread-safe via write lock.

- **Method: `Slice() []string`**  
  Returns a snapshot of all interned strings as a slice (copying values, not pointers).
  - Not safe for concurrent writes during iteration.
  - Suitable for inspection or debugging.

## Semantic Use Cases

- **Memory optimization**: Avoid repeated allocation of identical strings (e.g., in parsing, serialization).
- **Pointer-based identity checks**: Use `==` on returned pointers to test string equality efficiently.
- **Concurrent safety**: Designed for use in multi-goroutine environments (e.g., HTTP servers, pipelines).

## Design Notes

- Uses `strings.Clone()` to decouple interned strings from original input lifetimes.
- Interning is **append-only**—no removal mechanism provided (implied by semantics of a simple interner).
- Returns `*string` to enable fast equality comparisons and reduce memory footprint.

> **Note**: This is a minimal, efficient interner—ideal for read-heavy or batched deduplication scenarios.
