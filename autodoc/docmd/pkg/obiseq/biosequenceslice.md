# `obiseq` Package: BioSequence Collection Management

The `obiseq` package provides a high-performance, memory-efficient implementation for managing collections of biological sequences (`BioSequence`) in Go. Its core type is `BioSequenceSlice`, a slice of pointers to `BioSequence` objects, optimized for batch processing in metagenomic pipelines.

### Key Functionalities

- **Memory Pooling & Allocation Control**:  
  `NewBioSequenceSlice` and `MakeBioSequenceSlice` allow creating slices with optional capacity hints.  
  `EnsureCapacity(capacity)` dynamically grows the underlying slice while logging warnings or panicking on persistent allocation failures.

- **Efficient Element Management**:  
  - `Push(sequence)`: Appends a sequence to the end.  
  - `Pop()`: Removes and returns the last element (nil-safe).  
  - `Pop0()`: Efficiently removes and returns the first element.  

- **Collection Metadata Queries**:  
  - `Len()`: Returns number of sequences in the slice.  
  - `Size()`: Computes total sequence length (summing all `.Len()`).  
  - `NotEmpty()`: Boolean check for non-empty collections.  

- **Attribute Aggregation**:  
  `AttributeKeys(skip_map, skip_definition)` aggregates all attribute keys across sequences into a set—useful for schema inference or validation.

- **Sorting Capabilities**:  
  - `SortOnCount(reverse)`: Sorts by read count (descending/ascending).  
  - `SortOnLength(reverse)`: Sorts by sequence length.

- **Taxonomy Integration**:  
  `ExtractTaxonomy(taxonomy, seqAsTaxa)` builds or extends a taxonomic tree from sequence paths.  
  When `seqAsTaxa=true`, it injects pseudo-taxonomic labels for individual sequences (e.g., `OTU:SEQ0000012345 [seqID]@sequence`), enabling unified taxonomic/rarefaction workflows.

### Design Highlights

- Minimal allocations via manual slice management and `slices.Grow`.  
- Explicit niling of popped elements to aid garbage collection.  
- Integrated logging (via `logrus`) for allocation issues—critical in large-scale NGS data processing.  
- Designed to support `BioSequenceBatch`, a higher-level abstraction for streaming or parallelizable sequence batches.
