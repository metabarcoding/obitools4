Here's a **semantically structured Markdown overview** (≤200 lines) of the `obistats` package, documenting **only public-facing functionality**, written in English and grouped by thematic modules:

```markdown
# `obistats` Package — Public API Overview

The `obistats` package delivers lightweight, numerically robust statistical and combinatorial utilities for Go. Designed for performance-critical applications (e.g., benchmarking, bioinformatics), it avoids external dependencies beyond core math libraries and focuses on **accuracy**, **type safety**, and **modularity**.

---

## 🧮 Numerical Utilities

- `maxint`, `minint`: Return the maximum/minimum of two integers.  
- `sumint(xs []int) int`: Computes sum over integer slice.  
- `bisect(f, low, high float64) (root, success bool)`: Finds root via bisection; requires `f(low)*f(high)<0`.  
- `bisectBool(f, low, high int) (x1, x2 int)`: Locates boolean transition point; panics if `f(low)==f(high)`.  
- `series(f func(int) float64) (sum, converged bool)`: Infinite sum via convergence detection.

---

## 📊 Descriptive Statistics

- `Max[T constraints.Float|constraints.Integer](data []T) T`: Max value in slice (signed ints/floats).  
- `Min[T ...]`: Min over all integer types.  
- `Mode[int...]`: Most frequent value in signed int slice (map-based).  

---

## 📐 Central Tendency & Dispersion

- `Median[T Number](data []T) float64`: Non-mutating median (copy + sort).  
- `Mean[T Number](data []T) float64`: Arithmetic mean.  

---

## 📈 Weighted & Unweighted Samples

- `Sample` struct: Encapsulates values, optional weights (`Weights []float64`), and `Sorted bool`.  
- Methods:  
  - `Mean()`, `GeoMean()` (weighted), `Sum()`, `Weight()`  
  - `Variance()`, `StdDev()` (unweighted only) via Welford’s algorithm  
  - `Percentile(p float64)` (Hyndman–Fan R8), `IQR()`  
  - Bounds: min/max (`O(1)` if sorted & unweighted)  

---

## 📉 Probability Distributions

- **Beta-Binomial**:  
  - `LogProb(x)`, ` Prob(x)` (PMF),  
  - `CDF()`/`LogCDF()`: Analytical via hypergeometric (`HypPFQ`) + log-beta.  
  - Moments: mean, variance; mode with edge-case handling.

- **Normal**: `Mu`, `Sigma`; methods:  
  - PDF/CDF/InvCDF (Acklam’s algorithm), Rand(), `Bounds()`.

- **Student *t***:  
  - PDF/CDF via log-gamma & regularized incomplete beta (`mathBetaInc`).  

- **Kolmogorov–Smirnov for Beta**:  
  - `BetaKolmogorovDist(data []float64, α β float64)`: Max deviation between empirical CDF of cumulative sums and theoretical Beta CDF (uses `1/(i+1)` estimator).

---

## 🧪 Statistical Tests

- **Two-sample tests**:  
  - `TTest()`: Welch’s *t*-test (unequal variances).  
  - `UTest()`: Mann–Whitney *U* (non-parametric; handles ties, exact for small samples).  
- **One-sample/paired**: `TwoSampleTTest`, `PairedTTest`, `OneSampleTTest`.  
- All return structured result (`P` p-value, sample sizes, alt. hypothesis).  

---

## 📦 Nonparametric Distribution Helpers

- **Mann–Whitney U distribution (`UDist`)**:  
  - Exact PMF/CDF via DP (no ties) or Cheung–Klotz algorithm (with ties).  
  - `PMF(U)`, `CDF(U)`; supports tie multiplicities.

---

## 🔢 Combinatorics & Log-Space Arithmetic

- `Lchoose(n, x int) float64`: log-binomial coefficient via `math.Lgamma`.  
- `Choose(n, x int) float64`: exponentiated log-binomial (note: arg order in impl may be reversed).  
- `LogAddExp(x, y float64)`: Stable `log(eˣ + eʸ)` using max+`Log1p`.

---

## 🧩 Random Sampling

- `SampleIntWithoutReplacement(n, max int) []int`:  
  - Uniform sampling *without replacement* in O(*n*) time/memory.  
  - Uses reservoir-style mapping with swap trick for uniqueness.

---

## 📊 Benchmark Analysis & Formatting

- **`Collection`, `Table`, `Row`**: Structured aggregation and display of benchmark metrics.  
- **Metrics processing**:
  - Outlier removal (Tukey’s fences), min/mean/max.  
- **Delta comparison**:
  - `FormatDiff()`: Symmetric ±% deviation; semantic direction (`+1`/`−1`).  
  - `GeoMean()` row for overall summary.  
- **Formatting**:
  - `Scaler`: SI-scaled unit-aware formatting (e.g., `"1.23 ms/op"`).  
  - `timeScaler`, unit detection (`hasBaseUnit`).  

---

## 🧭 Sorting & Ordering

- `Order` type: Custom row sort function.  
  - Predefined: `ByName`, `ByDelta`.  
- `Sort(t *Table, order Order)`: Stable sort via `sort.SliceStable`.

---

## 🧰 Utility Helpers

- `mathSign(x float64)`: Sign function (`NaN` → `NaN`).  
- Precomputed factorials: `smallFact[0..20]`.  

> ⚠️ *All functions assume valid inputs; invalid parameters (e.g., `N≤0`, α/β ≤ 0) may panic.*  
> ✅ *No mutability of input slices unless explicitly stated (e.g., `Sample.Sort()`).*  
