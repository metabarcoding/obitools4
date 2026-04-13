# `obikmer` K-mer Set Group Builder — Functional Overview

The `KmerSetGroupBuilder` enables scalable construction of k-mer indexes from biological sequences, supporting both new and incremental (append) workflows. It operates in two phases: **collection** of super-kmers into partitioned temporary files (`.skm`), and **finalization**, where partitions are processed in parallel into final k-mer indexes (`.kdi`).

## Core Features

- **K-mer & Minimizer Configuration**:  
  Supports `k ∈ [2,31]`; auto-computes optimal minimizer size (`m ≈ k/2.5`) and partition count (up to `4^m`, capped at 4096).

- **Functional Options for Filtering**:  
  - `WithMinFrequency(n)`: Keep only k-mers with frequency ≥ *n* (enables deduplication).  
  - `WithMaxFrequency(n)`: Discard k-mers with frequency > *n*.  
  - `WithEntropyFilter(threshold, levelMax)`: Remove low-complexity k-mers (entropy ≤ threshold).  
  - `WithSaveFreqKmers(n)`: Save top-*n* most frequent k-mers per set to `top_kmers.csv`.

- **Concurrent & Pipeline-Aware Processing**:  
  Uses a two-stage pipeline: *I/O-bound readers* (2–4 goroutines) feed k-mers to *CPU-bound workers*, one per core, maximizing throughput.

- **Partitioned I/O & Thread Safety**:  
  Super-kmers are written to per-partition `.skm` files using mutex-protected writers, enabling safe concurrent `AddSequence()` calls.

## Workflow

1. **Build Phase**:  
   - Input sequences → super-kmers extracted via minimizer-based partitioning.  
   - Super-kmers written to `.build/set_*/part_*.skm`.

2. **Finalization (`Close()`)**:  
   - `.skm` files loaded → canonical k-mers extracted.  
   - K-mers sorted, counted (frequency spectrum), and filtered per config.  
   - Final `.kdi` files written; `spectrum.bin`, and optionally `top_kmers.csv`.  
   - Metadata (`metadata.toml`) generated; `.build/` cleaned.

3. **Append Mode**:  
   `AppendKmerSetGroupBuilder()` extends an existing group, inheriting its parameters and appending new sets.

## Output Artifacts

- `.kdi`: Sorted, deduplicated (and optionally filtered) k-mers.  
- `spectrum.bin`: Per-set frequency spectrum (`count → #k-mers`).  
- `top_kmers.csv` (optional): Top *N* k-mers per set with counts.  
- `metadata.toml`: Global and per-set metadata (k, m, partitions, counts).

## Design Highlights

- **Memory-efficient**: Streams large `.skm` files; reuses slices to minimize GC pressure.  
- **Scalable**: Parallel finalization scales with CPU cores and I/O bandwidth.  
- **Robust error handling**: Early termination on first failure; cleanup of partial state.

