# `obitax` Package: Taxonomic Data Handling

The `obitax` package provides structured support for managing collections of taxon nodes in a biological taxonomy.

- **Core Type**: `TaxonSlice` encapsulates an ordered list of `*TaxNode`s and a reference to their parent `Taxonomy`.
- **Construction**: Created via `(taxonomy *Taxonomy).NewTaxonSlice(size, capacity)`, initializing a typed slice with optional pre-allocation.
- **Accessors**:
  - `Get(i int) *TaxNode`: retrieves the raw node at index.
  - `Taxon(i int) *Taxon`: wraps a node with its taxonomy context, enabling richer operations.
  - `Len() int`: returns the current number of nodes.

- **Mutation Methods**:
  - `Set(index, taxon)`: replaces a node at given index (taxonomy-mismatch panics).
  - `Push(taxon)`: appends a taxon to the end (also enforces taxonomy consistency).
  - `ReduceToSize(n)`: truncates slice to first *n* elements.

- **Utility Features**:
  - `Reverse(inplace)`: reverses node order — either in-place or as a new slice.
  - `String() string`: formats the entire path as `"id@sci_name@rank"` entries, separated by `|`, in *reverse* (leaf-to-root) order — ideal for lineage strings.

- **Safety & Semantics**:
  - Nil-safety in all methods (returns `nil` or zero).
  - Enforces taxonomy coherence: mixing taxa from different taxonomies triggers a panic.
  
This package enables efficient, type-safe manipulation of hierarchical biological classification paths (e.g., for sequence annotation or metabarcoding output).
