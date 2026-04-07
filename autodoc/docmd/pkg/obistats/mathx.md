# `obistats` Package: Semantic Overview

The `obistats` package provides low-level statistical and combinatorial utilities in pure Go, focusing on numerical robustness and performance.

- **Sign Function (`mathSign`)**  
  Returns the sign of a `float64`: `-1`, `0`, or `+1`. Handles NaN by returning NaN.

- **Precomputed Factorials (`smallFact`)**  
  Precomputes factorials from `0!` to `20!` (fits in 64-bit signed integer), enabling fast exact binomial coefficient computation for small `n`.

- **Binomial Coefficient (`mathChoose`)**  
  Computes $\binom{n}{k}$ efficiently:
  - For `n ≤ 20`: uses integer arithmetic (multiplication + division) for exact results.
  - For larger `n`: leverages logarithms via `mathLchoose` and exponentiates (`exp(log(Choose))`) to avoid overflow.

- **Log-Binomial Coefficient (`mathLchoose`)**  
  Computes $\log \binom{n}{k}$ via the log-gamma function:  
  $$\log \binom{n}{k} = \lg(n+1) - \lg(k+1) - \lg(n-k+1)$$  
  Ensures numerical stability for large `n`, avoiding overflow/underflow.

- **Internal Helper (`lchoose`)**  
  Core implementation of log-binomial using `math.Lgamma`, reused by both exact and large-scale paths.

**Design Notes**:  
- Prioritizes correctness (e.g., NaN propagation, edge-case handling).  
- Balances speed and precision: exact integer arithmetic for small inputs; log-space computation for scalability.  
- Mirrors functionality from an internal benchmarking module, adapted here as a standalone utility.
