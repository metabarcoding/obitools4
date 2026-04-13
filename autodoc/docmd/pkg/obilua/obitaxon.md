# Lua Bindings for Taxonomic Operations in `obilua`

This Go package provides a set of **Lua-accessible functions** for manipulating taxonomic data through the `obitax` library. It exposes a custom Lua type, `"Taxon"`, enabling users to create and query hierarchical taxonomic entities directly from Lua scripts.

## Core Features

- **Taxon Type Registration**:  
  A new userdata type `Taxon` is registered in the Lua state, with methods exposed via a metatable and `"__index"` delegation.

- **Taxon Creation**:  
  The `Taxon.new(taxid, parent, sname, rank[, isroot])` constructor creates a new taxon node in the taxonomy. It supports optional root flag and raises errors on failure.

- **Scientific Name Management**:  
  `taxon:scientific_name([newname])` gets or sets the scientific name of a taxon.

- **Taxonomic Navigation**:  
  Methods allow upward/downward traversal:
  - `taxon:parent()` → returns the parent taxon (or nil if root).
  - `taxon:species()`, `.genus()`, `.family()` → return the nearest taxon at that rank.
  - `taxon:taxon_at_rank(rank)` → returns the ancestor taxon at a given rank (e.g., `"order"`, `"class"`).

- **String Representation**:  
  `taxon:string()` returns a human-readable string (typically the scientific name).

- **Integration with Taxonomy Context**:  
  All operations assume an active taxonomy context (enforced via `checkTaxonomy`), and taxon instances are wrapped as Lua userdata with proper type checking.

## Use Case

Ideal for scripting biodiversity pipelines (e.g., in OBITools), where users need to dynamically inspect or build taxonomies during sequence annotation, filtering, or reporting.
