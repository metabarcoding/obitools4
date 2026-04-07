## Semantic Description of `obiseq` Package Functionality

The `obiseq` package provides core bioinformatics utilities for nucleic acid sequence manipulation in Go. It centers around two key operations:

- **Nucleotide Complementation (`nucComplement`)**  
  Implements standard Watson-Crick base pairing rules: `A↔T`, `C↔G`. It also handles ambiguous or symbolic characters (e.g., `'n' → 'n'`, `'[ ↔ ]'`), preserving non-standard symbols like gaps (`'-'`) and missing data (`'.'`). This function serves as the atomic building block for reverse-complement logic.

- **Reverse Complementation (`BioSequence.ReverseComplement`)**  
  A method on the `BioSequence` type that returns a new (or in-place modified) sequence representing:
  - The *reverse* of the original nucleotide string, followed by  
  - Each base replaced with its complement (via `nucComplement`).  

  The method supports two modes:
  - **Non-destructive (`inplace=false`)**: Returns a new `BioSequence`, leaving the original unchanged.
  - **In-place (`inplace=true`)**: Modifies and returns the same object for memory efficiency.

  Crucially, it preserves associated quality scores (e.g., Phred-scaled sequencing qualities), reversing their order to match the reversed sequence—ensuring correctness in downstream analyses like alignment or variant calling.

Tests validate both functions across edge cases: degenerate bases, ambiguous symbols, and quality-aware sequences—confirming robustness for typical NGS (Next-Generation Sequencing) workflows.
