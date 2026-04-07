# Semantic Overview of `obitax` Filtering Functionalities

The `obitax` package provides composable, iterator-based filtering methods for taxonomic data structures. All filters return lazy or buffered iterators (`*ITaxon`) enabling efficient, streaming-style traversal without materializing full collections.

## Core Filtering Operation: `IFilterOnSubcladeOf`

- **Purpose**: Filters elements belonging to a specific taxonomic subtree.
- **Behavior**:
  - Accepts a `*Taxon` as reference root.
  - Yields only taxa for which `IsSubCladeOf(taxon)` returns true (i.e., descendants of the given taxon).
- **Overloads**:
  - On `*ITaxon`, `TaxonSet`, `TaxonSlice`, and `Taxonomy` — all delegate to the iterator variant.
  - Ensures consistent interface across container types.

## Composite Filtering: `IFilterBelongingSubclades`

- **Purpose**: Filters taxa belonging to *any* of a set of specified subclade roots.
- **Behavior**:
  - Accepts `*TaxonSet` of clades (roots).
  - Uses optimized path for single-clade case: reuses `IFilterOnSubcladeOf`.
  - For multiple clades, checks via `IsBelongingSubclades(clades)` in a goroutine.
  - Returns original iterator unchanged if input set is empty.

## Design Highlights

- **Iterator-Centric**: All operations are defined on `ITaxon`, promoting chaining and lazy evaluation.
- **Concurrency Support**: Filtering uses goroutines with buffered channels (`source`), enabling asynchronous stream processing.
- **Type Abstraction**: Unified API across `TaxonSet`, `Slice`, and full `Taxonomy` via delegation.
- **Performance Consideration**: Special handling for single-clade case avoids unnecessary iteration overhead.

These methods enable expressive, scalable taxonomic queries—ideal for phylogenetic analysis or biodiversity data pipelines.
