# Semantic Description of `obisummary` Package

The `obisummary` package provides statistical summarization of biological sequence data processed by OBITools4. It aggregates metadata and content-level features across sequences in an iterator, supporting parallel execution.

## Core Components

- **`DataSummary` struct**: Holds counters and maps tracking:
  - Global counts: reads, variants (unique sequences), total symbols.
  - Presence flags for special attributes (`merged_sample`, `obiclean_status/weight`).
  - Per-attribute-type counts: scalar, map (`map_tags`), and vector/vector-like tags.
  - Per-sample statistics (variant count, singleton counts, bad `obiclean` flags).

- **Helper functions**:
  - `sumUpdateIntMap`, `countUpdateIntMap`: Aggregate or increment map values.
  - `plusOne/PlusUpdateIntMap`: Increment specific keys.

- **`Add()` method**: Merges two `DataSummary`s (thread-safe accumulation).

## Main Functionality

- **`Update()` method**: Processes a single `BioSequence`, updating internal counters:
  - Reads count (via `.Count()`), variant and symbol counts.
  - Detects `merged_sample` or single-sample annotations to populate sample-level stats (e.g., singleton detection).
  - Classifies annotation keys into scalar, map, or vector categories.

- **`ISummary()` function**:
  - Parallelizes summarization across `nproc` workers using goroutines.
  - Aggregates partial summaries and returns a structured dictionary with:
    ```json
    {
      "count": { "variants", "reads", "total_length" },
      "annotations": {
        "scalar_attributes",
        "map_attributes",
        "vector_attributes",
        "keys": { scalar: {...}, map: {...}, vector: {...} }
      },
      "samples": {
        "sample_count",
        "sample_stats": { sample_name: { reads, variants, singletons [, obiclean_bad] } }
      }
    }
    ```

## Use Case

Designed for lightweight, high-performance profiling of sequence datasets (e.g., after `obiclean`, merging), enabling quick quality checks and metadata exploration in OBITools4 pipelines.
