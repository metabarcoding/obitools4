# Mann-Whitney U Distribution Implementation in `obistats`

The `obistats` package provides efficient computation of the **Mann-Whitney U distribution**, used in nonparametric hypothesis testing to compare two independent samples.

## Core Types

- **`UDist`**: Represents the discrete probability distribution of the U statistic for sample sizes `N1`, `N2`. It optionally handles **ties** via a tie-count vector `T`.

## Key Features

- ✅ **Exact distribution computation**, both with and without ties.
  - *No ties*: Uses dynamic programming (Mann–Whitney recurrence) in `O(N1·N2·U)` time.
  - *With ties*: Implements the linked-list-based algorithm from Cheung & Klotz (1997) via memoization (`makeUmemo`).

- ✅ **PMF & CDF evaluation**:
  - `PMF(U)` returns the probability mass at U.
  - `CDF(U)` computes cumulative probabilities using symmetry to minimize computation.

- ✅ **Support for tied ranks**:
  - `T` encodes tie multiplicities per rank; if nil, no ties are assumed.

- ✅ **Optimized recurrence**:
  - Exploits symmetry (`p_{n,m} = p_{m,n}`) and incremental DP to reduce memory/time.

- ✅ **Boundary handling**:
  - `Bounds()` returns support `[0, N1·N2]`.
  - `Step() = 0.5`, reflecting U’s discrete unit in tied cases.

## Algorithm Notes

- `p(U)` uses a 2D DP table (rows = *n*, columns = U), computing only necessary states.
- `makeUmemo` builds a 3D memoization table (`k`, `n1`, `2U`) for tied distributions.
- Performance bottlenecks noted in comments (e.g., map overhead) suggest future optimization paths.

## Use Case

Enables exact *p*-value calculation for the **Mann-Whitney U test**, especially valuable when:
  - Sample sizes are small-to-moderate (exact methods needed).
  - Data contain ties.
