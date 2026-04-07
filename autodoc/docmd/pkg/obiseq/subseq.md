# `obiseq.Subsequence` Functionality Overview

The `Subsequence()` method extracts a contiguous segment from a biological sequence (`BioSequence`), supporting both linear and circular topologies.

- **Input validation**: Checks ensure `from < to` (unless circular), positions are non-negative, and bounds respect sequence length.
- **Circular handling**: Positions exceeding the sequence length wrap around using modular arithmetic; debug logs record corrections.
- **Linear extraction**: When `from < to`, it slices the underlying nucleotide/peptide sequence and, if present, its quality scores.
- **Circular extraction**: When `from > to`, it concatenates two linear segments: from `from` → end, and start → `to`.
- **Metadata preservation**: Quality scores (if available) and annotations are copied to the new subsequence.
- **ID formatting**: The resulting sequence ID is suffixed with `[from..to]` (1-based indexing).
- **Mutation tracking**: A private `_subseqMutation()` adjusts stored pairing mismatch positions by subtracting the extraction shift, ensuring coordinate consistency post-extraction.

This enables robust subsequence generation for genomic analysis workflows involving circular genomes (e.g., plasmids) or fragmented reads.
