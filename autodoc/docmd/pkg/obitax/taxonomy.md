# `obitax` Package: Taxonomic Data Management

The `obitax` package provides a robust framework for managing hierarchical taxonomic classifications. Its core component is the `Taxonomy` struct, which encapsulates metadata (name, code), taxon identifiers (`ids`, `ranks`), names and name classes (`names`, `nameclasses`), node hierarchy (`nodes`, `root`), indexing for fast lookup, and validation logic.

## Key Functionalities

- **Initialization**: `NewTaxonomy()` creates a new taxonomy with configurable identifier alphabet and initializes internal data structures.
- **Identifier Handling**: `Id()` validates and converts string-based taxon IDs to internal representations; `TaxidString()` retrieves formatted identifiers (e.g., `"code:id [name]"`).
- **Taxon Access**: `Taxon()` fetches a taxon by ID, returning whether it's an alias; `AsTaxonSet()` exposes the full taxonomic node collection.
- **Structure Management**:
  - `AddTaxon()` inserts a new taxon with parent, rank, and root flags.
  - `AddAlias()` maps alternative IDs to existing taxa (supporting replacement).
- **Metadata Queries**: Methods like `RankList()`, `Name()`, and `Code()` expose taxonomy metadata.
- **Root Control**: `SetRoot()`/`Root()` manage the root node; `HasRoot()` checks its presence.
- **Path Insertion**: `InsertPathString()` builds or extends a taxonomy from an ordered list of taxon strings, enforcing parent-child consistency.
- **Phylogenetic Export**: `AsPhyloTree()` converts the taxonomy into a phylogeny-compatible tree (`obiphylo.PhyloNode`), enabling downstream evolutionary analysis.

All operations gracefully handle `nil` receivers via an internal `.OrDefault()` helper, ensuring safe usage in pipelines. Error reporting is explicit and contextualized (e.g., duplicate taxon, missing parent).
