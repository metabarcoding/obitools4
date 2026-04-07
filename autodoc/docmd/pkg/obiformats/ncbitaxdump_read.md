# NCBI Taxonomy Loader Module (`obiformats`)

This Go package provides functionality to parse and load NCBI taxonomy dump files into a structured `Taxonomy` object. It supports three core file types:

- **nodes.dmp**: Defines the taxonomic hierarchy via `taxid|parent_taxid|rank` records.
- **names.dmp**: Maps taxonomic IDs to names and name classes (e.g., "scientific name", "common name").
- **merged.dmp**: Tracks deprecated taxonomic IDs and their replacements.

Key features:
- Custom CSV parsing with `|` delimiter, comment support (`#`), and whitespace trimming.
- Support for loading *only scientific names* via the `onlysn` flag in `LoadNCBITaxDump`.
- Efficient buffered reading (`bufio.Reader`) for large files.
- Automatic root taxon (taxid `"1"`, i.e., *root*) assignment after loading.
- Alias resolution: deprecated taxids are mapped to current ones via `AddAlias`.
- Robust error handling with fatal logging on critical failures (e.g., missing root taxon, invalid parent references).

The main entry point is `LoadNCBITaxDump(directory string, onlysn bool)`, which constructs a fully initialized taxonomy from NCBI dump files. Designed for integration with `obitax` and `obiutils`, it enables downstream applications (e.g., metabarcoding pipelines) to perform taxonomic queries and filtering.
