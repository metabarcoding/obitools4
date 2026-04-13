# `obik summary`: K-mer Index Metadata and Statistics Tool

The `runSummary` function provides semantic insights into a k-mer index stored in the file system. It inspects and aggregates metadata from an `obikmer.KmerSetGroup`, summarizing structure, content size, and inter-set similarity.

## Core Functionality

- **Index Validation & Opening**: Opens a k-mer index directory using `obikmer.OpenKmerSetGroup`, returning an error if invalid or inaccessible.

- **Structural Summary**: Collects global properties:  
  - `k`, `m`: K-mer length and bloom filter bits per element.  
  - Partitions, total sets (`Size()`), and cumulative k-mer count (`Len()`).  
  - Total disk footprint across all files.

- **Per-set Statistics**: For each set (partitioned k-mer collection), records:  
  - `index`, unique ID, count of distinct kmers.  
  - Disk usage (summed over all `.kdi` files in its partition).  
  - Optional metadata (`map[string]interface{}`).

- **Disk Usage Estimation**: `computeSetDiskSize` recursively sums file sizes of all partition files for each set, ensuring accurate storage reporting.

- **Jaccard Similarity Matrix** *(optional)*: When enabled (`_jaccard` flag), computes pairwise Jaccard distances between sets via `JaccardDistanceMatrix()`, stored as an *n×n* symmetric matrix.

- **Multi-format Output**: Supports JSON (default), YAML, and CSV exports for interoperability with downstream tools.

## Semantic Use Cases

- **Index auditing**: Verify integrity and size of large-scale k-mer collections.  
- **Resource planning**: Estimate storage needs or detect anomalies in disk usage per set.  
- **Comparative analysis**: Use Jaccard matrix to assess overlap between experimental replicates or sample groups.  
- **Pipeline integration**: CSV output enables quick parsing in spreadsheets, dashboards, or CI/CD checks.

All outputs preserve metadata fields (e.g., sample annotations), supporting reproducibility and traceability.
