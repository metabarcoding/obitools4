# BioSequence Pairing Functionality

This package provides semantic tools for managing biological sequence pairings—typically used in genomics (e.g., paired-end reads). Key features:

- **Single-sequence pairing**:  
  - `IsPaired()` checks if a sequence is currently paired.  
  - `PairedWith()` returns the linked partner, or `nil`.  
  - `PairTo(p)` establishes a bidirectional link between two sequences.  
  - `UnPair()` safely severs the pairing on both ends.

- **Batch (slice) handling**:  
  - `IsPaired()` and `UnPair()` operate uniformly across all sequences in a slice.  
  - `PairedWith()` returns the corresponding paired slice (element-wise).  
  - `PairTo(p)` enforces length compatibility and pairs sequences index-by-index.  

- **Error handling**:  
  - Mismatched slice lengths during `PairTo` trigger a fatal log (via Logrus), preventing inconsistent pairings.

Semantically, the API supports both *atomic* and *bulk* pairing operations while preserving consistency through bidirectional references—ideal for processing paired-end sequencing data.
