# `obialign` Package: Sequence Alignment Utilities

The `obialign` package provides core functions for pairwise biological sequence alignment in Go, designed to work with `obiseq.BioSequence` objects.

- **Core Alignment Construction**: `_BuildAlignment()` and `BuildAlignment()` reconstruct aligned sequences from a precomputed alignment path (e.g., output by dynamic programming). It supports gap characters and reuses buffers for efficiency.

- **Quality-Aware Consensus Building**: `BuildQualityConsensus()` generates a consensus sequence from an alignment and per-base quality scores:
  - At mismatches, it retains the higher-quality base.
  - When qualities are equal and bases differ, an IUPAC ambiguity code is used (via `_FourBitsBaseCode`/`_Decode`).
  - Quality values are combined and adjusted for mismatches using a Phred-like error probability model.
  - Optionally records mismatch statistics in sequence attributes.

- **Performance & Memory Efficiency**: Uses preallocated buffers (via `PEAlignArena`) or fallback allocation, with slice recycling to minimize GC pressure.

- **Metadata Handling**: Preserves sequence IDs and definitions in output; supports optional mismatch reporting for downstream analysis.

- **Alignment Path Format**: The path is a sequence of signed integers encoding:
  - Negative steps → deletions in seqB (insertion in A),
  - Positive steps → insertions in B,
  - Consecutive pairs encode match/mismatch runs.

This package is part of the OBITools4 ecosystem, targeting high-throughput amplicon or metagenomic data processing.
