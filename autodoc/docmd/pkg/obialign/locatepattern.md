# `LocatePattern` Functionality Overview

The `obialign.LocatePattern` function implements a **local alignment algorithm** to find the best approximate match of a short DNA pattern (e.g., primer) within a longer biological sequence, using **dynamic programming**.

- **Input**:  
  - `id`: identifier for logging/error reporting.  
  - `pattern []byte`: the query sequence (e.g., primer).  
  - `sequence []byte`: the target read/contig.  

- **Constraints**:  
  - Pattern must be strictly shorter than the sequence (`len(pattern) < len(sequence)`).  

- **Scoring Scheme**:  
  - Match: `+0` (using IUPAC compatibility via `obiseq.SameIUPACNuc`).  
  - Mismatch/Gap: `-1`.  

- **Algorithm Features**:  
  - End-gap free alignment (no penalty for gaps at sequence ends), enabling flexible primer positioning.  
  - Uses a flattened buffer (`buffIndex`) for memory-efficient matrix storage (width × height).  
  - Tracks alignment path via `path` array: diagonal (`0`, match/mismatch), up (`+1`, deletion in pattern/left gap), left (`-1`, insertion/deletion).  
  - Backtracks from the bottom-right to find optimal local alignment start/end coordinates.  

- **Output**:  
  - `start`: starting index in `sequence`.  
  - `end+1`: ending index (exclusive) of best match.  
  - Error count: `-score`, i.e., number of mismatches/gaps in alignment.  

- **Use Case**:  
  Designed for high-throughput amplicon processing (e.g., primer trimming in metabarcoding pipelines like OBITools4).
