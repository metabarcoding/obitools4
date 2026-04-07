# Semantic Description of `obiseq` Statistics and Merging Features

This package provides infrastructure for **tracking, aggregating, and merging statistical occurrences** of sequence attributes across biological sequences (`BioSequence`). It supports both **count-based and weighted statistics**, with thread-safe operations.

## Core Components

- `StatsOnValues`: A concurrent map (`map[string]int`) with R/W locking to store occurrence counts per attribute value (e.g., taxon, primer, quality bin).
- `StatsOnDescription`: Defines *how* to extract and weight statistics from a sequence (e.g., count per read, or sum of quality scores).
- `StatsOnSlotName(key)`: Generates internal annotation keys (e.g., `"merged_taxon"`) to store precomputed statistics.

## Key Functionalities

1. **Per-Sequence Statistics Initialization & Update**
   - `StatsOn(desc, na)`: Ensures a statistics slot exists for attribute `desc.Key`, initializes if needed.
   - `StatsPlusOne(...)`: Adds contribution of a *single* sequence to the statistics (e.g., increment count for its taxon).

2. **Thread-Safe Aggregation**
   - `Merge(*StatsOnValues)`: Safely merges counts from another `StatsOnValues`, used to combine per-sequence stats.

3. **Sequence Merging with Stat Propagation**
   - `BioSequence.Merge(...)`: 
     - Combines two sequences (e.g., consensus/overlap).
     - Updates statistics for specified attributes (`statsOn`), preserving or aggregating counts.
     - Resolves conflicting annotations by deleting non-merged fields if mismatched.

4. **Bulk Merging**
   - `BioSequenceSlice.Merge(...)`: Efficiently merges *N* sequences into one, recycling inputs and updating statistics incrementally.

## Use Cases

- Tracking taxonomic assignments across merged reads.
- Aggregating primer or barcode counts in amplicon merging.
- Summarizing quality scores, abundance weights, or custom metadata during consensus building.

## Design Notes

- Uses `sync.RWMutex` for safe concurrent access.
- Supports only JSON-marshalable, serializable statistics (via `MarshalJSON`).
- Enforces type safety: only strings/integers/booleans allowed for attribute values.
