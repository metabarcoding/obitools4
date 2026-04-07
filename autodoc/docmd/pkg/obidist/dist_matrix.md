# `obidist` Package: Efficient Symmetric Distance/Similarity Matrix Management

The `*DistMatrix` type provides a memory-efficient, symmetric matrix implementation for distance or similarity data.

- **Storage Strategy**: Only the upper triangle (i < j) is stored, reducing memory usage from *O(n²)* to *n(n−1)/2*.
- **Diagonal Handling**: Diagonal entries are fixed (0.0 for distances, 1.0 for similarities); assignments to diagonal indices are silently ignored.
- **Symmetry Guarantee**: `Get(i, j)` and `Set(i, j, v)` automatically handle both (i,j) and (j,i), ensuring consistency.

## Constructors

| Function | Description |
|---------|-------------|
| `NewDistMatrix(n)` / `WithLabels(labels)` | Creates *n×n* distance matrix (diag = 0). |
| `NewSimilarityMatrix(n)` / `WithLabels(labels)` | Creates *n×n* similarity matrix (diag = 1). |

## Core Operations

- `Get(i, j)` / `Set(i, j, v)`: Access/update symmetric entries.
- `Size() int`, `GetLabel(i)` / `SetLabel(i, label)`: Query/mutate element labels.
- `Labels() []string`, `GetRow(i)` / `GetColumn(j)`: Retrieve full rows/columns (as copies).

## Analysis Helpers

- `MinDistance()`, `MaxDistance()` → `(value, i, j)` of the extremal off-diagonal entry.
- `Copy() *DistMatrix`: Deep copy for immutability-safe operations.
- `ToFullMatrix()` → `[][]float64`: Converts to dense representation (use sparingly).

Designed for clustering, phylogenetics, or any domain requiring fast symmetric matrix access with minimal footprint.
