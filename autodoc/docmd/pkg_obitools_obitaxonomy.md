# `obitaxonomy`: CLI-Oriented Taxonomic Data Utilities for OBItools4

The `obitaxonomy` Go package delivers modular, command-line-friendly tools for loading, filtering, navigating, and exporting taxonomic data within the OBItools4 ecosystem. It focuses on enabling reproducible, scriptable workflows for metagenomics and biodiversity informatics by abstracting complex taxonomy operations behind intuitive CLI flags.

## Public Functionalities

### Taxonomy Restriction & Filtering
- **`CLITaxonRestrictions()`**: Wraps a taxonomy iterator to apply user-defined clade restrictions via `--restrict-to-taxon` (`-r`). Supports taxon IDs or names (with optional regex), returning a filtered iterator over matching subtrees.
- **`CLIFilterRankRestriction()`**: Restricts the taxonomy iterator to taxa of a specific rank (e.g., `"species"`, `"family"`), controlled by `--rank` (`-R`). Returns a constrained iterator for downstream processing.

### Subtree Navigation & Iteration
- **`CLISubTaxonomyIterator()`**: Returns an iterator over the subtree rooted at a user-specified taxon ID (via `--dump`/`-D`). If no root is provided, exits with an error—enabling safe CLI-driven subtree extraction.

### CSV Export
- **`CLICSVTaxaIterator()`**: Transforms a taxonomy iterator into an ordered stream of CSV records. Configurable columns include:
  - Scientific name (`--without-scientific-name` to omit),
  - Taxonomic rank (omittable via `-R`),
  - Parent taxon ID (`--without-parent`/`-W`),
  - Full lineage path (via `--path`, `-P`),
  - Query source match (`--with-query`).
- **`CLICSVTaxaWriter()`**: Wraps `CLICSVTaxaIterator()`, handling output destination (`-` = stdout, file path otherwise), and integrates with CLI logging.

### Tree Export
- **`CLINewickWriter()`**: Exports a taxonomy subtree (from `--dump`) as Newick format. Supports:
  - Compression (`gzip` via `-z`),
  - Leaf labels (scientific name/rank/taxid toggles),
  - Root trimming (`--trim-root`),
  - Output to file or stdout.

### Data Acquisition
- **`CLIDownloadNCBITaxdump()`**: Fetches the latest NCBI taxonomy dump (`taxdump.tar.gz`) and saves it as `ncbitaxo_YYYYMMDD.tgz` (or custom name). Designed for one-click taxonomy setup.

### Utility & Inspection Helpers
- **`CLIRankRestriction()` / `CLIWithScientificName()`**: Expose parsed CLI flags for use in custom processing pipelines.
- **`--rank-list` (`-l`)**: Prints all available ranks in the loaded taxonomy (for introspection).
- **Pattern matching**: `--fixed` (`-F`) disables regex for taxon name queries, enabling literal string matching.

## Integration & Design Principles

- Built on `obitax` for core taxonomy operations.
- Fully compatible with OBItools4’s option parsing (`getoptions`) and iterator patterns.
- Designed for composition: integrates seamlessly with `obiconvert` (output formatting) and other CLI modules.
- All functions respect `-`, stdout/stderr conventions, logging levels (`--verbose`), and CLI flag parsing.
- No internal state mutation—functions are pure wrappers around iterator transformations.

## Target Use Cases

- Filtering metagenomic assignments to a clade of interest (e.g., `--restrict-to-taxon 9606` for *Homo sapiens*).
- Exporting species-level taxa to CSV/JSON for downstream analysis.
- Generating Newick trees from custom taxonomic subsets (e.g., all *Enterobacteriaceae*).
- Bootstrapping local taxonomy caches via `--download-ncbi`.
