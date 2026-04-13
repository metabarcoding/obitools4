# Semantic Description of `obirefidx` Package

The `obirefidx` package implements **reference database indexing** for high-throughput sequencing data, specifically targeting family-level taxonomic classification. It supports efficient clustering and k-mer-based indexing of reference sequences.

## Core Functionalities

### 1. **Sequence Clustering**
- `MakeStartClusterSliceWorker()` performs greedy hierarchical clustering based on sequence similarity.
- Uses **LCSS (Longest Common Subsequence)** alignment with error tolerance derived from a user-defined identity threshold.
- Assigns each sequence:
  - `clusterid`: identifier of its cluster centroid (head).
  - `clusterhead`: boolean flag indicating if it is a representative.
  - `clusteridentity`: alignment-based identity score to the head.

### 2. **K-mer & Taxonomy-Based Indexing**
- `MakeIndexingSliceWorker()` builds per-sequence indexes using:
  - Precomputed **4-mer frequency tables** (`obikmer.Table4mer`).
  - Taxonomic annotations (family, genus, species) from a `Taxonomy`.
- Indexing is parallelized over chunks of 10 sequences using worker goroutines.

### 3. **Family-Level Reference Index Construction**
- `IndexFamilyDB()` orchestrates the full pipeline:
  - Loads and validates reference sequences.
  - Computes k-mer counts for each sequence.
  - Annotates taxonomy (family/genus/species) using helper workers (`MakeSet*Worker`).
  - Clusters sequences at **≥90% identity** (hardcoded threshold for family-level).
  - Re-indexes only cluster centroids to reduce redundancy.
- Final indexed references retain full taxonomic context and k-mer signatures.

## Implementation Highlights
- **Parallelization**: Leverages goroutines with configurable worker count (`obidefault.ParallelWorkers()`).
- **Memory Efficiency**: Processes sequences in chunks and reuses buffers.
- **Progress Tracking**: Optional progress bar via `progressbar/v3`.
- **Logging & Validation**: Uses Logrus for structured logging and panics on critical errors (e.g., missing taxonomy).

## Use Case
Enables rapid sequence similarity search and taxonomic assignment in metabarcoding pipelines by precomputing compact, clustered reference indexes.
