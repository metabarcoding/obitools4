# `obipairing` Package — Functional Overview

The `obipairing` package enables robust merging of paired-end next-generation sequencing (NGS) reads within the OBITools4 ecosystem. It bridges input parsing, alignment configuration, and consensus assembly—supporting both high-accuracy overlap-based merging and lightweight fallback concatenation when overlaps are unreliable.

## Public API Summary

### CLI Interface (`obipairing/cli.go`)
- **Input specification**:  
  `--forward-reads` (`-F`) and `--reverse-reads` (`-R`) flags accept FASTQ/FASTA file paths.
- **Alignment tuning**:  
  - `_Delta` (`--delta`, default `5`) — buffer for refining initial overlap detection.  
  - `_MinOverlap` (`--min-overlap`, default `20`) — minimum overlap length.  
  - `_MinIdentity` (`--min-identity`, default `90`) — minimum % identity for valid alignment.  
  - `_GapPenalty` (`--gap-penalty`, default `2`) — gap cost multiplier vs mismatches.  
  - `_PenaltyScale` (`--scale`, default `1`) — global scoring scaling factor.
- **Alignment mode control**:  
  - Fast heuristic enabled by default; `--exact-mode` disables it.  
  - Absolute scoring in fast mode via `--fast-absolute`.
- **Output customization**:  
  `--without-stat` omits alignment statistics from consensus headers.
- Extends generic I/O options inherited from `obiconvert` for pipeline compatibility.

### Core Assembly Functions (`obipairing/assemble.go`)
- **`JoinPairedSequence(seqA, seqB *obiseq.BioSequence, inplace bool) (consensus *obiseq.BioSequence)`**  
  Concatenates forward and reverse reads with a `..........` (10-dot) separator.  
  - Quality scores for dots set to Phred `Q=0` if both inputs are quality-tracked.  
  - Supports in-place recycling (`inplace=true`) to reduce allocations.

- **`AssemblePESequences(...)`**  
  Performs high-fidelity paired-end assembly:  
  - Uses `obialign.PEAlign` with a two-stage process:
    1. **Fast heuristic** (`FAST`) to locate candidate overlap region.
    2. **Dynamic programming refinement**, extended by `_Delta`.
  - Validates alignment against thresholds (`minOverlap`, `minIdentity`).  
    Falls back to join if criteria unmet.  
  - Optionally annotates output with alignment metadata:
    ```go
    "mode"          → "alignment" or "join"
    "ali_length"    → overlap length
    "score_norm"    → normalized alignment score
    "identity"      → % identity over overlap
    "directionality"→ orientation (e.g., FR)
    ```
  - Supports in-place reuse (`inplace`) and absolute/relative scoring via `fastModeRel`.

- **`IAssemblePESequencesBatch(...)`**  
  Parallelizes assembly over batches of read pairs:  
  - Consumes iterators from `PairWith` (e.g., via `obiiter`).  
  - Launches configurable workers (`nworkers`) and channel buffer size.  
  - Internally reverse-complements the second read before alignment (`seqB.ReverseComplement()`).  
  - Yields assembled consensus sequences via an iterator.

### Configuration & Parameter Access
- Getter functions (`CLI*`) expose parsed CLI parameters (e.g., `CLIMinOverlap()`, `CLIGapPenalty()`), enabling downstream alignment modules to reuse CLI-defined settings.

### Annotation Semantics
Each assembled sequence carries annotations describing the assembly mode and, when applicable:
- Alignment scores (`ali_score`, `score_norm`)
- Overlap metrics (`ali_length`, `identity`)
- Fast-mode metadata (e.g., `"pairing_fast_score"`) when heuristic alignment is used.

Designed for scalability, low memory footprint, and integration with `obiseq`, `obiiter`, and alignment backends in OBITools4.
