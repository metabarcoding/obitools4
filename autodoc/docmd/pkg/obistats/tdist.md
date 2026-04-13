# `obistats.TDist`: Student's *t*-Distribution Implementation

This Go package provides a lightweight implementation of the **Student’s *t*-distribution**, commonly used in statistical inference (e.g., hypothesis testing, confidence intervals) when sample sizes are small or population variance is unknown.

## Core Components

- **`TDist` struct**:  
  Represents a *t*-distribution parameterized by degrees of freedom `V`.

- **`PDF(x)` method**:  
  Computes the *probability density function* at point `x`, using:
  $$
    f(x) = \frac{\Gamma\left(\frac{V+1}{2}\right)}{\sqrt{V\pi} \, \Gamma\left(\frac{V}{2}\right)} 
           \left(1 + \frac{x^2}{V} \right)^{-\frac{V+1}{2}}
  $$
  Leverages `lgamma` for numerical stability in Gamma function evaluation.

- **`CDF(x)` method**:  
  Computes the *cumulative distribution function*:
  - Returns `0.5` at symmetry point (`x == 0`);
  - Uses the **regularized incomplete beta function** `mathBetaInc` for `x > 0`;
  - Exploits symmetry: `CDF(-x) = 1 − CDF(x)` for `x < 0`.

- **`Bounds()` method**:  
  Returns a practical truncation interval `[-4, 4]`, sufficient for most visualizations or numerical integration over the central mass of the distribution.

## Dependencies & Notes

- Relies on standard library `math` and custom/internal helpers (`lgamma`, `mathBetaInc`) — likely from a shared internal module.
- Designed for performance and numerical robustness, suitable in statistical tooling or benchmark analysis (as suggested by the `obistats` package name and reference to a bench-related repo).
