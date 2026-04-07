# K-mer Set Group Builder — Semantic Description

This Go module (`obikmer`) provides a **disk-backed builder and accessor** for managing *k-mer sets* across multiple biological sequence datasets. It supports efficient construction, persistence, and querying of canonical *k*-mers (accounting for DNA reverse-complement symmetry), with optional frequency filtering.

### Core Functionalities

- **K-mer Set Group Construction**:  
  `NewKmerSetGroupBuilder` creates a builder configured with:
  - *k* (k-mer length),
  - *m* (minimal unique substring for partitioning),
  - number of sets (`nSets`),
  - and optional parameters like `WithMinFrequency`.

- **Sequence Ingestion**:  
  Sequences are added per set via `AddSequence(setID, bioseq)`. Internally:
  - Canonical *k*-mers are extracted (using `IterCanonicalKmers`),
  - Deduplicated and optionally filtered by occurrence frequency.

- **Persistence & Round-Trip**:  
  `builder.Close()` materializes the *k*-mer sets to disk (in temp or specified directory).  
  `OpenKmerSetGroup(dir)` reloads them — preserving all metadata and structure.

- **Metadata & Attributes**:  
  Supports custom identifiers (`SetId`) and key-value attributes (e.g., `"organism": "test"`), saved to disk via `SaveMetadata`.

- **Efficient Iteration**:  
  The iterator (`ksg.Iterator(setID)`) yields *sorted*, deduplicated canonical *k*-mers — using a k-way merge across internal partitions.

- **Frequency Filtering**:  
  `WithMinFrequency(n)` ensures only *k*-mers appearing ≥*n* times across inputs survive — enabling noise suppression (e.g., in error correction or abundance-based filtering).

- **Multi-set Support**:  
  Handles multiple independent *k*-mer sets (e.g., per sample or taxonomic group), verified via `Size()` and indexed access (`Len(setID)`).

### Testing Coverage

Comprehensive unit tests validate:
- Basic construction & correctness,
- Multi-sequence ingestion and deduplication,
- Frequency-based inclusion/exclusion logic,
- Cross-set isolation (`nSets > 1`),
- Metadata round-trip integrity.

This module is designed for scalable, reproducible *k*-mer indexing in metagenomic or amplicon analysis pipelines (e.g., OBITools4 ecosystem).
