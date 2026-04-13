# Statistical Functions in `obistats` Package

This Go package provides high-precision statistical functions for probability distributions, particularly the **regularized incomplete beta function**, used in hypothesis testing and confidence interval calculations.

## Core Functions

- **`mathBeta(a, b)`**  
  Computes the *complete beta function* $ B(a,b) = \frac{\Gamma(a)\Gamma(b)}{\Gamma(a+b)} $ using logarithms of the gamma function (`math.Lgamma`) for numerical stability.

- **`lgamma(x)`**  
  Wrapper around `math.Lgamma`, returning the natural logarithm of the absolute value of the gamma function.

- **`mathBetaInc(x, a, b)`**  
  Computes the *regularized incomplete beta function* $ I_x(a,b) $. This is essential for computing cumulative distribution functions (CDFs) of the beta, F-, and t-distributions.  
  - Uses *continued fraction evaluation* (via `betacf`) for accuracy.
  - Applies symmetry transformation ($ x \to 1-x $) when beneficial (per Numerical Recipes).
  - Returns `NaN` for invalid inputs (`x < 0 || x > 1`).

- **`betacf(x, a, b)`**  
  Implements the continued fraction expansion of $ I_x(a,b) $.  
  - Iteratively evaluates recurrence relations for even/odd terms.
  - Uses `epsilon = 3e-14` and `maxIterations = 200` for convergence.
  - Handles near-zero denominators via `raiseZero`.

## Use Cases

- Statistical hypothesis testing (e.g., Fisher’s exact test).
- Beta, binomial proportion confidence intervals.
- F-test and Student's t-distribution CDF computations.

## Implementation Notes

Based on *Numerical Recipes in C*, §6.4, with robustness enhancements for floating-point edge cases.
