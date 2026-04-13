# Semantic Description of `obikmer` Package

The `obikmer` package provides high-performance, disk-backed utilities for **k-mer manipulation and comparison** in biological sequences. Designed for scalability (e.g., metagenomics, NGS read processing), it supports canonical encoding, minimizer-based partitioning, streaming I/O formats (`.kdi`, `.skm`), entropy filtering, and scalable set operations — all while minimizing allocations.

---

## Core Encoding & Canonicalization

- **`EncodeKmer`, `DecodeKmer`**: Encodes/decodes DNA sequences to/from compact 62-bit `uint64`s (2 bits/base), preserving top 2 bits for error metadata.
- **`EncodeCanonicalKmer`, ` CanonicalKmer`**: Normalizes k-mers to their *biological canonical form* — the lexicographically smaller of a k-mer and its reverse complement.
- **`IterCanonicalKmers`, `IterCanonicalKmersWithErrors`**: Memory-efficient streaming of canonical k-mers from sequences; optionally tags ambiguous bases in top 2 bits.

## Minimizer-Based Partitioning

- **`DefaultMinimizerSize(k)`**, **`ValidateMinimizerSize(m, k, nworkers)`**: Computes and validates minimizer size `m` for parallelization (e.g., `ceil(k / 2.5)`).
- **`ExtractSuperKmers`, `IterSuperKmers(seq, k, m)`**: Extracts *super-k-mers* — maximal contiguous regions where all embedded `k`-mers share the same minimizer. Uses monotone deque for O(n) time.

## I/O Formats & Streaming

- **`.kdi` (K-Disk Index)**: Compact binary format for sorted `uint64` k-mers using delta-varint encoding. Includes optional `.kdx` sparse index for fast `SeekTo(target)`.
  - APIs: `NewKdiWriter`, `NewKdiReader`, `.Next() → (kmer, ok)`.
- **`.skm`**: Binary storage for *super-k-mers*, with 2-bit nucleotide packing (4× compression vs ASCII).
- **`.kdx`**: Sparse index for `.kdi`, storing `(kmer, byteOffset)` every *stride* entries (e.g., 4096), enabling O(log M) seeks.

## K-Way Merge & Deduplication

- **`KWayMerge([]*KdiReader)`**: Merges sorted `.kdi` streams, aggregating k-mer counts across inputs.
  - Uses min-heap for O(log *k*) per-output operations; supports streaming and deduplication.
  - Ideal for combining k-mer sets across samples or batches.

## Entropy Filtering & Complexity Detection

- **`KmerEntropy(kmer, k, levelMax)`**: Computes minimum normalized Shannon entropy across sub-word sizes (1 to `levelMax`) using circular canonical normalization.
  - Values near **0** indicate repeats (e.g., homopolymers); ~1 indicates high complexity.
- **`KmerEntropyFilter`**: Precomputed filter for batch processing (no allocations), with `Accept(kmer)` and fast entropy lookup.

## K-mer Set Management (`KmerSetGroup`)

A `KmerSetGroup` represents *N* disjoint, sorted k-mer sets (e.g., per sample), persisted on disk.

### Lifecycle & Construction
- **`NewKmerSetGroupBuilder(...)`**, **`AppendKmerSetGroupBuilder(dir)`**: Builds or extends groups via:
  - `AddSequence(setID, bioseq)`: Extracts canonical k-mers (with optional filtering).
  - Supports `WithMinFrequency`, `WithEntropyFilter`, and top-*N* tracking.
- **`Close()`**: Finalizes `.kdi`s, `spectrum.bin`, and optional `top_kmers.csv`.
- **`OpenKmerSetGroup(dir)`**: Loads existing group in read-only mode.

### Access & Metadata
- **`K()`, `M()`, `Partitions()`**, attributes via `GetStringAttribute(key)`.
- **`Contains(setID, kmer)`**: Parallel membership check across partitions.
- **`Iterator(setID)`**: Yields sorted k-mers via k-way merge.

### Set Algebra & Similarity
- **Set Operations**: `Union()`, `Intersect()`, `Difference()`, `QuorumAtLeast(q)` (≥ *q* sets), etc.
- **Pairwise Group Ops**: `UnionWith(other)`, `IntersectWith(other)` (per-set, compatible groups only).
- **Similarity Metrics**:  
  `JaccardDistanceMatrix()` = 1 − |A ∩ B| / |A ∪ B|  
  `JaccardSimilarityMatrix()` = |A ∩ B| / |A ∪ B|

### Utilities
- **`CopySetsByIDTo(ids, destDir)`**, `RemoveSetByID(id)`, `MatchSetIDs(patterns)`
- **`IsCompatibleWith(other)`**: Validates `(k, m, partitions)`.
  
## K-mer Indexing & Matching (`KmerMap`)

Generic hash map associating canonical k-mers to sequences containing them.

- **`Push(sequences)`**: Builds index (optionally with `maxocc` limit).
- **`Query(querySeq) → KmerMatch`**: Returns sequences sharing k-mers, with match counts.
- **Supports sparse mode** (`SparseAt ≥ 0`): Ignores central base (e.g., for ambiguous-position matching).
- **Result utilities**: `FilterMinCount`, `.Max()`, `.Sequences()`.

## K-mer Spectrum Analysis

- **`SpectrumEntry{Frequency, Count}`**, `KmerSpectrum`: Sorted frequency distribution.
- **`MapToSpectrum()`, `MergeTopN()`**, binary/CSV I/O (`WriteSpectrum`, `ReadSpectrum`).
- **Top-*N* collector** via min-heap for streaming frequency tracking.

## Utility & Helpers

- **`HammingDistance(a, b)`**: Bitwise distance between encoded k-mers.
- **Varint encoding/decoding** (`EncodeVarint`, `DecodeVarint`): 7-bit-per-byte compression for I/O.
- **Reverse complement**: Constant-time via lookup tables (`revcompnuc`, `kmermask`).

---

## Design Principles

- **Zero-allocation where possible** (buffer reuse, iterators).
- **Streaming-first**: Avoids loading large datasets into memory.
- **Disk-backed persistence** for reproducibility and scalability.
- **Canonicalization & symmetry**: Strand-aware (reverse complement) or circular normalization for robustness.

## Use Cases

- Metagenomic read clustering & error correction  
- Minimizer-based sketching (e.g., Mash/Sourmash analogs)  
- Scalable Jaccard-based similarity matrices across thousands of samples  
- Low-complexity region detection via entropy filtering  

All operations are tested, benchmarked, and optimized for high-throughput genomic workflows.
