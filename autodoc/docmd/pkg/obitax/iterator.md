# Semantic Description of `obitax` Package Functionalities

The `obitax` package provides a robust iterator-based API for traversing taxonomic data structures in Go. Its core component is the `ITaxon` interface, which implements a lazy, concurrent-safe iterator over taxon instances (`*Taxon`). Key features include:

- **Iterator Creation**: `ITaxon` can be instantiated via `NewITaxon()` or derived from collections:  
  - `TaxonSet.Iterator()`, `TaxonSlice.Iterator()` (sorted), and `Taxonomy.nodes.Iterator()`  
  - Goroutines feed taxa into a channel, enabling non-blocking iteration.

- **Control Methods**:  
  - `Next()` advances to the next taxon, returning success/failure.  
  - `Get()` retrieves the current taxon (must follow a successful `Next`).  
  - `Finished()` checks if iteration is complete.

- **Channel Management**:  
  - `Push(taxon)` sends a taxon into the iterator’s channel.  
  - `Close()` terminates iteration by closing the source channel.

- **Iterator Composition**:  
  - `Split()`: creates a new iterator sharing the same source and termination status (useful for parallel consumption).  
  - `Concat(...)`: merges multiple iterators sequentially into one.

- **Metadata Enrichment**:  
  - `AddMetadata(name, value)` wraps the iterator to inject metadata into each taxon via `SetMetadata`.

- **Subtree Traversal**:  
  - `ISubTaxonomy()` (on `*Taxon` or via `Taxonomy.ITaxon(taxid)`) performs a breadth-first traversal of descendant taxa, starting from the current taxon or given ID. It uses parent-child adjacency logic to expand the subtree incrementally.

- **Consumption Utility**:  
  - `Consume()` exhausts an iterator without processing (e.g., for side-effect-only pipelines).

All iterators are designed to be composable, memory-efficient (via channels), and safe for concurrent use. The package integrates with `obiutils` to manage pipeline registration/unregistration during subtree expansion.
