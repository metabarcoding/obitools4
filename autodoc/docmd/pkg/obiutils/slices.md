# `obiutils` Package Overview

The `obiutils` package provides generic, reusable utility functions for common slice operations in Go.

- **`Contains[T comparable](arr []T, x T) bool`**  
  Checks whether a given element exists in the slice. Uses generic type `T`, requiring only that it supports equality comparison.

- **`LookFor[T comparable](arr []T, x T) int`**  
  Returns the index of the *first* occurrence of `x`, or `-1` if not found. Also generic over comparable types.

- **`RemoveIndex[T comparable](s []T, index int) []T`**  
  Removes the element at `index`, returning a new slice. Works in O(1) time (amortized), using `append` to rebuild the slice.

- **`Reverse[S ~[]E, E any](s S, inplace bool) S`**  
  Reverses the slice elements. If `inplace = true`, modifies the original; otherwise, copies first and returns a reversed copy. Uses type constraint `~[]E` for flexibility across slice aliases.

All functions are designed to be:
- Type-safe via Go generics (no reflection),
- Efficient and idiomatic,
- Well-documented with clear parameter/return semantics.

Ideal for use in data processing, validation logic, or general-purpose slice manipulation.
