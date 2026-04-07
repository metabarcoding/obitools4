# Semantic Description of `obialign` Package

This Go package provides core utilities for **DNA sequence alignment scoring**, leveraging probabilistic models and log-space computations to ensure numerical stability.

## Key Functionalities

- **Four-bit nucleotide encoding**: Uses `_FourBitsBaseCode` (implied but not shown) to encode DNA bases as 4-bit values, enabling bitwise operations for fast comparison.

- **Bitwise match ratio (`_MatchRatio`)**: Computes a normalized overlap score between two encoded bases by counting shared bits, adjusting for presence/absence in each operand.

- **Log-space arithmetic helpers**:
  - `_Logaddexp`: Stable computation of `log(exp(a) + exp(b))`.
  - `_Log1mexp`, `_Logdiffexp`: Accurate log-domain operations for `log(1 − exp(a))` and `log(exp(a) − exp(b))`, critical for probability transformations.

- **Match/mismatch scoring (`_MatchScoreRatio`)**:
  - Derives log-probability-based scores for observed matches/mismatches using Phred-quality inputs (`QF`, `QR`).
  - Incorporates base composition priors (e.g., uniform 4-mer assumption via `log(3)`, `log(4)`).

- **Precomputed scoring matrices**:
  - `_NucPartMatch`: Precomputes match ratios for all base-pair combinations.
  - `_NucScorePartMatch{Match,Mismatch}`: Stores integer-scaled alignment scores (×10) for all Phred-quality pairs, enabling fast lookup during dynamic programming.

- **Thread-safe initialization**:
  - `_InitDNAScoreMatrix` ensures one-time setup of all matrices using a mutex guard, preventing race conditions.

All computations are designed for high performance and numerical robustness in large-scale sequence alignment tasks.
