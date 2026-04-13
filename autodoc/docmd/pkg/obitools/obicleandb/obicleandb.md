# `obicleandb` Package Overview

The `obicleandb` package provides semantic sequence curation and trust scoring for biological sequences (e.g., DNA barcodes) within the OBITools4 framework. It integrates taxonomic, alignment-based, and statistical methods to assess sequence reliability.

## Core Functionalities

1. **Taxonomic Filtering & Dereplication**  
   - Input sequences are first dereplicated (collapsed by identity) under taxonomic constraints (`taxid`), ensuring only unique sequences per taxon are retained.
   - Sequences must meet minimum taxonomy requirements (species, genus, family) or CLI-specified ranks.

2. **Taxonomic Annotation**  
   - Sequences are annotated with species, genus, and family tax IDs using the default taxonomy.

3. **Trust Scoring via Statistical Testing**  
   - Two complementary trust mechanisms are implemented:
     - `SequenceTrust`: Assigns a *local* confidence score based on sequence count (`1 − 1/(n+1)`), treating duplicates as evidence of reliability.
     - `SequenceTrustSlice`: Computes pairwise alignment-based distances (LCSS score ratio), then derives a *global* trust metric using median-normalized scores and effective group size estimation.

4. **Family/Genus-Level Discrimination (Mann–Whitney U Test)**  
   - `MakeSequenceFamilyGenusWorker` evaluates whether a query sequence is significantly closer (in alignment score) to conspecifics than to outgroup sequences.
   - Compares intra-genus/family distances vs. inter-family distances using fast LCS-based alignment (`obialign.FastLCSScore`).
   - Returns a *p*-value stored in `obicleandb_trusted`, indicating confidence that the sequence belongs to its assigned higher-rank taxon.

5. **Efficient Pairwise Distance Computation**  
   - `diagCoord` implements a compact triangular matrix indexing scheme to store only upper-triangle distances, minimizing memory usage.

6. **Pipeline Integration**  
   - `ICleanDB` orchestrates the full workflow: filtering → dereplication → annotation → trust scoring, returning a cleaned and trusted sequence iterator.

## Key Attributes Set

| Attribute | Meaning |
|----------|---------|
| `obicleandb_trusted` | Final confidence score (probability of correct taxonomic assignment) |
| `obicleandb_trusted_on` | Effective sample size used for scoring (e.g., weighted group count) |
| `obicleandb_level` | Taxonomic level used for comparison (`genus`, `family`, or `none`) |
| `obicleandb_median` | Median pairwise distance used as baseline for normalization |

## Design Principles

- **Parallelism**: Leverages batched, parallel workers via `obidefault` settings.
- **Modularity**: Workers are composable and reusable (e.g., `MakeSequenceFamilyGenusWorker`).
- **Robustness**: Handles edge cases (e.g., small sample sizes, missing taxonomy) gracefully.

This package supports high-throughput DNA metabarcoding pipelines by rigorously filtering and scoring sequences before downstream analysis (e.g., OTU clustering, diversity estimation).
