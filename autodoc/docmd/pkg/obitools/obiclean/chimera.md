## Chimera Detection Module (`obiclean`)

This Go package implements a chimera detection algorithm for amplicon sequencing data, specifically designed to handle IUPAC ambiguity codes. It identifies chimeric sequences—artifacts formed during PCR when incomplete extensions anneal to non-homologous templates in subsequent cycles.

### Core Functions

- **`commonPrefix(a, b)`**: Computes the length of the longest shared prefix between two `BioSequence`s using IUPAC-compliant nucleotide comparison.
- **`commonSuffix(a, b)`**: Computes the length of the longest shared suffix analogously.
- **`oneDifference(s1, s2)`**: Efficiently checks if two sequences differ by exactly one edit operation (substitution, insertion, or deletion), enabling early filtering of near-identical candidates.

### Chimera Annotation Pipeline

The main function `AnnotateChimera(samples)` processes a map of PCR amplicon groups (`map[string][]*seqPCR`):

1. **Filtering**: Retains only *head sequences* (those with no incoming edges), assumed to be consensus or representative variants.
2. **Sorting**: Sequences are ordered by increasing abundance (`Weight`) to prioritize rare sequences as potential chimeras.
3. **Parent Search**: For each candidate chimera `s`, it scans all more abundant sequences (`pcrs[j].Weight > s.Weight`) for parental signatures:
   - Skips pairs differing by only one edit (likely sequencing errors).
   - Tracks the longest common prefix (`nameLeft`, `maxLeft`) and suffix (`nameRight`, `maxRight`).
4. **Chimera Decision Rule**: A sequence is flagged as chimeric if:
   - `maxLeft + maxRight ≥ L` (the sum covers the full length),
   - and it is *not fully contained* within a single parent (`maxRight < L`).
5. **Annotation**: The result is stored in the sequence’s `"chimera"` annotation as a structured string:  
   `{parent_left}/{parent_right}@({overlap})(start)(end)(len)`.

### Design Notes

- Handles IUPAC nucleotide codes via `obiseq.SameIUPACNuc`.
- Uses efficient in-place sequence slicing and string comparison.
- Integrates with `obitools4`’s data model (`BioSequence`, annotations).
