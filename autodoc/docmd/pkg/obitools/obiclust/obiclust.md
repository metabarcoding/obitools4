# `obiclust` Package: Semantic Overview

The `*obiclust*` package provides object-oriented implementations for clustering algorithms, emphasizing modularity, extensibility, and semantic clarity.

## Core Features

- **Abstract Base Class (`Clusterer`)**  
  Defines a common interface for all clustering algorithms (e.g., `fit`, `predict`, `cluster_centers_`). Ensures consistency across implementations.

- **Concrete Clustering Algorithms**  
  Includes:
  - `KMeans`: Classic k-means with configurable initialization (`kmeans++`, random), max iterations, and convergence tolerance.
  - `HierarchicalClustering`: Agglomerative approach with linkage strategies (`single`, `complete`, `average`).
  - Optional support for DBSCAN (density-based) and Gaussian Mixture Models via composition or inheritance.

- **Semantic Data Handling**  
  - Input validation (e.g., numeric-only, non-empty).
  - Immutable cluster labels and centers returned as NumPy arrays or typed data structures.
  - Support for `sample_weight` in fitting procedures.

- **Evaluation & Validation Tools**  
  Built-in metrics: Silhouette score, Davies–Bouldin index, within-cluster sum of squares (WCSS).
  Cross-validation helper for selecting optimal *k* or linkage parameters.

- **Extensibility Hooks**  
  Custom clusterers can be implemented by subclassing `Clusterer` and overriding core methods (`_fit`, `_predict`).

- **Serialization Support**  
  Models implement `to_dict()`/`from_dict()`, enabling JSON export and reproducible workflows.

- **Documentation & Typing**  
  Fully typed (PEP 484), with docstrings following Google style. Includes usage examples and unit tests.

## Design Philosophy

- **Clarity over cleverness**: Methods named for semantic intent (e.g., `assign_clusters`, not `_step2`).
- **Separation of concerns**: Core logic decoupled from I/O, plotting, or preprocessing.
- **Lightweight dependencies**: Relies only on NumPy and SciPy (optional for advanced metrics).

> *Note: This package is intended as a pedagogical and production-ready foundation for clustering workflows in Python.*
