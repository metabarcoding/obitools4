# `obik ls` Command — Semantic Description

The `obik ls` command lists metadata about k-mer sets stored in a precomputed index directory. It is part of the `obik` CLI tool, designed for working with biological sequence k-mer sets (e.g., in metabarcoding workflows).

## Core Functionality

- **Input**: Accepts a single positional argument: the path to an index directory created by `obik build` or similar.
- **Index Access**: Uses `OpenKmerSetGroup()` to load a collection of k-mer sets (each representing one sample or taxonomic group).
- **Set Selection**: Optionally filters which k-mers sets to display via `CLISetPatterns()` (e.g., glob patterns like `"sample_*"` or regex-like filters).
- **Metadata Extraction**: For each selected set, retrieves:
  - `index`: numeric ID (position in the group),
  - `id`: human-readable identifier,
  - `count`: number of unique k-mers in the set.

## Output Formats

- **CSV** (default): Tabular format with header `index,id,count`. Properly escapes IDs containing commas or quotes.
- **JSON**: Pretty-printed array of `setEntry` objects with typed fields (`index`, `id`, `count`).
- **YAML**: Human-readable structured output using the same schema.

## Context & Error Handling

- Runs within a `context.Context` for cancellation and timeouts.
- Returns descriptive errors (e.g., invalid path, pattern matching failures).
- Falls back to CSV if an unknown format is requested.

## Use Cases

- Inspect contents of a k-mer index before downstream analysis.
- Validate indexing results (e.g., verify expected sample IDs and k-mer counts).
- Export metadata for scripting or integration with other tools.

> *Note: No k-mers themselves are printed—only metadata about the sets.*
