# ObiTax: Semantic Overview of Public Functionalities

`obitax` is a Go package for managing hierarchical taxonomic data in biodiversity pipelines. It provides thread-safe, iterator-based APIs to query, filter, and traverse taxonomies—while supporting robust defaults, string interning, type-safe identifiers (`Taxid`), and phylogenetic interoperability.

## ✅ Default Taxonomy Management
- **`.SetAsDefault()`**: Registers a `Taxonomy` instance as the global default.
- **`.OrDefault(panicOnNil bool)`**: Substitutes `nil` receivers with the default taxonomy (panics if none exists and `panicOnNil=true`).
- **`.HasDefaultTaxonomyDefined()` / `.OrDefault()`**: Enables safe fallback without boilerplate.

## 🔍 Core Filtering Operations (Iterator-Centric)
All filters return `*ITaxon`, enabling lazy, composable pipelines.

- **`.IFilterOnName(name string, strict bool, ignoreCase bool)`**  
  Filters taxa by name: exact match (`strict=true`) or regex (default). Case-insensitive if `ignoreCase`. Deduplicates via internal node ID.

- **`.IFilterOnTaxRank(rank string)`**  
  Filters taxa whose rank matches (normalized via taxonomy’s internalized ranks map). Supports chaining and concurrent iteration.

- **`.IFilterOnSubcladeOf(parent *Taxon)`**  
  Yields descendants of `parent` (via `.IsSubCladeOf()`). Works on iterators, sets, slices, and taxonomies.

- **`.IFilterBelongingSubclades(clades *TaxonSet)`**  
  Filters taxa belonging to any clade in `clades`. Optimized for single-clade case (reuses `.IFilterOnSubcladeOf`).

## 🌳 Hierarchical Navigation & Relationship Queries
- **`.IsSubCladeOf(parent *Taxon) bool`**: Checks if current taxon descends from `parent`.
- **`.IsBelongingSubclades(clades *TaxonSet) bool`**: Checks if current taxon—or any ancestor—is in `clades`.
- **`.IPath() *ITaxon`**: Iterates upward from taxon to root (breadth-first via `.IPath()`).
- **`.TaxonAtRank(rank string)` / shortcuts (e.g., `.Species()`, `.Genus()`)**: Traverse ancestors to find first match at given rank.

## 🧠 String Interning & Deduplication
- **`InnerString.Innerize(value string) *string`**: Thread-safe deduplication of strings (e.g., names, ranks). Returns shared pointer for equality checks.
- **`.Slice() []string`**: Snapshot of all interned strings (read-only).

## 🔢 Taxonomic Identifiers (`Taxid`)
- **`FromInt(int)` / `FromString(string) *string`**: Validates and normalizes IDs (e.g., `"tx:12345"` → interned `"12345"`). Enforces code prefix, filters to ASCII digits/letters.

## 📜 Taxon String Parsing
- **`ParseTaxonString(taxonStr string)`**: Parses `"code:taxid [name]@rank"` into structured components. Validates brackets, colons, and field presence.

## 🧬 Taxonomy & Node Model
- **`Taxon`**: Encapsulates node ID, parent/children links, scientific name (and alternatives), rank, and metadata.
  - `.Name(class)`, `.ScientificName()`: Flexible name access (case-insensitive matching via `IsNameEqual`/regex).
  - `.SetMetadata(key, value)`, `.GetMetadata(key)` / iteration: Extensible annotations.
  - `.String()`: Human-readable `"code:id [name]@rank"` format.

- **`Taxonomy`**: Manages full hierarchy:
  - `.AddTaxon()` / `.InsertPathString()`: Build trees incrementally.
  - `.Root()`/`.SetRoot()` / `.HasRoot()`: Root node control (required for LCA).
  - `.AsPhyloTree()` → `obiphylo.PhyloNode`: Export to phylogenetic format.

- **`TaxonSet`**: Efficient set of `*TaxNode`s with alias support:
  - `.Alias(id, taxon)`: Non-canonical ID mapping.
  - `.Sort()` → topologically sorted slice (parents before children).
  - `.AsPhyloTree(root)`.

- **`TaxonSlice`**: Ordered, type-safe path representation:
  - `.String()` → `"id@name@rank|..."` (leaf-to-root).
  - Enforces taxonomy coherence; panics on mismatch.

## 🧮 Lowest Common Ancestor (LCA)
- **`.LCA(t2 *Taxon) (*Taxon, error)`**: Computes most specific shared ancestor of two taxa in same rooted taxonomy. Uses path-based backward traversal.

## 🔄 Iterator Composition & Utilities (`ITaxon`)
- **`.Next()`, `.Get()` / `.Finished()`**: Standard iteration control.
- **`.Push(taxon)`, `.Close()`**, and **`Split() / Concat(...)`**: Goroutine-driven streaming, parallel consumption.
- **`.ISubTaxonomy()` / `.ITaxon(taxid)`**: Breadth-first subtree traversal from root or given ID.
- **`.AddMetadata(name, value)`**: Wraps iterator to inject metadata into each taxon.
- **`.Consume()`**: Exhausts an iterator (e.g., for side-effect-only pipelines).

## 🛡️ Safety & Robustness
- Nil-safe accessors (no panics unless explicitly configured).
- Explicit error messages for invalid inputs, cross-taxonomy queries, or unrooted hierarchies.
- Interning reduces memory footprint and accelerates equality checks.

> Designed for scalability in large-scale metabarcoding, biodiversity informatics, and phylogenetic pipelines.
