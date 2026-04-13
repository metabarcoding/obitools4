# `obiutils` Package: Set Implementation in Go

The `obiutils` package provides a generic, type-safe set data structure for Go (v1.18+), along with comprehensive unit tests.

## Core Features

- **Generic Set Type**: Implemented as `Set[T]`, using a map for O(1) membership checks.
- **Constructors**:
  - `MakeSet[T](...T)` returns a new set populated with given elements.
  - `NewSet[T]()` allocates an empty pointer to a set; useful for dynamic initialization.
- **Methods**:
  - `Add(...T)` inserts one or more elements (idempotent).
  - `Contains(T) bool` checks membership.
  - `Members() []T` returns a sorted slice of elements (deterministic iteration).
  - `String() string` provides human-readable representation (`[a b c]` format).
- **Set Operations**:
  - `Union(other Set[T]) Set[T]`: returns a new set with elements in either operand.
  - `Intersection(other Set[T]) Set[T]`: returns a new set with elements common to both.

## Test Coverage

Unit tests validate:
- Set creation (empty, single/multiple values).
- Element addition and membership.
- String formatting for various sizes.
- Correctness of union/intersection across edge cases (empty sets, disjoint/common elements).

All tests use `reflect.DeepEqual` for precise structural comparison and sort outputs where order is non-deterministic.

## Design Notes

- Immutable operations: methods return *new* sets rather than mutating in-place.
- No duplicate support (standard set semantics).
- Efficient storage via Go maps; no external dependencies.

> **Note**: This is a minimal, idiomatic set implementation—ideal for utility or testing contexts.
