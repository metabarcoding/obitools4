# Super K-mers Extraction Module (`obikmer`)

This Go package provides efficient tools for extracting **super k-mers** from DNA sequences using *minimizer-based sliding windows*. Super k-mers are maximal contiguous subsequences sharing the same minimal canonical minimizer in a window of size `k`.

## Core Functionality

- **`IterSuperKmers(seq, k, m)`**  
  Returns an iterator over `SuperKmer` structs. Each struct contains:
  - `Start`, `End`: genomic positions of the super k-mer in the original sequence  
  - `Minimizer`: canonical minimizer value (uint64) for that segment  
  - `Sequence`: the actual DNA subsequence  

- **`SuperKmer.ToBioSequence(...)`**  
  Converts a raw `SuperKmer` into an enriched `obiseq.BioSequence`, embedding metadata:
  - ID: `{parentID}_superkmer_{start}_{end}`  
  - Attributes: minimizer sequence (`minimizer_seq`), value, `k`, `m`, positions, and parent ID  

- **`SuperKmerWorker(k, m)`**  
  A `SeqWorker` adapter for pipeline integration (e.g., with `obiiter`). Processes a full BioSequence and returns all extracted super k-mers as a slice of `BioSequence`s.

## Algorithm Highlights

- Uses **canonical minimizers** (forward/reverse-complement minimum) to ensure strand-invariance  
- Maintains a monotonic deque for efficient *sliding-window minimizer* tracking (O(n) time complexity)  
- Supports DNA bases `A/C/G/T/U` case-insensitively via bitmasking (`seq[i] & 31`)  
- Enforces parameter constraints: `1 ≤ m < k ≤ 31`, sequence length ≥ `k`

## Use Cases

- Read partitioning in metagenomics (e.g., for error correction or clustering)  
- Efficient k-mer space segmentation without storing all individual kmers  
- Integration into modular bioinformatics pipelines via `SeqWorker` interface
