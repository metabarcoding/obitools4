# `obistats` Package: Mann-Whitney U-test Implementation

The `obistats` package provides a **non-parametric statistical test** for comparing two independent samples: the **Mann–Whitney U-test**, also known as the Wilcoxon rank-sum test.

## Core Functionality

- **`MannWhitneyUTest(x1, x2 []float64, alt LocationHypothesis)`**  
  Performs the test between two samples `x1` and `x2`, under a user-specified alternative hypothesis (`LocationLess`, `LocationDiffers`, or `LocationGreater`).

- Returns a structured result:  
  - Sample sizes (`N1`, `N2`)  
  - U statistic (with tie handling: ties contribute 0.5)  
  - Alternative hypothesis used (`AltHypothesis`)  
  - Achieved *p*-value (`P`)

## Key Features

- **Non-parametric**: No assumption of normality — suitable for ordinal data or non-Gaussian distributions.
- **Exact vs Approximate**:  
  - Uses *exact U distribution* for small samples (≤50 without ties, ≤25 with ties).  
  - Falls back to *normal approximation* for larger samples (with tie and continuity corrections).
- **Tie Handling**:  
  - Ranks averaged for tied values.  
  - Tie correction applied in variance estimation.
- **Error Handling**: Returns `ErrSampleSize` (empty input) or `ErrSamplesEqual` (all values identical).

## Implementation Notes

- Uses labeled merge to interleave sorted samples while preserving origin labels.
- Computes U via rank sums: `U1 = R1 − n₁(n₁+1)/2`.
- Supports one-tailed and two-tailed tests.
- Includes helper functions: `labeledMerge`, `tieCorrection`.

## References

Mann & Whitney (1947); Klotz (1966).  
Efficiency slightly lower than *t*-test on normal data, but more robust to outliers and distributional assumptions.
