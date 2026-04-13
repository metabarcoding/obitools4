# `obikmer` Package: Disk-Based K-mer Set Group Management

The `obikmer` package provides a streaming, disk-backed implementation for managing collections of *k*-mer sets (called **K-mer Set Groups**), optimized for large-scale metagenomic or genomic analyses.

### Core Concepts
- A **KmerSetGroup** stores *N* disjoint sets of sorted *k*-mers, partitioned into *P* files per set.
- Each group is defined by immutable parameters: `k` (*mer size), `m* (minimizer size), and *P* partitions.
- Data is stored on disk as `.kdi` files (sorted k-mers) with optional sparse indices (`.kdx`) for fast lookup.
- Metadata is serialized in TOML format (`metadata.toml`), supporting both group-level and per-set attributes.

### Key Functionalities

#### 1. **Lifecycle Management**
- `OpenKmerSetGroup(directory)` loads an existing index in read-only mode.
- `NewFilteredKmerSetGroup(...)` constructs a new group (e.g., after filtering).
- `SaveMetadata()` persists metadata changes to disk.

#### 2. **Accessors & Metadata**
- Basic properties: `K()`, `M()`, `Partitions()`, `Size()` (i.e., *N*), and group ID.
- Attribute API: get/set/delete user-defined metadata (group-level or per-set).
  - Supports type coercion (`GetIntAttribute`, `GetStringAttribute`).

#### 3. **Membership & Iteration**
- `Contains(setIndex, kmer)` checks presence using indexed binary search + linear scan across all partitions (parallelized).
- `Iterator(setIndex)` yields sorted *k*-mers via k-way merge of partition readers.

#### 4. **Similarity & Distance Metrics**
- `JaccardDistanceMatrix()` and `JaccardSimilarityMatrix()`: compute pairwise metrics in a streaming fashion.
  - Per-partition processing with parallel goroutines and sorted merge for accurate set intersection/union counts.

#### 5. **Set Management**
- `CopySetsByIDTo(ids, destDir)` copies selected sets (with metadata) to another group.
  - Supports compatibility checks and optional overwriting (`force`).
- `RemoveSetByID(id)` deletes a set, renumbers remaining sets for contiguous indices.
- Glob pattern matching: `MatchSetIDs(patterns)` resolves IDs like `"sample_*"`.

#### 6. **Compatibility & Utility**
- `IsCompatibleWith(other)` verifies same `(k, m, partitions)`.
- Helper methods: `PartitionPath`, `Spectrum(...)`, and spectrum file I/O.

### Design Highlights
- **Streaming**: Operations avoid loading full datasets into memory.
- **Immutability after creation** ensures consistency; modifications require explicit save operations.
- Thread-safe for concurrent partition processing (via `sync.Mutex`/`WaitGroup`).
