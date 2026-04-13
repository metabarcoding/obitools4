# `obistats` Package: Statistical Utilities for Weighted and Unweighted Samples

The `obistats` package provides a suite of statistical functions for analyzing numeric samples, supporting both unweighted and weighted data. Its core abstraction is the `Sample` struct—encapsulating values (`Xs`), optional weights (`Weights`), and a `Sorted` flag for performance optimization.

### Key Functionalities:

- **Bounds**: Computes min/max efficiently—O(1) when sorted and unweighted; otherwise scans the data.
- **Aggregation**: `Sum()` computes weighted/unweighted sums via incremental accumulation; `Weight()` returns total weight (or count if unweighted).
- **Central Tendency**:  
  - `Mean()` uses incremental weighted mean for numerical stability.  
  - `GeoMean()` computes geometric means (requires positive values), also supporting weights.
- **Dispersion**:  
  - `Variance()` and `StdDev()` compute sample variance/standard deviation (unweighted only; weighted versions raise a panic—*TODO*).  
  - Based on Welford’s online algorithm for numerical robustness.
- **Order Statistics**:  
  - `Percentile(p)` implements Hyndman & Fan’s R8 interpolation method (default in many tools). Handles weights via linear scan; constant-time if sorted and unweighted.  
  - `IQR()` returns interquartile range (`P75 − P25`).
- **Utility Methods**:  
  - `Sort()` sorts in-place (stably for weighted samples) and updates the `Sorted` flag.  
  - `Copy()` creates a deep copy for independent manipulation.

Designed with performance in mind, the package exploits sorting and incremental algorithms to minimize numerical error and improve runtime—especially valuable for large or repeated analyses. All functions gracefully handle edge cases (empty samples, zero weights) by returning `NaN` or appropriate bounds.
