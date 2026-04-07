# `obiseq` Package: Subsequence Extraction Functionality

The `Subsequence()` method enables extraction of a contiguous segment from biological sequence data (`BioSequence`). It supports both linear and circular (wrapped) slicing.

- **Input Parameters**:
  - `from`, `to`: 0-based inclusive indices defining the slice range.
  - `circular`: boolean flag enabling wrap-around when `from > to`.

- **Behavior**:
  - For linear (`circular = false`), `from ≤ to`, and indices within bounds `[0, len(seq))`.
  - For circular (`circular = true`), allows wrap-around (e.g., `from=3, to=2` on a 4-mer yields indices `[3,0,1]`).
  - Validates inputs: returns descriptive errors for:
    - `from > to` (non-circular),
    - out-of-bounds indices (`< 0` or `≥ length`),
    - invalid ranges.

- **Quality Support**:
  - When sequence includes base quality scores (`BioSequenceWithQualities`), the method preserves corresponding sub-slice of `Quality[]`.

- **Return Value**:
  - Returns a new `BioSequence` (or subclass) instance containing the extracted subsequence and its optional qualities.

- **Use Case**:
  - Ideal for region-of-interest extraction (e.g., primer binding sites, domain segments), especially in circular genomes or plasmids.

- **Testing**:
  - Unit tests (`TestSubsequence`) cover valid/invalid inputs, circular/non-circular modes, and quality consistency.

This functionality provides robust, semantics-aware slicing for biosequence manipulation in Go.
