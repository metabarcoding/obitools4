## BioSequence.Kmers(k int) — Semantic Description

The `Kmers` method is a generator function that yields all contiguous *k*-length subsequences (called **k-mers**) from a biological sequence (`BioSequence`).  

- It operates on `[]byte` data, assuming the underlying sequence is stored as a byte slice (e.g., DNA bases `A`, `C`, `G`, `T`).  
- Uses Go’s new iterator protocol (`iter.Seq[[]byte]`) for memory-efficient, lazy evaluation.  
- Validates input: returns an empty iterator if `k ≤ 0` or exceeds sequence length.  
- Iterates linearly from index `i = 0` to `len(seq) - k`, extracting slices of length *k*.  
- Each yielded value is a **non-copying slice view** (efficient, but mutable if original data changes).  
- Supports early termination: the consumer can stop iteration by returning `false` from the yield callback.  
- Designed for downstream tasks like sequence analysis, motif discovery, or hashing (e.g., in k-mer counting).  
- Does *not* handle reverse-complement or ambiguous bases—assumes raw sequence input.  

Usage example:  
```go
for kmer := range seq.Kmers(3) {
    fmt.Printf("%s\n", string(kmer))
}
```  
This yields all 3-mers (e.g., `"ACG"`, `"CGT"`...) in order.
