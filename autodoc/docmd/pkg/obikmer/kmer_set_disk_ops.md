# Semantic Description of `obikmer` Set Operations

This Go package implements scalable set operations over collections of *k*-mers stored in disk-backed, sorted structures (`.kdi` files). A `KmerSetGroup` represents a group of *N* disjoint sets (e.g., per-sample or per-partition), each containing sorted unique *k*-mers.

## Core Set Operations

- **`Union()`**: Computes the union across all *N* sets — a k-mer appears in output if present in ≥1 input set.
- **`Intersect()`**: Computes the intersection — a k-mer appears only if present in *all* sets.
- **`Difference()`**: Computes `set₀ \ (set₁ ∪ … ∪ setₙ₋₁)` — keeps k-mers unique to the first set.
- **`QuorumAtLeast(q)`**: Returns k-mers present in ≥ *q* sets.
- **`QuorumExactly(q)`**: Returns k-mers present in exactly *q* sets.
- **`QuorumAtMost(q)`**: Returns k-mers present in ≤ *q* sets.

## Pairwise Group Operations

- **`UnionWith(other)` / `IntersectWith(other)`**: Performs *per-set* binary operations between two compatible groups (same k, m, partitions, size). Result has *N* sets: `setᵢ = this.setᵢ ⊕ other.setᵢ`, where ⊕ is union or intersection.

## Implementation Highlights

- **Partitioned & Parallelized**: Each operation processes partitions in parallel using `runtime.NumCPU()` workers.
- **Streaming K-way Merge**: Uses efficient sorted-stream merging (via `KWayMerge`) to avoid loading full sets into memory.
- **Quorum Filtering**: Counts occurrences per k-mer across partitions by merging sorted streams and tallying hits.
- **Compatibility Check**: Ensures groups share metadata (k, m, partitions) before pairwise operations.
- **Disk Output**: All results materialized as new `KmerSetGroup` in a directory, with per-partition `.kdi` files and metadata.

All operations preserve sorted order and support large-scale genomic datasets via streaming, partitioning, and minimal memory footprint.
