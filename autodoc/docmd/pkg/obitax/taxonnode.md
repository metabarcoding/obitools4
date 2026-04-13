# `obitax` Package: Taxonomic Node Representation and Management

The `obitax` package provides a lightweight, pointer-based Go implementation for representing taxonomic nodes in biological classification systems.

## Core Data Structure

- **`TaxNode`**: Represents a single taxon (e.g., species, genus) with the following fields:
  - `id`: Unique taxon identifier (pointer to string).
  - `parent`: Identifier of the parent node in the taxonomy hierarchy.
  - `rank`: Taxonomic rank (e.g., `"species"`, `"family"`).
  - `scientificname`: Canonical scientific name (e.g., *Homo sapiens*).
  - `alternatenames`: Map of alternative names keyed by name class (e.g., `"common_name"`, `"synonym"`).

## Key Functionalities

- **String Representation**  
  `String(taxonomyCode)` returns a formatted label like `"NCBI:12345 [Homo sapiens]@species"` (or raw ID if enabled via `obidefault.UseRawTaxids()`).

- **Accessors**  
  - `Id()`, `ParentId()`: Retrieve identifiers.
  - `ScientificName()` / `Rank()`: Return name or rank (defaulting to `"NA"` if missing).
  - `Name(class)`: Fetch name by class (`"scientific name"` or alternate).

- **Mutators**  
  - `SetName(name, class)`: Assign scientific name or add/update alternate names.

- **Name Matching & Validation**  
  - `IsNameEqual(name, ignoreCase)`: Exact or case-insensitive match against scientific/alternate names.
  - `IsNameMatching(pattern)`: Regex-based pattern matching over all available names.

## Design Notes

- Uses pointers for optional fields (enables `nil` semantics).
- Graceful handling of missing data (`NA`, empty strings, safe dereferencing with `nil` checks).
- Integrates logging via Logrus (`log.Panic` on misuse, e.g., setting name of `nil` node).
- Designed for use in larger OBITools pipelines (e.g., with `obidefault` configuration).
