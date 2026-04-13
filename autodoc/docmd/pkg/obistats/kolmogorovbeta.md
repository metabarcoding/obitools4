# `BetaKolmogorovDist` Function — Semantic Description

The `obistats.BetaKolmogorovDist` function computes a **goodness-of-fit statistic** between an empirical dataset and the *cumulative distribution* (CDF) of a **Beta probability distribution** with specified parameters `α` and `β`. It implements an adapted version of the **Kolmogorov–Smirnov (KS) test**, tailored for Beta-distributed theoretical models.

### Key Functionalities:
- **Input**:  
  - `data []float64`: Empirical sample (assumed sorted if `preordered = true`).  
  - `alpha`, `beta float64`: Shape parameters of the target Beta distribution.  
- **Processing**:  
  - If not pre-sorted, data is copied and sorted ascendingly.  
  - For each ordered sample point `v_i`, it accumulates the sum `s = Σ_{j≤i} v_j`.  
  - Evaluates:  
    `|CDF_Beta(s; α, β) − empirical CDF_i|`, where the *empirical* cumulative probability at rank `i` is approximated as `1/(i+1)` — a common Bayesian/maximum-likelihood estimator (e.g., median-rank).  
  - Returns the **supremum** of these absolute deviations (i.e., max distance across all points).  

### Interpretation:
- A **small value** indicates the empirical cumulative sums align closely with the theoretical Beta CDF.  
- A **large value** suggests significant deviation — poor fit of aBeta(α,β) to the data.  
- Unlike standard KS tests (which use `i/n`), this uses `1/(i+1)` — suitable for small samples or Bayesian contexts.

### Dependencies:
- Uses `gonum.org/v1/gonum/stat/distuv.Beta` for CDF computation.  
- Uses `gonum.org/v1/gonum/floats.Max` for distance extremal computation.  
- `sort.Float64s` ensures ordered traversal.

> **Note**: The use of *cumulative sums* (`s`) rather than raw values is unconventional — possibly intended for data representing proportions or waiting times where the *integral* of observations matters.
