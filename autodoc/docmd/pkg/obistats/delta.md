# Semantic Description of `obistats` Delta Testing Functionality

This Go package (`obistats`) provides statistical tools for comparing performance metrics before and after code changes—typically used in benchmarking workflows.

- **`DeltaTest` type**: A function signature for comparing two `*Metrics` instances (old vs. new), returning a *p*-value (`float64`) and an optional error.
- **Purpose**: Determine whether two sets of samples likely originate from the same underlying distribution (i.e., detect significant performance regressions/improvements).

## Supported Tests

- **`NoDeltaTest()`**: A no-op test returning `(-1, nil)`, indicating *no statistical comparison* is performed.
- **`TTest()`**: Performs a two-sample Welch’s *t*-test on `RValues`, assessing whether means differ significantly.
- **`UTest()`**: Applies the Mann–Whitney *U* test (non-parametric), comparing distributions without assuming normality.

## Common Errors

- `ErrSamplesEqual`: All samples in one or both groups are identical.
- `ErrSampleSize`: Insufficient data points for reliable testing (e.g., < 2).
- `ErrZeroVariance`: One sample set has zero variance (no spread), breaking test assumptions.
- `ErrMismatchedSamples`: Sample lengths differ (not used here but part of the broader API).

## Design Rationale

- Built on top of internal benchmarking infrastructure (see `github.com/golang-design/bench`).
- Designed for modularity: callers can plug in different statistical tests as needed.
- Returns *p*-values directly, enabling threshold-based decision logic (e.g., `if p < 0.05 → alert`).
