# `obiclean` Package Functional Overview

The `obiclean` package implements a pipeline for cleaning and annotating high-throughput sequencing data, particularly focused on PCR amplicon error correction and chimera removal.

## Core Data Structures

- `seqPCR`: Represents a sequence within one PCR/sample, tracking:
  - raw read count (`Count`)
- post-clustering abundance weight (`Weight`)
- sequence pointer, edges to parent/child variants (for mutation graph), and cluster membership.

## Key Functionalities

### 1. **Sample-wise Aggregation**
- `buildSamples`: Distributes sequences across samples using metadata tags, storing per-sample counts.

### 2. **Graph Construction & Filtering**
- `BuildSeqGraph`: Builds a mutation graph (edges = point mutations) across samples.
- `FilterGraphOnRatio`: Removes low-abundance variants based on abundance ratio thresholds.

### 3. **Annotation & Status Assignment**
- `annotateOBIClean`: Adds per-sequence annotations:
  - `"obiclean_head"`: Boolean indicating if sequence is a cluster head (i.e., not derived from another).
  - `"obiclean_singletoncount"`, `"internalcount"`, `"headcount"`: Counts of sequences in each status category.
- `Status`/`Weight`: Getter/setter functions for sample-specific annotations (`obiclean_status`, `obiclean_weight`).

### 4. **Mutation & Cluster Tracking**
- `GetMutation`: Retrieves or initializes mutation map (e.g., `"A->T@42"`).
- `Mutation`: Populates mutation annotations based on graph edges.
- `GetCluster`/`Status`: Manage per-sample cluster membership and status labels (`h`=head, `i`=internal node, `s`=singleton).

### 5. **Filtering & Output**
- CLI-driven filtering options:
  - `OnlyHead`: Keep only cluster heads.
  - `NotAlwaysChimera`: Exclude sequences flagged as chimera in *all* samples.
  - `MinSampleCount`: Retain only sequences appearing ≥ N times across samples.

### 6. **Optional Outputs**
- `AnnotateChimera`: Adds chimera flags (if enabled).
- Graph export to GML files (`SaveGMLGraphs`), ratio tables, and empirical distribution CSV.

## Design Highlights

- Batch processing with progress bars.
- Extensive use of sequence annotations (not in-place modification).
- Flexible type coercion for annotation values (`interface{}` → typed maps).

This module is part of the OBITools4 ecosystem for NGS data processing.
