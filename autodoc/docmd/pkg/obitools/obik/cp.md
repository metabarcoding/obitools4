## `obik cp`: K-mer Index Set Copy Command

The `cp` subcommand copies selected or all k-mer sets from a source index to a new destination directory.

### Core Functionality
- **Source & Destination**: Requires two positional arguments: `<source_index>` (existing k-mer index) and `<dest_dir>` (new directory for copied sets).
- **Set Selection**:
  - By default, copies *all* k-mer sets in the source index.
  - Supports `--set PATTERN` options to filter and copy only sets whose IDs match the given glob-style patterns.
  - Fails if no set matches any provided pattern.

- **Overwrite Control**:
  - Uses `--force` flag to allow overwriting an existing destination index (or its contents).
  - Without `--force`, copying into a non-empty or conflicting directory is prevented.

- **Underlying Operations**:
  - Opens the source index via `obikmer.OpenKmerSetGroup`.
  - Matches patterns using `MatchSetIDs`, resolves IDs via `SetsIDs`/`SetIDOf`.
  - Copies selected sets using the method `CopySetsByIDTo`, which creates a new k-mer index at `<dest_dir>`.

### Logging & Feedback
- Logs the number of sets being copied and source/destination paths.
- After completion, reports how many sets are present in the new index (`dest.Size()`).

### Error Handling
- Validates argument count.
- Wraps and reports errors from index opening, pattern matching, copying, etc., with context.

This command enables selective migration or duplication of k-mer-based biological sequence indexes (e.g., for taxonomic classification), supporting flexible workflows in OBITools4 pipelines.
