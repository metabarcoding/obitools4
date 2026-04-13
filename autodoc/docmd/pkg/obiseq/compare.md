# Semantic Description of `obiseq` Comparison Functions

The `obiseq` package provides utility functions for comparing biological sequence records (`*BioSequence`) based on different fields. These comparators are designed to support sorting, deduplication, or grouping operations in bioinformatics workflows.

- **`CompareSequence(a, b *BioSequence) int`**  
  Compares the raw nucleotide or amino acid sequences (`a.sequence`) lexicographically using `bytes.Compare`. Returns:
  - `<0` if `a < b`,  
  - `0` if equal,  
  - `>0` if `a > b`.

- **`CompareQuality(a, b *BioSequence) int`**  
  Compares the base quality scores (`a.qualities`) lexicographically (as byte strings), following same semantics as above. Useful for sorting reads by quality profiles.

- **Commented-out `CompareAttributeBuilder(key string)`**  
  A planned higher-order function to generate custom comparators based on sequence attributes (e.g., `RG`, `NM`). It would:
  - Extract attribute values using `.GetAttribute(key)`.
  - Handle missing attributes (treat absent as "less than" present).
  - Eventually support typed comparisons for ordered types.

These functions assume `BioSequence` implements a consistent internal structure with `.sequence []byte` and `.qualities []byte`. They enable flexible, field-based ordering in collections of sequencing records.
