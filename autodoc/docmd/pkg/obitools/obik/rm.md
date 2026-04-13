## `obik rm` Command — Semantic Description

The `rm` subcommand removes one or more *k-mer sets* from a prebuilt OBITools4 k-mer index directory.

### Core Functionality
- **Target**: A valid k-mer index directory (containing serialized `KmerSetGroup` data).
- **Input**: One or more glob-like patterns via repeated `--set PATTERN` flags.
- **Validation**:
  - Requires at least one pattern (`--set`);
  - Ensures the index directory exists and is readable;
  - Confirms at least one set matches each provided pattern.

### Execution Flow
1. Parses and collects all `--set` patterns via CLI.
2. Opens the k-mer index (`obikmer.OpenKmerSetGroup`) at `index_directory`.
3. Matches patterns to internal set IDs using fuzzy/regex-style matching (`MatchSetIDs`).
4. Collects the full list of set IDs to be removed *before* deletion (to avoid index shifting).
5. Removes sets **in reverse order** to preserve indices during bulk deletion.
6. Logs each removal step and final index size.

### Safety & Observability
- Uses structured logging (`logrus`) for traceable, human-readable output.
- Wraps errors with contextual messages (e.g., `failed to remove set "SRR123"`).
- Fails fast if any removal fails, leaving the index in a consistent (partial) state.

### Use Case
Enables selective cleanup of sample- or experiment-specific k-mer sets from a shared index—e.g., after filtering, reprocessing, or quality control.
