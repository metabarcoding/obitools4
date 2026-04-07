# `obitax` Package: Taxon String Parser

The `obitax` package provides a robust parser for structured taxonomic strings used in biodiversity data processing.

## Core Functionality

- **`ParseTaxonString(taxonStr string)`**  
  Parses strings in the format: `code:taxid [scientific name]@rank`.

- **Input Format Requirements**  
  - `code`: Taxonomy identifier (e.g., "GBIF", "NCBI")  
  - `taxid`: Numeric or alphanumeric taxonomic ID (e.g., "123456")  
  - `scientific name`: Enclosed in square brackets (e.g., "[Homo sapiens]")  
  - `rank`: Optional taxonomic rank after `@` (e.g., "species", defaults to `"no rank"` if missing)

- **Robustness Features**  
  - Trims whitespace around all components.  
  - Handles multiple `@` symbols (returns error).  
  - Validates bracket pairing and ordering.  
  - Ensures `code:taxid` contains exactly one colon separator.

- **Error Handling**  
  Returns descriptive errors for:
    - Missing or malformed brackets
    - Invalid number of `@` separators
    - Absent colon in code:taxid segment  
    - Empty fields (code, taxid, or scientific name)

- **Use Cases**  
  Ideal for parsing legacy biodiversity records (e.g., from OBIS, GBIF), where taxon strings are semi-structured and need reliable extraction before indexing or matching against reference databases.

## Example

Input: `"GBIF:248093 [Homo sapiens]@species"`  
Output components:
- `code = "GBIF"`
- `taxid = "248093"`
- `scientificName = "Homo sapiens"`
- `rank = "species"`

Returns empty strings and an error for invalid inputs.
