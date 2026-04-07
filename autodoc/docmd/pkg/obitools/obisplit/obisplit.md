# `obisplit` Package: Semantic Description

The `obisplit` package provides functionality to **split biological sequences** based on the detection of user-defined pattern pairs (e.g., primer or barcode sites), commonly used in metabarcoding workflows.

- **Core Types**:  
  - `SplitSequence`: defines a pattern pair (forward/reverse) with an associated name.  
  - `Pattern_match`: stores details of a detected pattern instance (name, coordinates, errors, orientation).

- **Pattern Detection (`LocatePatterns`)**:  
  Scans a sequence for all occurrences of forward and reverse patterns using approximate matching (allowing errors). It:  
  - Converts the input sequence to an indexed format for efficient pattern search.  
  - Extracts matches, normalizes coordinates and reverse-complements backward hits.  
  - Sorts results by start position.  
  - Removes overlapping matches, keeping the one with fewer errors.

- **Sequence Splitting (`SplitPattern`)**:  
  Splits the input sequence into fragments *between* matched patterns. Each fragment is annotated with metadata:  
  - `obisplit_frg`: fragment number (1-based).  
  - `obisplit_nfrg`: total number of fragments.  
  - `obisplit_group`: pair-wise group name (e.g., `"primerA-primerB"` or `"extremity"`, for terminal regions).  
  - `obisplit_set`: the relevant pattern group (e.g., `"primerA"`), or `"NA"`.  
  - `obisplit_location`: genomic coordinates (1-based, inclusive).  
  - Left/right pattern info: name, match string, and error count.

- **Pipeline Integration (`SplitPatternWorker`, `CLISlitPipeline`)**:  
  - Exposes splitting logic as a reusable `SeqWorker`.  
  - Wraps it into an iterable pipeline supporting parallel execution via standard OBITools4 infrastructure.

- **Use Case**:  
  Designed for demultiplexing and amplicon trimming in high-throughput sequencing data, where sequences are partitioned between known molecular markers.
