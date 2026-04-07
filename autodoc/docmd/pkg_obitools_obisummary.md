# `obisummary` Package: Semantic Description

The `obisumsummary` package delivers lightweight, high-performance statistical summarization of biological sequence data processed by OBITools4. It enables rapid profiling of metadata and content-level features across large sequence sets—especially useful post-processing (e.g., after `obiclean` or merging)—while supporting parallel execution for scalability.

## Core Data Model

- **`DataSummary` struct**: Central container tracking:
  - Global metrics: number of reads, unique variants (distinct sequences), and total symbols.
  - Presence flags for special annotations: `merged_sample`, `obiclean_status`/`weight`.
  - Categorized annotation metadata:
    - Scalar attributes (single-value per sequence).
    - Map-like tags (`map_tags`), where each key maps to counts.
    - Vector or vector-like attributes (multi-value per sequence).
  - Per-sample statistics: variant count, singleton detection, and `obiclean`-related flags (e.g., bad reads).

## Low-Level Helpers

- **Map aggregation utilities**:
  - `sumUpdateIntMap`: Accumulates integer values across maps.
  - `countUpdateIntMap`, `plusOne/PlusUpdateIntMap`: Increment counters for keys (e.g., attribute or sample names).

- **`Add()` method**: Thread-safe merge of two `DataSummary`s—enables parallel accumulation.

## Main Processing Logic

- **`Update()` method**: Processes one `BioSequence`, updating:
  - Read count (via `.Count()`) and sequence-level metrics.
  - Variant detection via unique sequences; symbol count (total length).
  - Sample-aware logic: detects `merged_sample` or per-sample annotations to populate sample-level stats (e.g., singleton identification).
  - Annotation classification: routes keys into scalar, map, or vector buckets.

- **`ISummary()` function**: Parallel summarization engine:
  - Distributes work across `nproc` goroutines.
  - Aggregates partial summaries via atomic operations (`Add()`).
  - Returns a structured map with:
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

## CLI Integration (`obisummary` subpackage)

- **Option registration**:
  - `SummaryOptionSet()`: Registers flags for output format (`--json-output`, `--yaml-output`) and map attributes to summarize (`-map <attr>`).
  - `OptionSet()`: Extends above with input-handling options (e.g., file/iterator sources) from `obiconvert`.

- **Runtime introspection**:
  - `CLIOutFormat()`: Returns `"yaml"` (default) or `"json"`, with YAML only active if JSON is *not* requested.
  - `CLIHasMapSummary()` / `CLIMapSummary()`: Check and retrieve requested map attributes.

- **Design notes**:
  - Uses global state (e.g., `__json_output__`, `__map_summary__`) for compatibility with [`go-getoptions`](https://github.com/DavidGamba/go-getoptions).
  - Scope strictly limited to CLI configuration—no data processing logic resides here.
