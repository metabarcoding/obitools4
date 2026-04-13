# ObiTax Lua Module Documentation

This Go package (`obilua`) provides **Lua bindings** for the `obitax` taxonomy management module of OBItools4, enabling scripting in Lua with rich taxonomic operations.

## Core Features

- **Type Registration**: Registers two main types in the Lua state: `Taxonomy` and `Taxon`.
- **Factory Functions**:
  - `obitax.Taxonomy.new(name, code [, charset])`: Creates a new taxonomy instance.
  - `obitax.Taxonomy.default()`: Returns the globally configured default taxonomy (raises error if none exists).
  - `obitax.Taxonomy.has_default()`: Boolean check for existence of a default taxonomy.
  - `obitax.Taxonomy.nil`: Represents the nil taxon (used for missing data).

## Taxonomy Object Methods

- `name()`: Returns the taxonomy name (e.g., `"NCBI"`).
- `code()`: Returns the internal code used for taxonomic identifiers (e.g., `"txid"`).
- `taxon(id)`: Retrieves a taxonomic node by ID; returns:
  - the corresponding *Taxon* object,
  - raises an error if not found or on alias resolution when `FailOnTaxonomy()` is enabled.

## Taxon Object Support

- A dedicated `registerTaxonType` (not shown here) exposes a Lua-accessible *Taxon* type with methods like `rank`, `parent`, and string representation.

## Integration

- Built on top of standard OBItools4 types (`obitax.Taxonomy`, `obiutils.AsciiSetFromString`).
- Leverages GopherLua for seamless interoperability between Go and Lua.
