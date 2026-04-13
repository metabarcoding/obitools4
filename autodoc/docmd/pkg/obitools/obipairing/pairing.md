# `obipairing` Package — Semantic Overview

The `obipairing` package provides tools for assembling paired-end sequencing reads in the OBITools4 framework. It supports two main strategies: **overlap-based assembly** (when reads overlap sufficiently) and simple **concatenation with a separator**, when no reliable alignment is possible.

### Core Functions

- `JoinPairedSequence(seqA, seqB *obiseq.BioSequence, inplace bool)`:  
  Merges two sequences with a fixed `..........` (10-dot) separator. If both inputs have quality scores, the dots are assigned a Phred score of 0.

- `AssemblePESequences(...)`:  
  Performs high-fidelity assembly using the `obialign.PEAlign` algorithm:
  - Detects optimal overlap via a fast heuristic (`FAST`) followed by dynamic programming refinement.
  - Validates alignment against `minOverlap`, `minIdentity` thresholds; falls back to join if criteria fail.
  - Optionally annotates results with alignment statistics (score, length, identity, directionality).
  - Supports in-place recycling of input sequences to reduce memory usage.

- `IAssemblePESequencesBatch(...)`:  
  Parallelizes assembly over batches of paired reads using an iterator interface:
  - Consumes `PairWith`-generated iterators.
  - Launches configurable number of workers (`nworkers`) and channel buffer size (via `sizes`).
  - Internally reverses the second read (`seqB.ReverseComplement`) before alignment.
  - Returns an iterator of assembled consensus sequences.

### Key Parameters

- `gap`, `scale`: Gap penalty and scaling factor for alignment scoring.
- `delta`: Extension margin around the initial FAST overlap region.
- `minOverlap`, `minIdentity`: Thresholds to accept an alignment over simple joining.
- `fastAlign` / `fastModeRel`: Controls use of fast heuristic and scoring mode (absolute/relative).
- `withStats`, `inplace`: Toggle statistics output and in-place sequence reuse.

### Output Semantics

Each assembled read is annotated (via `Annotations()`) with:
- `"mode"`: either `"alignment"` or `"join"`.
- Alignment stats (`ali_length`, `score_norm`, etc.) when applicable.
- FAST-specific metadata if used (e.g., `"pairing_fast_score"`).

Designed for scalability and low memory footprint, the package integrates tightly with `obiseq`, `obiiter`, and alignment modules in OBITools4.
