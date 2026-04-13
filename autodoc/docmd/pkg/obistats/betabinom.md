# Beta-Binomial Distribution Implementation in `obistats`

This Go package provides a complete statistical implementation of the **Beta-Binomial distribution**, a compound discrete probability distribution where the success probability of a Binomial distribution follows a Beta distribution.

## Core Features

- **Struct Definition**:  
  `BetaBinomial` encapsulates the distribution parameters: number of trials (`N > 0`) and Beta shape parameters `Alpha` and `Beta`, both strictly positive. Optional random source (`Src`) supports reproducible sampling.

- **Probability Mass Function (PMF)**:  
  - `LogProb(x)` computes the natural logarithm of the PMF at integer `x ∈ [0, N]`.  
  - `Prob(x)` returns the PMF value via exponentiation.

- **Cumulative Distribution Function (CDF)**:  
  - `LogCDF(x)` evaluates the log-CDF using an analytical expression involving:
    - Log-binomial coefficient (`Lchoose`)
    - Log-beta function (`mathext.Lbeta`)  
    - Generalized hypergeometric function `HypPFQ` (via `scientificgo.org/special`).  
  - `CDF(x)` returns the standard CDF as `exp(LogCDF(x))`.  

- **Statistical Moments**:
  - Mean: $N \cdot \frac{\alpha}{\alpha + \beta}$
  - Variance: $N \cdot \frac{\alpha \beta (\!N + \alpha + \beta\!)}{(\alpha+\beta)^2 (\alpha+\beta+1)}$
  - Standard deviation: square root of variance.

- **Mode**:  
  Returns the most probable count. Special cases handled:
    - `NaN` if both $\alpha, \beta \leq 1$
    - $0$ if only $\alpha \leq 1$
    - $N$ if only $\beta \leq 1$

- **Utility Methods**:
  - `LogCDFTable(x)` builds a cumulative log-probability table up to `x`, useful for fast lookup or numerical stability.
  - `NumParameters()` returns the number of distribution parameters (3: $N$, $\alpha$, $\beta$).

- **Input Validation**:  
  Panics on invalid parameters (non-positive `N`, $\alpha$, or $\beta$), ensuring correctness.

This module supports high-precision statistical computations using specialized mathematical libraries (`gonum.org/v1/gonum/mathext`, `scientificgo.org/special`).
