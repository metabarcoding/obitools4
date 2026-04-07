# `obidist`: Efficient Symmetric Distance and Similarity Matrix Management

The `obidist` Go package provides memory-efficient, symmetric matrix implementations for pairwise **distance** and **similarity** computations — ideal for clustering, phylogenetics, or any domain requiring fast access with minimal footprint. It enforces structural guarantees (symmetry, fixed diagonal) and offers safe, label-aware operations.

## Core Types

| Type | Description |
|------|-------------|
| `DistMatrix` | Symmetric *n×n* matrix for **distances**; diagonal entries are always `0.0`. |
| `SimilarityMatrix` | Symmetric *n×n* matrix for **similarities**; diagonal entries are always `1.0`. |

Both types store only the upper triangle (`i < j`) to reduce memory from *O(n²)* to *n(n−1)/2*. All access (`Get`, `Set`) is automatically mirrored for symmetry.

## Constructors

| Function | Description |
|---------|-------------|
| `NewDistMatrix(n)` / `WithLabels(labels []string)` | Creates a distance matrix of size *n×n* (diag = 0). Labels are optional. |
| `NewSimilarityMatrix(n)` / `WithLabels(labels []string)` | Creates a similarity matrix of size *n×n* (diag = 1). Labels are optional. |

> **Note**: Passing `labels` with length ≠ *n* panics; empty labels (`nil`) are allowed.

## Core Operations

| Method | Description |
|--------|-------------|
| `Get(i, j) float64` | Returns value at *(i,j)*; enforces symmetry (reads stored upper triangle). |
| `Set(i, j, v float64)` | Sets value at *(i,j)*; silently ignores diagonal assignments. |
| `Size() int` | Returns *n*, the matrix dimension. |
| `GetLabel(i) string`, `SetLabel(i int, label string)` | Read/write the *i*-th element’s label. |
| `Labels() []string` | Returns a **copy** of all labels (safe mutation). |
| `GetRow(i) []float64`, `GetColumn(j) []float64` | Returns full row/column as a **new slice** (symmetric copy). |

> All index access panics on out-of-bounds (`i < 0` or `≥ n`). Diagonal writes (e.g., `Set(i, i, v)`) are silently ignored.

## Analysis & Utility Methods

| Method | Description |
|--------|-------------|
| `MinDistance() (val float64, i, j int)` | Returns smallest off-diagonal value and its indices. For *n ≤ 1*, returns `(0, -1, -1)`. |
| `MaxDistance() (val float64, i, j int)` | Returns largest off-diagonal value and its indices. For *n ≤ 1*, returns `(0, -1, -1)`. |
| `Copy() *DistMatrix` | Deep copy (including labels). Safe for concurrent use or immutability. |
| `ToFullMatrix() [][]float64` | Returns a dense *n×n* copy (upper/lower triangles + diagonal). Use sparingly for large matrices. |

## Edge Cases & Guarantees

- **Empty matrix** (*n = 0*): All methods behave safely (e.g., `Size()` → `0`, min/max → `(0, -1, -1)`).
- **Singleton matrix** (*n = 1*): Only diagonal exists → min/max return `(0, -1, -1)`.
- **Label integrity**: `Labels()` and row/column copies use defensive duplication.
- **No normalization enforced** on similarity values (e.g., `[-∞, +∞]` allowed), but diagonals are *always* fixed.

Designed for correctness-first scientific workflows, with rigorous unit tests covering bounds checks and symmetry.
