# `obiseq` Package: Semantic Overview

The `obiseq` package provides a robust, thread-safe implementation of biological sequence objects in Go. It defines the core `BioSequence` type and associated utilities for handling nucleotide sequences (DNA/RNA), quality scores, annotations, features, memory management, and metadata operations.

### Core Functionalities

- **Construction & Initialization**  
  - `NewEmptyBioSequence(cap)` creates an empty sequence with optional preallocated capacity.  
  - `NewBioSequence(id, seq, def)` builds a basic sequence with ID (case-normalized), byte-level sequence (`[]byte`), and definition.  
  - `NewBioSequenceWithQualities(...)` extends the above with per-base quality scores (`[]byte` or `Quality`).  

- **Accessors & Properties**  
  - `Id()`, `Definition()` return metadata fields.  
  - `Sequence()` returns the normalized (lowercase) sequence as a copy of internal bytes.  
  - `Len()` returns the length (number of bases).  
  - `String()` provides a human-readable sequence string.  

- **Quality & Feature Support**  
  - `HasQualities()` checks if quality scores are present.  
  - `Qualities()`, `SetQualities(...)` manage per-base quality data (with fallback to default values).  
  - `Features()` retrieves optional feature annotations as a string.  

- **Annotation System**  
  - `Annotations()`, `HasAnnotation()` allow inspection of arbitrary metadata (key-value map).  
  - Thread-safe via internal `sync.Mutex`, exposed through `AnnotationsLock()`.  

- **Utility & Safety**  
  - `Recycle()` safely resets internal slices and annotations (enables object pooling). Handles nil receivers gracefully.  
  - `Copy()` performs deep copy of all fields, including annotations and locks (new mutex).  
  - `MD5()` computes the MD5 hash of the sequence bytes.  

- **Analysis Methods**  
  - `Composition()` returns a nucleotide count map (`a`, `c`, `g`, `t`, and `'o'` for others), case-insensitive.

All operations are designed with performance, safety (nil-safety, copy semantics), and extensibility in mind—ideal for bioinformatics pipelines requiring immutable or pooled sequence handling.
