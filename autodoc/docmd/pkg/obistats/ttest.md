# Statistical Hypothesis Testing Module (`obistats`)

This Go package provides implementations of common **t-tests** for comparing sample means under different assumptions. It supports one- and two-sample tests, paired or unpaired designs.

## Core Types

- **`TTestResult`**: Encapsulates the outcome of a t-test, including:
  - Sample sizes (`N1`, `N2`)
  - Test statistic value (`T`)
  - Degrees of freedom (`DoF`)
  - Alternative hypothesis type (`AltHypothesis`: `LocationDiffers`, `LocationLess`, or `LocationGreater`)
  - Computed *p*-value (`P`)

- **`TTestSample` interface**: Requires methods `Weight()`, `Mean()`, and `Variance()` — enabling reuse with summary statistics.

## Supported Tests

1. **`TwoSampleTTest(x1, x2)`**  
   Standard Student’s *t*-test for two independent samples assuming **equal variances** and normality.

2. **`TwoSampleWelchTTest(x1, x2)`**  
   Welch’s *t*-test for two independent samples **without equal-variance assumption**, using Satterthwaite approximation for degrees of freedom.

3. **`PairedTTest(x1, x2)`**  
   Paired *t*-test for dependent samples (e.g., before/after), testing mean of differences against `μ0`.

4. **`OneSampleTTest(x)`**  
   One-sample *t*-test comparing sample mean to a known population mean `μ0`.

## Error Handling

- Returns errors for invalid inputs: zero sample size (`ErrSampleSize`), zero variance (`ErrZeroVariance`), or mismatched paired sample lengths (`ErrMismatchedSamples`).

## Implementation Notes

- *p*-values are computed using the cumulative distribution function (CDF) of the Student’s *t*-distribution.
- Designed for statistical rigor and modularity, reusing internal utilities (e.g., `Mean`, `StdDev`) from a shared module.
