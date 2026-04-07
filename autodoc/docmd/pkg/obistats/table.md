# `obistats` Package: Benchmark Statistics and Comparison

The `obistats` package provides semantic tools to analyze, compare, and display benchmark results—typically from Go’s `testing.B` benchmarks. It enables structured reporting of performance changes across configurations (e.g., before/after code modifications).

### Core Concepts
- **`Collection`**: Aggregates benchmark metrics across groups, benchmarks, and configurations.
- **`Table` & `Row`**: Represent formatted tabular output for human-readable comparison (e.g., in CLI tools like `benchstat`).
- **Metrics per row**: Include mean, variance, sample size (`n`), and statistical test results.

### Key Functionalities
- **Statistical summarization**: Computes means, variances, and other stats via `computeStats()`.
- **Delta comparison** (2-config mode):
  - Performs statistical tests (`UTest` by default) to assess significance.
  - Calculates percent change: `((new/old) − 1) × 100%`.
  - Marks improvements (`+1`) or regressions (`−1`), respecting metric semantics (e.g., lower time/op is better; higher MB/s is better).
- **Handling edge cases**:
  - Skips rows with missing data (e.g., one config absent).
  - Notes issues: zero variance, insufficient samples, or identical values.
- **Geometric mean aggregation**:
  - Adds a `[Geo mean]` row summarizing overall performance across benchmarks.
  - Excludes zero-mean entries to avoid distortion (e.g., allocations of `0`).
- **Metric normalization**:
  - Maps raw units (`ns/op`, `B/op`) to semantic names (e.g., `"time/op"`, `"alloc/op"`).
  - Supports prefixed units (`foo-ns/op` → `foo-time/op`).

### Output Customization
- Supports sorting via user-defined order (`c.Order`).
- Configurable significance level `α` (default: 0.05) for p-value filtering.
- Optional geomean inclusion (`c.AddGeoMean`).

Designed for integration into benchmark analysis pipelines (e.g., CLI tools), `obistats` focuses on **semantic clarity**, **statistical rigor**, and **actionable insights**.
