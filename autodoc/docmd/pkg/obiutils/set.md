# `obiutils.Set` — Generic Set Implementation in Go

This package provides a generic, type-safe set data structure for Go (1.20+), leveraging generics (`comparable` constraint). It supports common set operations with intuitive APIs.

## Core Features

- **Generic Type Support**: `Set[E]` works for any comparable type (e.g., `int`, `string`, custom structs with equality).
- **Memory-Efficient Representation**: Implemented as a map from element to empty struct (`struct{}{}`), minimizing memory overhead.
- **Immutability by Default**: Methods like `Union` and `Intersection` return *new* sets; in-place mutation is explicit (e.g., via `Add()`).

## Key Functions & Methods

| Function/Method | Description |
|-----------------|-------------|
| `MakeSet[E](vals ...E)` | Creates and returns a new set populated with given values. |
| `NewSet[E](vals ...E)` | Same as `MakeSet`, but returns a pointer (`*Set[E]`). |
| `(s Set[E]) Add(vals ...E)` | Inserts one or more elements into the set (in-place). |
| `(s Set[E]) Contains(v E) bool` | Checks membership of an element. O(1). |
| `(s Set[E]) Members() []E` | Returns all elements as a slice (order not guaranteed). |
| `(s Set[E]) String() string` | Human-readable representation via `fmt.Sprintf`. |
| `(s Set[E]) Union(s2 Set[E])` | Returns a new set containing elements from both sets. |
| `(s Set[E]) Intersection(s2 Set[E])` | Returns a new set with elements common to both sets. |

## Example Usage

```go
s1 := obiutils.MakeSet(1, 2, 3)
s2 := obiutils.NewSet("a", "b")
fmt.Println(s1.Contains(2)) // true
union := s1.Union(MakeSet(3, 4))
fmt.Println(union.Members()) // e.g., [1 2 3 4]
```

> Designed for clarity, performance, and idiomatic Go usage.
