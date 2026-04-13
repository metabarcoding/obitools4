# `obiutils` Package Overview

This Go package provides utility functions for common data conversion, serialization, and reflection tasks.

- **Custom Error Types**: Defines typed errors (`NotAnInteger`, `NotAFloat64`, etc.) for precise type validation failures.
- **Interface-to-Type Casting**: Offers robust conversion functions:
  - `InterfaceToFloat64Map`, `InterfaceToIntSlice`, etc., handling nested interfaces and type coercion (e.g. `int` → `float64`, slices of `interface{}`).
- **File I/O**: `ReadLines` reads a file line-by-line into a string slice, handling buffered reading efficiently.
- **Concurrency**: `AtomicCounter` returns an incrementing integer generator—thread-safe via mutex, optionally starting from a given value.
- **JSON Serialization**: `JsonMarshal` and `JsonMarshalByteBuffer` provide UTF‑8–preserving JSON encoding (avoids Go’s default HTML escaping).
- **Reflection Helpers**:
  - `IsAMap`, `IsASlice`, `IsAnArray` detect container types.
  - `HasLength`, `Len`, and `IsAContainer` abstract length operations across maps, slices, arrays, or custom types with a `.Len()` method.
- **Deep Copying**: `MustFillMap` performs deep copying of nested structures using `go-deepcopy`.

All functions prioritize safety, type correctness, and usability in data-heavy or concurrent applications.
