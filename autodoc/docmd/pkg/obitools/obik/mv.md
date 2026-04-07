## `obik mv`: Semantic Description

The `mv` command implements a **safe, pattern-based move operation** for k-mer set indices in the `obik` toolchain.

### Core Functionality
- **Moves one or more k-mer sets** from a source index directory to a destination directory.
- Supports **selective move via glob-like set patterns** (`--set PATTERN`), or moves all sets if none specified.
- Uses a **copy-first, then delete** strategy to ensure atomicity and prevent data loss on failure.

### Key Behaviors
- **Validation**: Requires at least two positional arguments: `<source_index>` and `<dest_index>`.
- **Pattern resolution**: Matches user-provided patterns against existing set IDs using `MatchSetIDs`. Fails if no sets match.
- **Forced overwrite**: Respects the `--force` flag (via `CLIForce()`) to allow overwriting existing sets in destination.
- **Order preservation**: Removes source sets *in reverse order* to avoid index renumbering side effects during sequential deletion.
- **Logging**: Reports progress and final state (e.g., number of sets moved, resulting counts in source/destination).

### Semantic Semantics
- Treats k-mer sets as **named, discrete units** within a `KmerSetGroup`.
- The operation is *not* in-place: it physically relocates data, updating directory contents.
- Designed for use with large-scale metagenomic or metabarcoding workflows where k-mer indexing is central.

### Error Handling
Returns descriptive errors for:
- Missing arguments  
- Source index open failure  
- Pattern matching failures (e.g., no matches)  
- Copy or deletion errors with context (`%w` wrapping)

> **Note**: The command assumes `obitools4/pkg/obikmer.KmerSetGroup` provides robust set management (copy, remove, query by ID).
