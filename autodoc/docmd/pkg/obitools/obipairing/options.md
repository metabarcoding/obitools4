# `obipairing` Package Functional Overview

The `obipairing` package provides command-line interface (CLI) support and core logic for **paired-end read merging** in NGS data processing. It defines configuration options, input parsing, and alignment parameters used to merge forward and reverse sequencing reads into consensus sequences.

## Key Features

- **Input Handling**: Accepts paired FASTQ/FASTA files via `--forward-reads` (`-F`) and `--reverse-reads` (`-R`) flags.
- **Alignment Parameters**:
  - `_Delta`: Extra overlap buffer (default: `5`) for refining alignment after fast detection.
  - `_MinOverlap`: Minimum overlap length required (default: `20`).
  - `_MinIdentity`: Minimal sequence identity threshold for valid overlaps (default: `90%`).
  - `_GapPenalty`: Multiplier for gap penalties relative to mismatch scores (default: `2.0`).
  - `_PenaltyScale`: Global scaling factor for scoring (default: `1.0`).
- **Alignment Modes**:
  - Fast heuristic alignment enabled by default (`--exact-mode` disables it).
  - Optional absolute scoring in fast mode via `--fast-absolute`.
- **Output Control**:
  - Statistics (e.g., overlap length, identity) can be excluded from consensus headers using `--without-stat`.
- **Integration**:
  - Extends generic input/output options from `obiconvert` for unified pipeline compatibility.
- **Core Functions**:
  - `CLIPairedSequence()`: Reads and pairs forward/reverse sequences.
  - Getter functions (`CLI*`) expose parsed parameters for downstream alignment/merging logic.

This module serves as the configuration and orchestration layer before actual sequence overlap detection, alignment scoring, and consensus generation.
