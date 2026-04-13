```markdown
# DNA Scoring and Matching Utilities in `obialign`

This module provides low-level utilities for computing nucleotide alignment scores using probabilistic and bit-encoded representations.

- **Bit Encoding**: Nucleotides are encoded in 4-bit groups (e.g., `A=0b0001`, `C=0b0010`, etc.), enabling efficient bitwise comparison.
- **`_MatchRatio(a, b)`**: Computes a normalized match ratio between two encoded bytes based on shared bits:
  `ratio = common_bits / (bits_in_a × bits_in_b)`.
- **`_FourBitsCount`**: Precomputed lookup table for Hamming weight (popcount) of 4-bit values.
- **Log-space Arithmetic**: Helper functions (`_Logaddexp`, `_Logdiffexp`, `_Log1mexp`) ensure numerical stability in probabilistic computations.
- **Phred-scaled Quality Integration**:
  `_MatchScoreRatio(QF, QR)` derives log-odds match/mismatch scores from Phred quality values (`QF`, `QR`), modeling sequencing error probabilities.
- **Precomputed Matrices**:
  - `_NucPartMatch[i][j]`: Match ratios for all nucleotide pairs (from 4-bit codes).
  - `_NucScorePartMatchMatch/Mismatch[i][j]`: Integer-scaled match/mismatch scores (×10) for quality pairs `(i, j)` in `[0..99]`.
- **Thread-Safe Initialization**: `_InitDNAScoreMatrix()` ensures one-time, synchronized initialization of all scoring tables via a mutex.

Designed for high-performance alignment kernels where speed and numerical robustness are critical.
```
