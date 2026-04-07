# Semantic Overview of `obiseq` Package Functionalities

This Go package (`obiseq`) provides memory-efficient utilities for managing slices and annotations—key data structures in biosequence processing.

## Slice Management

- **`GetSlice(capacity int) []byte`**  
  Retrieves a reusable `[]byte` with ≥ requested capacity. For capacities ≤1024 bytes, it pulls from a `sync.Pool` (`_BioSequenceByteSlicePool`). Larger slices are freshly allocated.

- **`RecycleSlice(s *[]byte)`**  
  Clears and recycles small slices (≤1024 bytes) back to the pool. For large slices (≥100 KB), it nils them and triggers explicit `runtime.GC()` every ~256 MB of discarded memory to prevent heap bloat.

- **`CopySlice(src []byte) []byte`**  
  Efficiently copies a source slice into a pooled or newly allocated destination, preserving semantics without unnecessary allocations.

## Annotation Management

- **`BioSequenceAnnotationPool`**  
  A `sync.Pool` for reusable map-based annotations (`map[string]string`, inferred from usage), initialized with capacity 1.

- **`GetAnnotation(values ...Annotation) Annotation`**  
  Fetches an annotation map from the pool, optionally pre-populated via shallow copy of input annotations using `obiutils.MustFillMap`.

- **`RecycleAnnotation(a *Annotation)`**  
  Clears all keys from an annotation map and returns it to the pool for reuse.

## Design Rationale

The package prioritizes low-latency, high-throughput scenarios (e.g., NGS data pipelines) by minimizing GC pressure via:
- Tiered pooling strategy (`small` vs `large`)
- Explicit garbage collection triggers for large-object churn
- Safe reuse patterns avoiding aliasing or stale references

All operations are thread-safe via `sync.Pool` and atomic counters.
