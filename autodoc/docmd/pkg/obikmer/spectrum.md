# K-mer Spectrum Analysis Package (`obikmer`)

This Go package provides tools for analyzing k-mer frequency distributions in biological sequences.

## Core Data Structures

- **`SpectrumEntry`**: Represents a bin in the k-mer frequency spectrum:  
  `Frequency`: how often a k-mer was observed; `Count`: number of distinct k-mers with that frequency.

- **`KmerSpectrum`**: A sorted list of non-zero `SpectrumEntry`s (ascending by frequency), enabling efficient statistics and serialization.

## Key Functionalities

### Spectrum Management
- `MapToSpectrum()` / `ToMap()`: Convert between map and structured spectrum representations.
- `MergeSpectraMaps()` / `MergeTopN()`: Combine spectral or top-k data from multiple sources.
- `MaxFrequency()` returns the highest observed k-mer count.

### I/O & Persistence
- Binary format (`KSP\x01` magic header) with varint encoding for compact storage:
  - `WriteSpectrum()` / `ReadSpectrum()`: Save/load full spectra to disk.
- CSV export:
  - `WriteTopKmersCSV()`: Outputs top-k k-mers with their sequences (decoded from uint64) and frequencies.

### Top-N K-mer Tracking
- Uses a **min-heap** to efficiently maintain the *N most frequent* k-mers in streaming scenarios:
  - `NewTopNKmers(n)`: Initialize collector.
  - `Add(kmer, freq)`: Insert/update while respecting capacity *n*.
  - `Results()`: Return top-kmers sorted descending by frequency.

## Design Highlights
- Memory-efficient: Uses `uint64` for k-mers (suitable up to *k* ≤ 32).
- Streaming-friendly: Top-N collector supports incremental updates.
- Thread-safety note: External synchronization required for concurrent access.

