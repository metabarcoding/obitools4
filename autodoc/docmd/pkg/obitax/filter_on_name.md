# Semantic Description of `IFilterOnName` Functionality in the `obitax` Package

The `IFilterOnName` method enables filtering taxonomic data (`Taxon`) instances by name, supporting both **exact** and **pattern-based matching**, with optional case-insensitive comparison.

- Two overloaded versions exist:  
  - On `*Taxonomy`: delegates to its iterator.  
  - On `*ITaxon`: performs the actual filtering logic.

- **Parameters**:  
  - `name` (`string`) – search term or regex pattern.  
  - `strict` (`bool`) — if true, performs exact name equality; otherwise treats `name` as a regex.  
  - `ignoreCase` (`bool`) — when true, performs case-insensitive matching (applies to both modes).

- **Core behavior**:  
  - Uses a `map` (`sentTaxa`) to avoid duplicate taxa (based on internal node ID).  
  - For `strict = true`: compares names using a dedicated equality method (`IsNameEqual`).  
  - For `strict = false`: compiles and applies a regex pattern (`regexp.MustCompile`) — prepends `(?i)` for case-insensitive matching.  
  - Filtering runs in a **goroutine**, streaming results into a new `ITaxon` iterator.  
  - Source channel is properly closed after iteration.

- **Return value**: a new `*ITaxon` iterator containing only matching taxa — preserving immutability and enabling chaining.

- **Use cases**:  
  - Find exact species names (e.g., *Homo sapiens*).  
  - Search using partial or regex patterns (e.g., `^Pan.*` for *Panthera* and related genera).  
  - Case-insensitive lookups (e.g., "homo sapiens", "HOMO SAPIENS").

The design emphasizes **efficiency**, **correctness** (deduplication), and **flexibility** in taxonomic querying.
