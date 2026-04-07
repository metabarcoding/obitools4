# Semantic Description of `obikmer` Package Functionalities

The `obikmer` package provides disk-backed operations on *k*-mer sets derived from biological sequences. It supports scalable set algebra and similarity computations via the `KmerSetGroup` type.

## Core Features

- **Sequence-to-*k*-mer Indexing**: Sequences are converted into *k*-mers (of length `k`) and stored in a group of sets (`KmerSetGroup`), with one set per sequence. Minimizer-based sampling (parameter `m`) reduces redundancy.

- **Set Operations on Disk**: Efficient disk-resident implementations of standard set operations:
  - `Union`: Merges all *k*-mers from selected sets.
  - `Intersect`: Retains only *k*-mers present in all input sets.
  - `Difference` (`A \ B`): Keeps *k*-mers present in set A but not in B.
  - `QuorumAtLeast(r)`: Returns *k*-mers appearing in ≥`r` sets (generalizes union (`r=1`) and intersection (`r=n`)).

- **Consistency Guarantees**: Operations obey mathematical identities (e.g., `|A ∪ B| = |A| + |B| − |A ∩ B|`), validated via unit tests.

- **Similarity & Distance Metrics**:
  - `JaccardDistanceMatrix()`: Computes pairwise Jaccard *distances* (1 − similarity) between all sets.
  - `JaccardSimilarityMatrix()`: Computes pairwise Jaccard *similarities* (`|A ∩ B| / |A ∪ B|`).
  - Identical sets yield distance = `0.0`, disjoint ones give `1.0`; similarity is complementary.

## Design Principles

- **Temporary Directory Usage**: All operations use OS temp dirs for isolation and cleanup.
- **Testing-Focused API**: Helper functions (`buildGroupFromSeqs`, `collectKmers`) simplify test setup.
- **Scalability**: Disk-backed design avoids memory overflow for large sequence collections.

This package enables robust, reproducible *k*-mer set analysis in bioinformatics pipelines—especially useful for metagenomic binning, error correction, or read clustering.
