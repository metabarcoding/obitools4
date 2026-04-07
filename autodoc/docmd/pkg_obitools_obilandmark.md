# `obilandmark` Package: Semantic Documentation

The `obilandmark` package implements a **reference-free, landmark-based embedding and indexing pipeline** for biological sequences within the OBITools4 ecosystem. It enables scalable, low-dimensional representation of sequence libraries by projecting them into a distance space defined by curated landmark sequences—ideal for clustering, classification, and fast similarity search in metabarcoding or metagenomics workflows.

## Public Functionalities

### `MapOnLandmarkSequences(library, landmarks)`
Projects each sequence in a biological library onto Euclidean coordinates using pre-selected landmark sequences.  
- **Input**: A sequence `library` (e.g., FASTA/FASTQ iterator) and a list of landmark sequences.  
- **Algorithm**: Computes pairwise alignment scores between each sequence and all landmarks using `FastLCSScore`, converting them into distance-based coordinates.  
- **Output**: A matrix of shape `(n_sequences, n_landmarks)` where each row is a point in landmark space (`seqworld`).  
- **Features**: Parallel execution (configurable workers), progress bar, and buffered streaming for large datasets.

### `CLISelectLandmarkSequences(options)`
Main orchestration function that performs landmark selection, embedding, and annotation in a single CLI-driven pipeline.  
- **Landmark Selection**: Iteratively selects `n` landmarks (default: 200) via k-means clustering on initial random samples, minimizing cluster inertia over two refinement rounds.  
- **Embedding**: Calls `MapOnLandmarkSequences()` to compute coordinates for all sequences in the library.  
- **Annotation**: Augments each sequence record with:
  - `landmark_coord`: full coordinate vector (distances to all landmarks),
  - optional `landmark_id` for sequences selected as landmark representatives.  
- **Taxonomic Indexing**: If taxonomy is provided, builds a `GeomIndexSequence` per sequence—enabling efficient taxonomic search via geometric proximity.

### `LandmarkOptionSet(options)`
Registers CLI options specific to landmark configuration.  
- Adds the `-n` / `--center` flag (type: integer), defaulting to **200**, controlling the number of landmarks selected.

### `OptionSet(options)`
Aggregates option sets required by the pipeline:  
- Input/output handling (`obiconvert.InputOptionSet`, `.OutputOptionSet`)  
- Taxonomy loading support (optional, via `obioptions.LoadTaxonomyOptionSet`)  
- Landmark-specific options (`LandmarkOptionSet`)

### `CLINCenter()`
Returns the integer value of `-n / --center`, i.e., the number of landmarks to select (default: 200).

## Design Principles

- **Scalability**: Uses buffered I/O and parallel workers to process large sequence libraries efficiently.  
- **Modularity**: Integrates with core OBITools4 modules (`obialign`, `obistats`, `obiutils`, `obitax`, `obirefidx`).  
- **CLI-first**: Designed for batch processing pipelines; defaults ensure sensible behavior out-of-the-box.  
- **Extensibility**: Annotation schema supports future enhancements (e.g., `landmark_class` via commented stubs).

## Use Cases

- Reference-free sequence clustering and dimensionality reduction  
- Fast similarity search via geometric indexing in taxonomic space  
- Preprocessing for machine learning on sequence libraries (e.g., classification, anomaly detection)  

> **Note**: Only public interfaces are documented. Internal helpers (e.g., clustering utilities, alignment wrappers) remain implementation details.
