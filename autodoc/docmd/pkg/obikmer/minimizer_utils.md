# Minimizer Size Utilities in `obikmer`

This Go package provides helper functions to compute and validate the **minimizer size** `m` in k-mer-based genomic algorithms (e.g., minimizer schemes for sequence comparison or indexing).

## Core Functions

- **`DefaultMinimizerSize(k)`**  
  Returns a *recommended* minimizer size: `ceil(k / 2.5)`, clamped to `[1, k−1]`.  
  → Ensures `m` is reasonably large for uniqueness while keeping window size (`k − m + 1`) manageable.

- **`MinMinimizerSize(nworkers)`**  
  Computes the *minimum* `m` such that there are ≥ `nworkers` distinct minimizers:  
  solves `4^m ≥ n_workers`, i.e., `ceil(log₄(nworkers))`.  
  → Guarantees enough diversity for parallelization (e.g., hashing-based distribution across workers).

- **`ValidateMinimizerSize(m, k, nworkers)`**  
  Enforces constraints on `m`:  
    - Lower bound: ≥ `MinMinimizerSize(nworkers)` (warns & adjusts if violated)  
    - Hard bounds: `1 ≤ m < k`  
  → Prevents invalid or inefficient parameter choices.

## Semantic Purpose

These functions ensure that minimizer-based workflows are:
- **Theoretically sound** (sufficient entropy for parallelism),
- **Practically viable** (avoiding degenerate cases like `m = 0` or `m ≥ k`),
- **User-friendly** (providing sensible defaults + clear warnings on adjustment).
