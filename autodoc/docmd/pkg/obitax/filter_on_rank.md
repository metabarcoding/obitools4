# Semantic Description of `IFilterOnTaxRank` Functionality in the *obitax* Package

The `IFilterOnTaxRank` method enables semantic filtering of taxonomic data by rank (e.g., `"species"`, `"genus"`). It is implemented across multiple core types—`ITaxon`, `TaxonSet`, `TaxonSlice`, and `Taxonomy`—providing a unified interface for rank-based selection.

- **Core behavior**: Returns an `*ITaxon` iterator containing only taxa whose node’s rank matches the input string.
- **Rank normalization**: Internally, it resolves the requested `rank` against a taxonomy’s internal rank map via `ptax.ranks.Innerize(rank)`, ensuring consistent mapping and case-insensitive or canonical representation handling.
- **Efficiency**: Reuses the resolved rank pointer (`prank`) across consecutive taxa from the same `Taxonomy`, avoiding redundant lookups.
- **Concurrency-safe iteration**: Uses a goroutine to stream filtered results into the new iterator’s channel (`newIter.source`), enabling lazy evaluation and memory-efficient processing of large datasets.
- **Polymorphic dispatch**: Overloaded methods on `TaxonSet`, `TaxonSlice`, and `Taxonomy` delegate to the base iterator implementation, preserving consistency across input types.
- **Non-destructive**: Does not mutate source collections; instead produces a new iterator, supporting functional-style chaining.

This design supports scalable taxonomic querying in phylogenetic or biodiversity analysis pipelines, where filtering by hierarchical rank is essential.
