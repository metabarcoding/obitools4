# `obistats` Package: Normal Distribution Utilities

The `obistats` package provides a lightweight, efficient implementation of the **normal (Gaussian) distribution**, including core statistical operations.

## Core Type
- `NormalDist`: Represents a normal distribution with parameters:
  - `Mu` (mean)
  - `Sigma` (standard deviation)

## Predefined Constants
- `StdNormal`: A standard normal distribution (`Mu = 0`, `Sigma = 1`).
- `invSqrt2Pi`: Precomputed constant for performance optimization.

## Key Methods
| Method | Description |
|--------|-------------|
| `PDF(x)` | Computes the **probability density function** at point `x`. |
| `pdfEach(xs [])` | Vectorized PDF evaluation over a slice of values (optimized for standard normal). |
| `CDF(x)` | Computes the **cumulative distribution function** at point `x` via error function (`erfc`). |
| `cdfEach(xs [])` | Vectorized CDF evaluation over a slice. |
| `InvCDF(p)` | Computes the **inverse CDF (quantile function)** using Acklam’s algorithm with refinement. Handles edge cases (`p = 0`, `1`) and numerical stability. |
| `Rand(r *rand.Rand)` | Generates a random sample from the distribution (uses Go’s built-in `NormFloat64`). |
| `Bounds()` | Returns a practical support interval: `[Mu − 3·Sigma, Mu + 3·Sigma]` (≈99.7% coverage). |

## Implementation Notes
- Optimized paths for standard normal (`Mu = 0`, `Sigma = 1`) reduce computation cost.
- Uses Go’s standard math library (`math.Erfc`, `math.Log`, etc.).
- Designed for performance and numerical accuracy in statistical applications.

> *Note: Duplicates functionality from an internal module (`bench`), likely for reuse in public packages.*
