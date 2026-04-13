## `obialign` Package: Semantic Overview (≤50 lines)

The `obialign` package provides a lightweight, high-performance utility for **detecting single-edit-distance relationships** between biological sequences (`obiseq.BioSequence`). Its core function, `D1Or0`, determines whether two sequences are either **identical** or differ by exactly **one substitution, insertion, or deletion (indel)**.

- `abs[k]`: A generic helper computing absolute values for integers or floats (via Go generics).
- `D1Or0(...)`: Returns a 4-tuple:
  - **`int` (first)**: `0` if identical, `1` if differing by one edit, `-1` otherwise.
  - **`int` (second)**: Position of the differing site (`-1` if identical).
  - **`byte`, `byte`**: Mismatched characters (or `'-'` for gaps indicating indels).

**Algorithmic strategy:**
1. Early rejection if length difference exceeds 1.
2. Forward scan until first mismatch → identifies left boundary of divergence.
3. Backward scan from ends to find rightmost match boundary.
4. Validates whether the mismatch region allows exactly one edit:
   - Single substitution: equal lengths, single divergent position.
   - Insertion/deletion: length differs by 1 and only one non-overlapping character remains.

Designed for speed in **OTU/ASV dereplication or error correction** pipelines (e.g., metabarcoding), where rapid filtering of near-identical sequences is critical. Does *not* compute full alignments; optimized for binary decision-making under strict edit constraints.
