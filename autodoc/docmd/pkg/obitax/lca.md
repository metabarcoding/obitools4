# Semantic Description of `obitax.LCA()` Functionality

The `LCA` method computes the **Lowest Common Ancestor (LCA)** of two taxonomic entities (`Taxon` instances) within a shared hierarchical taxonomy.

- **Input**: A pointer to another `*Taxon` (`t2`) and the receiver taxon (`t1`).  
- **Output**: A `*Taxon` representing their LCA, or an error detailing why computation failed.

### Core Logic
- **Nil Safety**: Handles cases where one or both taxa are `nil`, returning the non-nil taxon (or an error if *both* are nil or lack internal `Node` references).
- **Validation Checks**:
  - Ensures both taxa belong to the *same* `Taxonomy`.
  - Verifies that the taxonomy is **rooted** (i.e., has a defined root node).
- **Path-Based Traversal**:  
  - Retrieves the full path from each taxon to the root via `Path()` (assumed to return an ordered list of nodes).
  - Traverses both paths *backwards* (from root toward leaves) until divergence is detected.
  - The first divergent node marks the boundary; the LCA is the last *common* ancestor (i.e., `slice[i+1]` after loop exit).

### Semantic Meaning
- The LCA represents the most specific taxonomic node that *contains both taxa* in its subtree.
- This operation is foundational for tasks like:
  - Taxonomic classification consistency checks,
  - Phylogenetic inference (e.g., computing taxon distances),
  - Hierarchical aggregation in biodiversity analyses.

### Error Handling
Explicit errors cover:
- Invalid inputs (`nil` taxa, missing nodes),
- Cross-taxonomy queries,
- Unrooted taxonomy (undefined root → no unique LCA possible).

This implementation assumes a **directed acyclic graph** (specifically, a tree) structure for the taxonomy hierarchy.
