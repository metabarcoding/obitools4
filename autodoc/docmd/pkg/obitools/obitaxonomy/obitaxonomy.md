This Go package `obitaxonomy` provides CLI-oriented utilities for manipulating and exporting taxonomic data within the OBITools4 framework.

- **`CLITaxonRestrictions()`**: Applies user-defined taxonomic clade and rank filters to a taxonomy iterator, returning a constrained view.
- **`CLIFilterRankRestriction()`**: Filters the taxonomy iterator to include only taxa matching a specified taxonomic rank (e.g., "species", "genus").
- **`CLISubTaxonomyIterator()`**: Returns an iterator over a subtree of the default taxonomy, starting from a specified node; exits if no sub-taxonomy is selected via CLI.
- **`CLICSVTaxaIterator()`**: Converts a taxonomy iterator into an CSV record stream, supporting optional inclusion of scientific names, ranks, paths, parent taxa IDs, and raw taxids.
- **`CLICSVTaxaWriter()`**: Wraps `CLICSVTaxaIterator()` to produce a CSV writer, handling output destination and terminal execution.
- **`CLINewickWriter()`**: Exports a taxonomy subtree as Newick format (with optional compression, rank/scientific name inclusion, taxid support), writing to file or stdout.
- **`CLIDownloadNCBITaxdump()`**: Downloads the latest NCBI taxonomy dump (`taxdump.tar.gz`) and saves it as `ncbitaxo_YYYYMMDD.tgz` (or a user-specified filename).

All functions integrate with CLI flags and logging, support output redirection (`-` for stdout), and rely on standardized iterators from the `obitools4/pkg/...` ecosystem.
