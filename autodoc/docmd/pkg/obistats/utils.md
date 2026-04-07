# Semantic Description of `obistats` Package

The `obistats` package provides numerically stable statistical utilities for combinatorics and log-space arithmetic, primarily intended for use in bioinformatics or probabilistic modeling.

- **`Lchoose(n, x int) float64`**:  
  Computes the natural logarithm of the binomial coefficient "n choose x" using the log-gamma function (`math.Lgamma`). This avoids overflow/underflow inherent in direct computation of large factorials.

- **`Choose(n, x int) float64`**:  
  Returns the (floating-point approximation of the) binomial coefficient by exponentiating `Lchoose`. *Note*: The argument order in the implementation (`math.Exp(Lchoose(x,n))`) appears reversed—likely a typo; should be `Lchoose(n,x)`.

- **`LogAddExp(x, y float64) float64`**:  
  Computes `log(exp(x) + exp(y))` in a numerically stable way. Uses the identity:  
  `log(eˣ + eʸ) = max(x, y) + log(1 + exp(-|x - y|))`, implemented via `math.Log1p` for precision near zero.  
  Handles NaNs/infinities with logging and fallback.

All functions rely on `math` for core operations, and use Logrus (`log.Errorf`) to warn about invalid inputs (e.g., non-finite values).

Use cases include:  
- Exact p-value computation in overrepresentation tests (e.g., hypergeometric),  
- Log-probability accumulation in hidden Markov models or Bayesian networks,  
- Stable mixture model likelihood evaluations.
