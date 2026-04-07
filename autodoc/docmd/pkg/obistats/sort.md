# `obistats` Package: Semantic Overview

This Go package provides utilities for sorting benchmark result tables, derived from an internal module. It focuses on semantic ordering of performance data.

## Core Concepts

- **`Order` type**: A function signature defining custom sort logic for table rows (`func(t *Table, i, j int) bool`).
- **Predefined orders**:
  - `ByName`: Sorts rows alphabetically by benchmark name.
  - `ByDelta`: Orders rows based on magnitude of percentage change (`PctDelta`), adjusted by directionality via `Change`.
- **Helper functions**:
  - `Reverse(order Order)`: Returns a new order that inverts the comparison result.
- **Core utility**:
  - `Sort(t *Table, order Order)`: Performs an in-place stable sort of table rows using the provided ordering function.

## Design Intent

- Enables flexible, domain-aware sorting (e.g., by performance delta or name).
- Supports both ascending and descending sorts via `Reverse`.
- Uses stable sorting (`sort.SliceStable`) to preserve relative order of equal elements.

## Use Case

Ideal for benchmark comparison tools where users need intuitive, configurable table layouts—especially when analyzing performance regressions or improvements.
