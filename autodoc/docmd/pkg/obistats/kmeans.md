# `obistats` Package: K-Means Clustering Implementation

The `obistats` package provides a concurrent, type-generic implementation of the **K-means clustering algorithm** for numerical datasets.

## Core Utilities
- `SquareDist` / `EuclideanDist`: Compute squared and Euclidean distances between vectors (generic over `float64` or `int`).  
- `DefaultRG`: Returns a seeded random number generator (`*rand.Rand`) for reproducibility control.

## Data Structure
- `KmeansClustering`: Encapsulates dataset (`*obiutils.Matrix[float64]`), cluster assignments, centers, and metadata (sizes, distances to nearest center).  
- Supports dynamic addition of clusters via `AddACenter()`.

## Initialization & Management
- `MakeKmeansClustering`: Initializes the structure with data, number of clusters *k*, and RNG.  
- `SetCenterTo`, `AddACenter`: Assign or grow centers; uses **k-means++**-inspired weighted sampling for new centers.  
- `ResetEmptyCenters`: Reinitializes empty clusters using distance-weighted sampling.

## Core Algorithm Steps
- `AssignToClass`: Parallel assignment of points to nearest centers (uses goroutines + mutex).  
- `ComputeCenters`: Computes new cluster centroids *as the closest original data point* to the arithmetic mean (robust for non-Euclidean spaces).  
- `Run`: Executes iterative refinement until convergence (`max_cycle` iterations or inertia drop ≤ threshold).

## Accessors & Diagnostics
- `K()`, `N()`, `Dimension()`: Return number of clusters, dataset size, and feature dimension.  
- `Inertia()`: Sum of squared distances to assigned centers (convergence metric).  
- `Centers`, `Classes`, `Sizes`: Expose internal clustering state.

## Design Highlights
- Fully concurrent (goroutine-based) for performance.  
- Generic distance functions support both `int` and `float64`.  
- Explicit handling of edge cases (empty clusters, convergence).  
- Logging via `logrus` for debugging (`obilog.Warnf`).  

> *Note: High-level wrapper functions (e.g., standalone `Kmeans`) are commented out but outline intended API usage.*
