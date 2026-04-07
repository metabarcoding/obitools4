# BioSequence Reverse Complement Functionality

This Go package (`obiseq`) provides utilities for computing the reverse complement of biological sequences (e.g., DNA), including support for quality scores and structured metadata.

## Core Functions

- **`nucComplement(n byte) byte`**  
  Returns the nucleotide complement using a lookup table (`_revcmpDNA`). Handles special cases:  
  - `.` / `-` → unchanged (gaps)  
  - `[`, `]` → swapped (`[` ↔ `]`)  
  - A–Z letters → complemented (case-insensitive via bitwise masking)  
  - Unknown characters → `'n'`

- **`BioSequence.ReverseComplement(inplace bool) *BioSequence`**  
  Performs reverse complement on the sequence and (if present) its quality string:  
  - If `inplace = false`, a copy is made; original preserved.  
  - Reverses indices and complements each base using `nucComplement`.  
  - Also reverses the quality array symmetrically.  
  - Caches result in `sequence.revcomp` for reuse.

- **`BioSequence._revcmpMutation() *BioSequence`**  
  Adjusts mutation metadata (e.g., `"pairing_mismatches"`) to reflect the reversed-complement orientation:  
  - Reverses and complements symbolic mutation strings (e.g., `"A>T"` → `"T>A"`).  
  - Updates positional indices to match reversed sequence coordinates.

- **`ReverseComplementWorker(inplace bool) SeqWorker`**  
  Returns a reusable `SeqWorker` function for batch processing: applies reverse complement to each sequence in a stream.

## Design Notes

- Uses ASCII bitwise tricks (`&31`, `|0x20`) for case-insensitive indexing and lowercase output.  
- Supports non-standard symbols (e.g., IUPAC ambiguity codes via lookup table).  
- Integrates quality scores and structured attributes seamlessly.  

> Ideal for NGS preprocessing pipelines where orientation matters (e.g., paired-end alignment, variant calling).
