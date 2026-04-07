## CSV Taxonomy Loader for OBITools4

This Go module provides a function `LoadCSVTaxonomy` to parse and load taxonomic data from CSV files into an internal taxonomy structure.

### Key Features:
- **Robust CSV Parsing**: Uses Go’s `encoding/csv` with configurable options (comment lines, lazy quotes, whitespace trimming).
- **Column Mapping**: Dynamically identifies required columns: `taxid`, `parent`, `scientific_name`, and `taxonomic_rank`.
- **Error Handling**: Validates presence of all required columns; fails early with descriptive errors.
- **Taxonomy Construction**:
  - Builds a hierarchical taxonomy using `obitax.Taxon` objects.
  - Ensures existence of a root node; returns error otherwise.
- **Metadata Extraction**:
  - Derives taxonomy name and short code (e.g., prefix before `:` in first taxid).
  - Logs key metadata for traceability.
- **Scalable Design**:
  - Processes records line-by-line (memory-efficient).
  - Supports large datasets via streaming CSV reading.

### Input Format:
CSV must contain exactly four columns (case-sensitive headers):
- `taxid`: Unique taxon identifier.
- `parent`: Parent taxonomic node ID (empty for root).
- `scientific_name`: Binomial or descriptive name.
- `taxonomic_rank`: e.g., *species*, *genus*.

### Output:
Returns a fully populated `obitax.Taxonomy` object ready for downstream phylogenetic or sequence classification tasks.
