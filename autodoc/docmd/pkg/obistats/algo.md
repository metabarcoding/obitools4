# `obistats` Package — Semantic Overview

The `obistats` package provides lightweight, general-purpose numerical utilities in Go. It includes:

- **Basic arithmetic helpers**:  
  - `maxint`, `minint`: return the maximum/minimum of two integers.  
  - `sumint(xs []int) int`: computes the sum over a slice of integers.

- **Root-finding via bisection**:  
  - `bisect(...)`: numerically finds a root of a real-valued function within `[low, high]`, using the classical bisection method. Returns `(root, success)`.  
  - Requires `f(low)` and `f(high)` to have opposite signs; panics otherwise.

- **Boolean bisection**:  
  - `bisectBool(...)`: locates the transition point where a boolean function flips (e.g., threshold detection). Returns adjacent points `(x1, x2)` straddling the change. Panics if `f(low) == f(high)`.

- **Series summation**:  
  - `series(...)`: computes the infinite sum ∑ₙ₌₀^∞ f(n) by iterating until convergence (i.e., `y == yp` within floating-point precision).  
  - *Note*: Fast but may suffer from rounding errors for slowly converging or oscillating series.

All functions are designed for performance and simplicity, with no external dependencies beyond `fmt` (for error messages). The package is a stripped-down copy of internal utilities, likely used in performance-critical or statistical computations.
