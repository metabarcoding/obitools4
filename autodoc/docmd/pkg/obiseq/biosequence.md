# BioSequence: A High-Performance Biological Sequence Representation

The `obiseq` package defines the `BioSequence` struct, a memory-efficient and thread-safe container for biological DNA sequences. Beyond raw sequence data (`[]byte`), it supports rich metadata and operations essential for NGS pipelines.

## Core Features

- **Metadata Fields**:  
  - `id`: Unique sequence identifier.  
  - `source`: Filename (without path/extension) of origin.  
  - `definition`: Optional descriptive text, stored in annotations.

- **Sequence & Quality Support**:  
  - Stores sequence as lowercase `[]byte` (normalized via in-place lowercasing).  
  - Quality scores (`Quality = []uint8`) with fallback to default Phred+40 values when missing.  
  - Methods for incremental writing (`Write`, `WriteByte`) and clearing.

- **Annotations & Features**:  
  - Generic `Annotation` map (`map[string]interface{}`) for flexible metadata.  
  - Thread-safe access via `annot_lock` mutex (explicit locking/unlocking methods).  
  - Raw feature table storage (`[]byte`, e.g., EMBL/GenBank features).

- **Biological Relationships**:  
  - `paired`: Pointer to mate/read-pair sequence.  
  - `revcomp`: Pointer to reverse-complement variant (lazy or precomputed).

- **Introspection & Utility**:  
  - `Len()`, `HasSequence()`, `Composition()` (nucleotide counts: a,c,g,t,o).  
  - MD5 checksums (`MD5()` and `MD5String()`) for deduplication.  
  - Memory footprint estimation (`MemorySize()`), critical for streaming/batching.

- **Efficiency Optimizations**:  
  - `NewBioSequenceOwning`/`TakeQualities`: Zero-copy slice adoption (caller must not reuse input).  
  - `Recycle()`: Reuses slices via pool-aware functions (`RecycleSlice`, etc.).  
  - Global counters track creation/destruction/in-memory sequences for diagnostics.

- **Safety & Compatibility**:  
  - Copy semantics via `Copy()` (deep copy of slices + annotations).  
  - Validation: `HasValidSequence` enforces allowed characters (`a-z`, `-`, `.`, `[`, `]`).  
  - Uses unsafe string conversion for quality ASCII output (Phred shift configurable via `obidefault`).

Designed for scalability in large-scale metabarcoding workflows (e.g., OBITools4), balancing performance, correctness, and extensibility.
