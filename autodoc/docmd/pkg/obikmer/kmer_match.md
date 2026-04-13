# Semantic Description of `obikmer` Package

The `obikmer` package implements efficient k-mer matching between query sequences and an indexed reference using **canonical k-mers** partitioned by minimizer-based hashing.

- `QueryEntry` represents a single canonical k‑mer with its origin: sequence index and 1-based position.
- `PreparedQueries` groups queries into sorted buckets per partition, enabling batched and parallelized matching.
- `PrepareQueries` scans input sequences using *super-kmers* (with window size `m`) to compute minimizers, assigns each k‑mer to a partition via modulo hashing, and sorts buckets by k‑mer value.
- `MergeQueries` combines two sets of prepared queries across batches using a merge-sort strategy, correctly offsetting sequence indices to preserve global ordering.
- `MatchBatch` performs parallel matching per partition: each goroutine runs a **merge-scan** between sorted queries and the corresponding KDI (K-mer Disk Index) stream.
  - Efficient seeking is used only when beneficial, avoiding costly syscalls for small skips.
  - Matches are recorded with thread-safe per-sequence mutexes; final positions within each sequence are sorted post-match.
- `matchPartition` implements the core merge-scan: it opens a KDI reader, seeks to relevant regions of the index, and walks both query list and k‑mer stream in lockstep.

The design supports **large-scale batch processing**, incremental query accumulation, and high-performance parallel lookup—ideal for metagenomic or biodiversity sequencing workflows.
