# OBITag Taxonomic Identification Module

This Go package (`obitag`) provides tools for **taxonomic assignment of biological sequences** using reference databases and alignment-based similarity scoring.

## Core Functionalities

- **`MatchDistanceIndex`**:  
  Maps a distance value to the closest taxonomic entry in an indexed map (`distanceIdx`). It performs binary search on sorted distance keys and returns the corresponding taxon (taxid, rank, scientific name). Falls back to root if no match is found.

- **`FindClosests`**:  
  Identifies the most similar reference sequences to a query sequence using:
  - **4-mer frequency overlap** (`Common4Mer`) for fast pre-screening.
  - **LCS-based alignment scoring** (Longest Common Subsequence) for precise similarity measurement.
  - Returns top matches, edit distance (`maxe`), sequence identity score, best match ID, and indices.

- **`Identify`**:  
  Performs full taxonomic classification:
  - Uses `FindClosests()` to retrieve best matching references.
  - Leverages precomputed reference indices (`OBITagRefIndex`) to resolve taxonomic assignments per distance level.
  - Computes the **Lowest Common Ancestor (LCA)** of all matching taxa to assign robust taxonomy.
  - Marks unidentifiable sequences with root taxon and sets metadata attributes (rank, identity %, match count).

- **`IdentifySeqWorker`**:  
  Wraps `Identify()` into a reusable sequence worker function for batch processing.

- **`CLIAssignTaxonomy`**:  
  High-level CLI entry point:
  - Filters and indexes reference sequences (4-mer counting, taxon validation).
  - Builds a `SeqWorker` pipeline for parallel execution.
  - Supports logging, filtering of invalid references, and configurable concurrency.

## Key Features

- **Hybrid speed/accuracy**: Uses k-mer pre-screening + LCS alignment.
- **Index caching**: Reuses taxonomic indexes per reference to avoid recomputation.
- **Robustness**: Gracefully handles missing taxonomy data and invalid inputs via fallbacks to root.
- **Extensibility**: Designed for integration into larger OBITools4 pipelines.

## Dependencies

Uses core modules from `obitools4`: sequence (`obiseq`), taxonomy (`obitax`), alignment (` obialign`), k-mer analysis (`obikmer`), iteration utilities, and logging.
