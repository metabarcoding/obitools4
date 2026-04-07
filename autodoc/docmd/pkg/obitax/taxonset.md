# TaxonSet: Semantic Description of Functionality

The `TaxonSet` type manages a collection of taxonomic entities within a hierarchical taxonomy system. It stores mappings from unique identifiers (pointers to strings) to `TaxNode` instances, supporting both canonical taxa and aliases.

- **Construction**: Created via `(Taxonomy).NewTaxonSet()`, initializing an empty set and linking it to a specific taxonomy.

- **Basic Queries**:
  - `Get(id)`: Retrieves the corresponding taxon (or nil).
  - `Len()`: Returns count of *unique* taxa, excluding aliases.
  - `Contains(id)`, `IsATaxon(id)`, and `IsAlias(id)` enable precise taxon/alias distinction.

- **Insertion & Management**:
  - `Insert(node)`: Adds or updates a taxon node.
  - `InsertTaxon(taxon)`: Safe insertion with taxonomy validation; auto-creates set if nil.
  - `Alias(id, taxon)`: Registers an alias (non-canonical ID pointing to a real node), incrementing internal `nalias` counter.

- **Hierarchy & Iteration**:
  - `Sort()`: Returns a topologically sorted slice of taxa (parents before children), respecting tree structure.
  - `Taxonomy()`: Provides access to the parent taxonomy.

- **Phylogenetic Export**:
  - `AsPhyloTree(root)`: Converts the set into a rooted phylogenetic tree (`obiphylo.PhyloNode`), embedding taxon names, ranks, and parent relationships as node attributes.

In essence, `TaxonSet` enables efficient storage, lookup, validation, and structural manipulation of taxonomic data—supporting both biological classification logic (e.g., alias resolution, hierarchy traversal) and downstream interoperability with phylogenetic tools.
