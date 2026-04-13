# Semantic Description of `IUniqueSequence` Functionality

The `IUniqueSequence` function performs **dereplication** of biological sequence data — i.e., grouping identical or near-identical sequences while preserving metadata and counts. It operates on an `obiiter.IBioSequenceBatch` iterator.

## Core Workflow

1. **Input Processing**  
   Accepts an input sequence iterator and optional configuration via `WithOption`.

2. **Parallelization Strategy**  
   Supports configurable parallel workers (`nworkers`). When `SortOnDisk()` is enabled, it falls back to single-threaded processing for disk-based sorting.

3. **Data Splitting Phase**  
   - Uses `HashClassifier` to partition input into buckets (controlled by `BatchCount`).  
   - Ensures deterministic chunking for reproducibility.

4. **Storage Choice**  
   - *In-memory*: via `ISequenceChunkOnMemory`.  
   - *Disk-based*: via `ISequenceSubChunk` + external sorting (requires single worker).

5. **Uniqueness Classification**  
   - Builds a composite classifier combining:
     - Sequence identity (`SequenceClassifier`)
     - Optional annotation categories (e.g., sample, primer), with NA handling.
   - If no annotations are specified, only raw sequence identity is used.

6. **Singleton Filtering**  
   Optionally excludes singleton reads (count = 1) via `NoSingleton()` option.

7. **Parallel Dereplication**  
   - Spawns worker goroutines to process chunks.
   - Each worker applies `ISequenceSubChunk` + deduplication logic per classifier group.

8. **Output Merging**  
   - Aggregates results using `IMergeSequenceBatch`, preserving:
     - Sequence counts
     - Statistics (if enabled)
     - NA handling and ordering

## Key Features

- **Scalable**: Supports both memory-efficient (disk) and high-speed (RAM) modes.
- **Configurable**: Via functional options (`Options`).
- **Thread-safe**: Uses `sync.Mutex` for deterministic ordering.
- **Metadata-aware**: Incorporates annotation-based grouping (e.g., sample, primer).
