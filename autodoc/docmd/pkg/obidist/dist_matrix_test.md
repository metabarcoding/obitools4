# `obidist` Package: Semantic Feature Overview

The `obidist` Go package provides two core data structures for managing **distance** and **similarity matrices**, with built-in guarantees suitable for scientific computing (e.g., clustering, phylogenetics). Key features include:

- **`DistMatrix`**: A symmetric `n×n` matrix representing pairwise distances, where:
  - Diagonal entries are *always* `0.0` (self-distance).
  - Off-diagonals obey symmetry: `dist(i, j) == dist(j, i)`.
  - Automatic enforcement via dedicated `Set()`/`Get()` methods.
  
- **`SimilarityMatrix`**: A symmetric matrix where:
  - Diagonal entries are *always* `1.0`.
  - Off-diagonals represent similarity scores (e.g., between `0` and `1`, though not enforced).
  - Symmetry is similarly guaranteed.

Both matrix types support:
- **Optional labels**: Associate human-readable identifiers (e.g., sample names) with rows/columns.
- **Safe bounds checking**: Panics on out-of-range access (tested via `defer/recover`).
- **Deep copy support**: Ensures isolation between original and copied instances.
- **Utility methods**:
  - `MinDistance()` / `MaxDistance()`: Return extremal values and their indices.
  - `GetRow(i)`: Retrieve a full row as a slice (symmetric copy).
  - `ToFullMatrix()`: Export the matrix as an immutable 2D slice.

Edge cases are rigorously handled:
- Empty (`n=0`) and singleton (`n=1`) matrices return `(0.0, -1, -1)` for min/max.
- Label mutations do not affect internal state via defensive copying.

All behaviors are validated through comprehensive unit tests, emphasizing correctness and robustness.
