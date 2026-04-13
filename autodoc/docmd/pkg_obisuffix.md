# obisuffix: Suffix Array Package for Biological Sequence Analysis

The `obisuffix` package implements a suffix array tailored to biological sequences, enabling efficient lexicographic ordering and prefix analysis across multiple input sequences. It supports DNA, RNA, and protein data via integration with `obiseq.BioSequenceSlice`, making it suitable for repeat detection, k-mer mining, and alignment-free comparison workflows.

## Core Data Structures

### `Suffix`
Represents a single suffix by storing:
- `Idx int`: Index of the source sequence in the input slice.
- `Pos int`: Starting position (0-based) within that sequence.

### `SuffixArray`
Encapsulates:
- `Data []Suffix`: Sorted list of all suffixes.
- `Sequences obiseq.BioSequenceSlice`: Original input sequences (immutable reference).
- `Common []int`: Cached longest common prefix lengths between adjacent suffixes (`Data[i]` and `Data[i+1]`). Lazily computed.

## Public Functions

### `BuildSuffixArray(data obiseq.BioSequenceSlice) *SuffixArray`
Constructs a suffix array from one or more biological sequences:
- Enumerates **every** suffix of every sequence (i.e., for a sequence `s`, adds all `(Idx, Pos)` where `0 ≤ Pos < len(s)`).
- Sorts suffixes lexicographically using a deterministic comparator (`SuffixLess`):
  - Primary: Compare nucleotide/amino-acid content character-by-character.
  - Tie-breakers (if prefixes match up to min length):
    1. Shorter suffix comes first.
    2. Lower `Idx` (sequence index).
    3. Earlier `Pos`.
- Precomputes and caches the common-prefix array via internal call to `CommonSuffix()`.

### `(*SuffixArray) CommonSuffix() []int`
Computes the length of the longest common prefix (LCP) between each adjacent pair in `Data`:
- Returns a slice of length `len(Data)-1`, where `Common[i] = LCP(Data[i], Data[i+1])`.
- Uses memoization: If already computed (e.g., after `BuildSuffixArray`), returns the cached result.
- Avoids redundant comparisons by leveraging sorted order and early termination.

### `(*SuffixArray) String() string`
Returns a formatted, human-readable table for inspection:
- Columns: `Common`, `Idx`, `Pos`, and the actual suffix string (via `.Substring()`).
- Useful for debugging, educational demos, or visualizing repeat patterns and overlaps.

## Semantic Guarantees & Design Choices

- **Deterministic ordering**: Tie-breaking rules ensure reproducibility across runs and platforms.
- **Memory efficiency**: Stores only indices (not copies of suffixes), critical for large genomic datasets.
- **Biological fidelity**: Respects alphabet semantics (e.g., `A < C < G < T` for DNA) via underlying sequence comparison.
- **Lazy evaluation**: `CommonSuffix()` is invoked only when needed (e.g., on first call to `.String()`, or explicitly), avoiding unnecessary work.
- **Transparency**: All public fields are accessible, enabling downstream analysis without encapsulation barriers.

## Typical Use Cases

- Detecting tandem repeats or low-complexity regions across multi-sequence datasets.
- Building suffix arrays for *de novo* assembly validation or error correction.
- Serving as a building block in alignment-free metrics (e.g., Jaccard similarity over shared *k*-mers).
- Supporting pattern mining in metagenomic or pangenome collections.

> **Note**: This package focuses on *exact* suffix matching; probabilistic or approximate extensions are out of scope.
