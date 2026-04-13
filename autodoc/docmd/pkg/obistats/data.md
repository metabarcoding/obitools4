# `obistats` Package Overview

The `obistats` package provides data structures and utilities for analyzing benchmark results in Go. It enables aggregation, statistical summarization, and comparison of performance metrics across multiple configurations.

## Core Types

- **`Collection`**: Holds benchmark results grouped by configuration, group label (e.g., parameter combinations), and metric unit. It tracks:
  - Ordered lists of `Configs`, `Groups`, and `Units`.
  - A map from group names to ordered lists of benchmark functions (`Benchmarks`).
  - `Metrics`, keyed by `(Config, Group, Benchmark, Unit)`.
  - Optional parameters for significance testing (`DeltaTest`, `Alpha`), geometric mean inclusion, and result ordering/splitting.

- **`Key`**: Uniquely identifies a metric for one benchmark run, combining configuration source (`Config`), group label (`Group`), benchmark name (sans `"Benchmark"` prefix), and unit.

- **`Metrics`**: Stores raw (`Values`) and cleaned (`RValues`, with outliers removed via IQR) measurements, plus derived statistics: `Min`, `Mean`, and `Max`.

## Key Functionality

- **Statistical summarization**:  
  - Outlier removal using Tukey’s fences (Q1 ± 1.5×IQR, Q3 + 1.5×IQR).  
  - Computation of min/mean/max over cleaned data.

- **Formatting helpers**:  
  - `FormatMean()`: Returns formatted mean (e.g., scaled or raw).  
  - `FormatDiff()`: Computes and formats symmetric deviation as ±% (based on min/max vs. mean).  
  - `Format()`: Combines both into `"mean ±diff"` style.

- **Dynamic collection building**:  
  - `addMetrics()` creates or retrieves metrics for a given key, while maintaining ordered lists of unique configs/groups/units and benchmarks-per-group.

> ⚠️ *Note*: The file includes commented-out methods (`AddFile`, `AddData`, etc.) referencing an external `benchfmt` package—these are placeholders and not part of the active API in this excerpt.
