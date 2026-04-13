# `obiclust` Package: Semantic Overview

The `*obiclust*` package provides object-oriented implementations for clustering algorithms, emphasizing modularity, extensibility, and semantic clarity—while `*opicluster/obiclean*` extends this to biological sequence data (e.g., amplicons, OTUs/ASVs), integrating alignment-aware similarity and abundance-sensitive heuristics.

## Core Clustering Infrastructure (`obiclust`)

### Abstract Base Class: `Clusterer`
- Defines a unified interface for all clustering algorithms.
- Public methods:
  - `fit(X, sample_weight=None)`: Learns cluster structure from data.
  - `predict(X)`: Assigns each sample to the nearest cluster (returns NumPy array of labels).
  - `cluster_centers_`: Immutable attribute storing learned centroids.
- Designed for subclassing: custom clusterers override `_fit()` and `_predict()`.
  
### Concrete Algorithms
- **`KMeans`**
  - Configurable initialization: `kmeans++`, random.
  - Parameters: max iterations, convergence tolerance (`tol`).
- **`HierarchicalClustering`**
  - Agglomerative strategy with linkage options: `single`, `complete`, `average`.
- *(Optional extensions)* DBSCAN, GaussianMixture via composition or inheritance.

### Semantic Data Handling
- Input validation: numeric-only matrices, non-empty inputs.
- Outputs are immutable NumPy arrays (labels/centers).
- Supports per-sample weights during fitting.

### Evaluation & Validation
- Built-in metrics: Silhouette score, Davies–Bouldin index, WCSS.
- Cross-validation helper (`select_k`, `tune_linkage`) for hyperparameter selection.

### Serialization & Typing
- `to_dict()` / `from_dict()`: Enables JSON persistence and reproducibility.
- Fully typed (PEP 484), Google-style docstrings, and usage examples included.

### Design Principles
- **Readability**: Method names reflect intent (e.g., `assign_clusters`, not `_step2`).
- **Separation of concerns**: Core logic decoupled from plotting, I/O, or preprocessing.
- **Minimal dependencies**: NumPy (required), SciPy (optional for metrics).

## Biological Sequence Clustering (`opicluster/obiclean`)

### Distance/Similarity Mode
- Switches between:
  - **Similarity mode** (default): higher scores = more related.
  - **Distance mode** (`--distance`): lower distances = closer.

### Normalization Strategies
Controls how alignment scores are scaled before clustering:
- `NoNormalization`: raw score.
- `NormalizedByShortest` (`--shortest`)
- `NormalizedByLongest` (`--longest`)
- `NormalizedByAlignment` (default, via `--alignment`) — uses aligned length.

### Clustering Strategy
- **Exact clustering** (`--exact`): optimal but computationally heavy.
- Greedy heuristic (default) for scalability.

### Sample-Aware Processing
- Groups sequences by sample origin (`--sample`, `-s`).
- Filters low-sample-count variants via `--min-sample-count`.
- Ordering options:
  - By length (`--length-ordered`) or abundance (`--abundance-ordered`).
  - Optional ascending sort: `--ascending-sorting`.

### Abundance Refinement
- **Ratio-based merging** (`--ratio`, `-r`): merges low-abundance sequences into high-abundance parents if their ratio ≤ threshold.
- **Head selection** (`--head`, `-H`): outputs only sequences flagged as “representative” in ≥1 sample.

### Output & Diagnostics
- **Graph export** (`--save-graph`): DAG in GraphML format (for debugging).
- **Ratio table export** (`--save-ratio`): CSV of edge abundance ratios.
- Threshold control via `--distance`, `--threshold`.

### Pipeline Integration
- Extends I/O options from `obiconvert`: seamless FASTA/FASTQ input/output, compatible with standard NGS pipelines.
