# `obistats` Package: Semantic Description

The `obistats` package provides utility functions for **formatting and scaling benchmark measurements** in Go, especially tailored for performance benchmarks (e.g., `go test -bench`). Its core component is the **`Scaler` type**, a function that converts raw numeric values into human-readable, unit-aware strings.

- **`Scaler func(float64) string`**: A function type that formats a numeric measurement (e.g., time, memory usage, throughput) into an appropriately scaled and unit-annotated string.

- **`NewScaler(val float64, unit string) Scaler`**: Dynamically selects the best scaling strategy based on:
  - The measurement value (`val`)
  - Its unit (e.g., `"ns/op"`, `"MB/s"`, `"B/op"`)

  It applies **SI prefixes** (`k`/`M`/`G`/`T`) with adaptive precision (0–2 decimal places) to ensure readability and consistency across table rows.

- **`timeScaler(ns float64)`**: Specialized scaler for time-based units (`ns/op`, `ns/GC`). It selects optimal unit (s, ms, µs, ns) and precision based on magnitude.

- **`hasBaseUnit(s, unit string) bool`**: Helper to detect if a full unit string (e.g., `"bytes/op"`, `"MB/s"`) includes or matches a base unit.

Key features:
- Supports common Go benchmark units: time (`ns/op`), memory (`B/op`, `bytes/op`), throughput (`MB/s`)
- Ensures consistent formatting across rows (e.g., all values in a row use same scale)
- Avoids unnecessary trailing zeros and uses SI conventions
- Designed for compatibility with internal benchmarking infrastructure (originally from `golang-design/bench`)

Intended use: formatting tables of benchmark results where readability and unit consistency are critical.
