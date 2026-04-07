# `obik` Index Command — Semantic Description

The `runIndex` function implements a high-performance, parallelizable k-mer indexing pipeline for biological sequence data (e.g., DNA/RNA reads). It constructs or extends a *k-mer set group*—a structured collection of k-mers with metadata and filtering.

### Core Functionality

- **Directory-based indexing**: Outputs are stored in a user-specified directory (`--out`), supporting both creation of new indices and incremental appending to existing ones via a `metadata.toml` manifest.
- **Configurable k-mer parameters**:
  - K-mer size `k` (2–31, validated).
  - Minimizer size `m`, used for space-efficient hashing.
- **Frequency filtering**:
  - Minimum occurrence (`--minocc`) excludes rare k-mers (e.g., sequencing errors).
  - Optional maximum occurrence (`--maxocc`) filters overrepresented k-mers (e.g., contaminants).
- **Entropy-based filtering**: Removes low-complexity/k-mer bias using an entropy threshold over a sliding window (`--entropy-threshold`, `--entropy-size`).
- **Top-frequency k-mers preservation**: Optionally saves the most frequent *N* k-mers for downstream analysis (`--save-freq-kmers`).

### Parallel Processing

- Sequences are read concurrently using `CLIReadBioSequences`.
- A worker pool (`nworkers`, derived from system defaults) processes batches in parallel via `obiiter.IBioSequence`.
- Thread-safe counting of processed sequences (`atomic.Int64`) ensures correctness.

### Metadata & Tagging

- Supports three levels of metadata:
  - **Group-level attributes** (`--set-tag`, `-S`) applied to the entire index.
  - **Set-level metadata** (`-T` / `_setMetaTags`) applied to the newly added k-mer set.
  - **Per-set ID** (`--index-id`) for identification in multi-dataset indices.

### Finalization & Output

- `builder.Close()` finalizes the index and persists k-mers to disk (likely as binary or compressed format).
- Metadata is re-saved with updated statistics and filtering flags.
- Final summary logs total k-mers in the new set, directory path, and processing stats.

### Dependencies & Integration

- Built on top of `obitools4` ecosystem: sequence parsing (`obiconvert`, `obiiter`), k-mer management (`obikmer`), and defaults handling (`obidefault`, `logrus` logging).
- Designed for CLI usage (via `getoptions`) and integration into larger bioinformatics workflows.
