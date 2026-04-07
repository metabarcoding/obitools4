# `obistats` Package — Semantic Overview

The `*obistats*` Go package provides lightweight, type-generic statistical utilities for numerical data.

## Core Functions

- **`Median[T Number](data []T) float64`**  
  Computes the median of a slice. Internally copies and sorts input data to avoid mutation, handling both even- and odd-length slices correctly. Returns `0` for empty input.

- **`Mean[T Number](data []T) float64`**  
  Calculates the arithmetic mean by summing all elements (converted to `float64`) and dividing by count.

## Type Constraints

- Uses Go generics (`constraints.Float | constraints.Integer`), enabling use with `int`, `float32`, `float64`, etc.

## Design Notes

- Non-mutating (`Median` works on a copy).
- Simple, efficient implementations—no external dependencies beyond `golang.org/x/exp/constraints` and `slices`.
- Focused on central tendency measures only—no variance, std dev, or distribution stats.

## Use Case

Ideal for small-to-medium numerical datasets where minimal dependencies and clarity are prioritized over advanced statistics.
