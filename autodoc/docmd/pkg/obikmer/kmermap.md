# Semantic Description of `KmerMap` Functionality

The provided Go package implements a **k-mer indexing and matching system** for biological sequences (`BioSequence`). It supports both standard and *sparse* k-mer representations (where one position is masked, typically for handling ambiguous bases or symmetry).

### Core Data Structures
- `KmerMap[T]`: A generic hash map associating *normalized* k-mers (type `T`, e.g., uint64 encoded in 2 bits per base) to lists of sequences containing them.
- `KmerMatch`: A map from sequence pointers to k-mer match counts, used for query results.

### Key Features
1. **K-mer Normalization**  
   - Handles both forward and reverse-complement k-mers.
   - Selects the lexicographically smaller representation (canonical form).
   - Supports *sparse* k-mers: when `SparseAt ≥ 0`, the central base is ignored (replaced by `#` in string view), and k-mers are symmetrically normalized.

2. **Efficient Indexing (`Push`)**  
   - Builds an index of all canonical k-mers from a set of sequences.
   - Optionally limits per-k-mer storage (`maxocc`), useful for filtering high-frequency k-mers (e.g., contaminants).

3. **Querying (`Query`)**  
   - Given a query sequence, returns all sequences in the index sharing k-mers with it.
   - Counts per-sequence how many shared k-mers exist (used for similarity estimation or clustering).

4. **Result Utilities (`KmerMatch`)**  
   - `FilterMinCount`: Remove low-count matches.
   - `Max()`, `Sequences()`: Retrieve best match or all matched sequences.

5. **Construction (`NewKmerMap`)**  
   - Automatically adjusts k-mer size: odd for sparse mode, even otherwise.
   - Precomputes bitmasks for efficient k-mer manipulation (masking, shifting).
   - Integrates progress bar during indexing.

### Use Cases
- Read clustering (e.g., OTU/ASV picking).
- Error correction via k-mer abundance.
- Sequence similarity search or contamination screening.

The implementation leverages low-level bit operations for performance and memory efficiency, especially critical in large-scale NGS data processing.
