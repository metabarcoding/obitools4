# Semantic Description of `obikmer` Entropy Functions

The `obikmer` package provides high-performance tools to compute **Shannon entropy** for DNA *k*-mers, with a focus on detecting low-complexity sequences via sub-word repetition analysis.

## Core Functionality

- **`KmerEntropy(kmer, k, levelMax)`**:  
  Computes the *minimum normalized Shannon entropy* across all sub-word sizes from `1` to `levelMax`.  
  - Decodes the encoded *k*-mer (2 bits/base) into a DNA string.  
  - For each word size `ws`, extracts all overlapping substrings, normalizes them to their **circular canonical form**, and counts frequencies.  
  - Normalized entropy = `(log(N) − Σ(nᵢ log nᵢ)/N) / emax`, where `emax` is the theoretical max entropy given sequence length and alphabet constraints.  
  - Returns min entropy across `ws ∈ [1, levelMax]`. Values near **0** indicate repeats (e.g., `AAAAA…`); values near **1** suggest high complexity.

- **`KmerEntropyFilter`**:  
  A reusable, precomputed filter for batch processing millions of *k*-mers efficiently:  
  - Pre-builds normalization tables (for circular canonical forms), entropy lookup values (`emax`, `logNwords`), and frequency tables.  
  - Avoids repeated allocations — critical for performance in pipelines (e.g., read filtering).  
  - **Not goroutine-safe** — each thread must instantiate its own filter.

- **`NewKmerEntropyFilter(k, levelMax, threshold)`**:  
  Initializes a filter with precomputed tables and sets the entropy rejection `threshold`.  

- **`Accept(kmer)` / `Entropy(kmer)`**:  
  - `Accept()` returns `true` if entropy > threshold (i.e., *k*-mer is complex enough to pass).  
  - `Entropy()` computes entropy using precomputed tables — ~10× faster than standalone calls.

## Design Highlights

- **Circular canonical normalization** ensures symmetry (e.g., `AT` ≡ `TA`).  
- **Sub-word-level entropy** captures local repetitiveness better than global *k*-mer uniqueness.  
- Optimized for **speed and memory reuse**, suitable for large-scale genomic data filtering.
