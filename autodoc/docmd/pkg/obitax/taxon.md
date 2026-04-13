# `obitax` Package: Taxonomic Data Model and Navigation

The `obitax` package provides a semantic model for representing, querying, and manipulating taxonomic hierarchies in biodiversity data processing. Its core abstraction is the `Taxon` type, which encapsulates both structural (node ID, parent/child relationships) and semantic (scientific name, rank, metadata) information.

### Core Features

- **Taxon Representation**: Each `Taxon` links to a taxonomy and its underlying node, supporting multiple name classes (e.g., "scientific name", "common name"), customizable ranks, and extensible metadata via key-value pairs.
- **String Interoperability**: Implements `String()` for human-readable output (`taxonomy:taxid [name]`) and provides typed accessors like `ScientificName()`, `Rank()`, or `IsRoot()`.

### Name Handling & Matching

- Flexible name retrieval via `Name(class)`, case-insensitive equality (`IsNameEqual`), and regex-based matching (`IsNameMatching`). Names are interned for memory efficiency.

### Hierarchical Navigation

- **Path Traversal**: `IPath()` yields an iterator from current taxon up to root; `Path()` materializes this as a slice. Enables efficient lineage queries.
- **Rank-Based Lookup**: Methods like `TaxonAtRank(rank)`, or convenience wrappers (`Species()`, `Genus()`, `Family()`), allow targeted retrieval of higher-level ancestors.
- **Child Management**: Supports dynamic tree extension via `AddChild()`, parsing taxon strings and enforcing taxonomy consistency.

### Metadata Support

- Rich metadata operations: `SetMetadata`, `GetMetadata`, key/value iteration, and typed conversion (`MetadataAsString`). Enables attaching arbitrary annotations (e.g., confidence scores, source references).

### Robustness & Safety

- Nil-safe accessors prevent panics; logging and error handling ensure correctness (e.g., fatal on missing root in `IPath()`).
- Interning of names/ranks/classes (`Innerize`) reduces duplication and speeds comparisons.

Designed for scalability in large-scale metabarcoding pipelines, `obitax` bridges raw taxonomic data with high-level analytical operations.
