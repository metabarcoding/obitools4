# `obiclust` Package: Semantic Overview

The `*opicluster/obiclean*` module provides command-line clustering functionality for biological sequence data (e.g., amplicons, OTUs/ASVs), integrating alignment-based similarity and abundance-aware heuristics.

## Core Clustering Logic
- **Distance/Score Mode**: Switches between alignment *similarity* (default) and *distance*-based clustering (`--distance`).
- **Normalization Strategy**: Controls how alignment scores are normalized:
  - `NoNormalization`: raw score.
  - `NormalizedByShortest` (`--shortest`)
  - `NormalizedByLongest` (`--longest`)
  - `NormalizedByAlignment` (default, via `--alignment`) — uses aligned length.
- **Clustering Algorithm**: Supports both *exact* (`--exact`, optimal but slower) and greedy heuristics (default).

## Input & Sample Handling
- **Sample Attribute**: Configurable metadata field (`--sample`, `-s`) to group sequences by sample origin.
- **Minimum Sample Support**: Filters out sequences appearing in fewer than `--min-sample-count` samples.
- **Sequence Ordering**:
  - By length (`--length-ordered`) or abundance (`--abundance-ordered`).
  - Optional ascending sort order (`--ascending-sorting`) — default is descending.

## Abundance-Based Refinement
- **Ratio Threshold** (`--ratio`, `-r`): Merges low-abundance sequences into high-abundance parents if their count ratio ≤ threshold.
- **Head Selection** (`--head`, `-H`): Restricts output to sequences flagged as “head” in at least one sample (e.g., representative centroids).

## Output & Diagnostics
- **Graph Export** (`--save-graph`): Dumps the clustering DAG in GraphML format for inspection/debugging.
- **Ratio Table** (`--save-ratio`): Saves edge abundance ratios (CSV) to analyze clustering confidence.
- **Threshold Control** (`--distance`, `--threshold`): Sets the max distance/similarity cutoff to merge sequences into a cluster.

## Integration
- Extends `obiconvert` I/O options (input/output formats), enabling seamless pipeline integration.
