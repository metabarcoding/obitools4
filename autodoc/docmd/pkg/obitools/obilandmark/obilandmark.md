# Semantic Description of `obilandmark` Package

The `obilandmark` package implements a **landmark-based sequence embedding and indexing pipeline**, primarily for large-scale biological sequence analysis.

## Core Functionality

- **`MapOnLandmarkSequences()`**: Projects each sequence in a library into a Euclidean space defined by *landmarks* (reference sequences).  
  - For each sequence, computes distances to all landmark sequences using the **FastLCSScore** alignment algorithm.  
  - Outputs a `library_size × n_landmarks` matrix of float coordinates (`seqworld`).  
  - Parallelized with configurable workers; supports progress bar visualization.

- **`CLISelectLandmarkSequences()`**: Orchestrates landmark selection, embedding, and annotation:
  - **Iteratively selects landmarks** via k-means clustering on initial random samples (2 rounds), refining clusters to minimize inertia.
  - **Annotates sequences** with:
    - `landmark_coord`: full coordinate vector (distances to all landmarks),
    - optional `landmark_id` for sequences selected as landmarks,
    - (commented-out) future support for `landmark_class`.
  - If taxonomy is available, builds a **geometric reference index** per sequence (`GeomIndexSesquence`) for efficient taxonomic search.

## Design Highlights

- **Scalable**: Uses buffered channels and parallel workers to handle large datasets.
- **Modular integration** with core OBItools4 components: alignment (`obialign`), statistics (`obistats`, `obiutils`), taxonomy (`obitax`), indexing (`obirefidx`).
- **CLI-ready**: Uses default settings (workers, progress bar) and integrates with batch iterators.

## Use Case

Enables **low-dimensional embedding of sequences** for downstream tasks (clustering, classification, indexing), especially useful in metabarcoding or metagenomics where reference-free representation and fast similarity search are critical.
