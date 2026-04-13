# `obiapat`: High-Performance Approximate Pattern Matching for Biological Sequences  

The `obiapat` Go package delivers **fast, memory-safe approximate pattern matching** over biological sequences (DNA/RNA), leveraging a C-based implementation of the **Apat algorithm**. Designed for NGS preprocessing (e.g., primer detection, adapter trimming), it supports fuzzy matching with mismatches/indels, reverse-complement search, circular topology handling, and efficient non-overlapping match filtering—all while integrating seamlessly with the OBITools4 ecosystem.

## Core Concepts

- **`ApatPattern`**: Compiled pattern object (≤64 bp) supporting:
  - IUPAC ambiguity codes (`W`, `R`, `[AT]`)
  - Negated bases (`!A` = "not A")
  - Fixed-position anchors (`#`)
- **`ApatSequence`**: Lightweight wrapper around `obiseq.BioSequence`, enabling optimized pattern scanning with optional circular indexing and memory recycling.

## Public API

### Pattern Construction & Transformation  
- **`MakeApatPattern(pattern string, errormax int, allowsIndel bool) (*ApatPattern, error)`**  
  Compiles a pattern string into an executable automaton. Supports:
  - `errormax`: Max allowed errors (substitutions only if `allowsIndel=false`; indels included otherwise).
  - Pattern syntax: e.g., `"A[T]C!GT#"` → matches "A", then any A/T, then C, allows 1 mismatch at position `!G`, requires exact match at anchored `#T`.
- **`ReverseComplement() *ApatPattern`**  
  Returns a new pattern representing the reverse complement (essential for strand-agnostic DNA searches).
- **`Len() int`**  
  Returns the pattern’s length in bases.

### Matching & Search Operations  

- **`FindAllIndex(seq *ApatSequence, start, end int) [][3]int`**  
  Returns all valid matches in `[start_pos, end_pos, error_count]` format within `seq[start:end)`.  
  - Supports partial sequence scans (e.g., for sliding windows).
- **`IsMatching(seq *ApatSequence, start, end int) bool`**  
  Fast boolean check: does the pattern match *anywhere* in `seq[start:end)` within error tolerance?
- **`BestMatch(seq *ApatSequence, start, end int) (start, end, errors int)`**  
  Finds the *lowest-error* match in a region. For indel patterns, performs local realignment to refine alignment boundaries.
- **`FilterBestMatch(seq *ApatSequence, start, end int) [][3]int`**  
  Returns **non-overlapping matches**, prioritizing lower-error occurrences (greedy selection from best to worst).
- **`AllMatches(seq *ApatSequence, start, end int) [][3]int`**  
  Computes all valid matches (including indel-aware realignment), then filters to non-overlapping set using `FilterBestMatch`.

### Resource Management  
- **`Free()`**  
  Explicitly releases C-level resources. Finalizers auto-cleanup, but manual `Free()` is recommended in hot loops for predictable memory use.

## PCR Simulation Module (`PCRSim` family)

Implements *in silico* PCR with configurable primer tolerance and amplicon constraints:

- **`PCRSim(seq obiseq.BioSequence, opts ...Option) []Amplicon`**  
  Simulates PCR on a single sequence. Options include:
  - `OptionForwardPrimer(pattern string, errormax int)` / `OptionReversePrimer(...)`
  - `OptionMinLength(n)`, `OptionMaxLength(n)` → filter amplicons by size
  - `OptionWithExtension(len int, strict bool)` → add flanking regions (trim if `strict=false`)
  - `OptionCircular(bool)` → handle circular DNA topology
- **`PCRSlice(seqs []obiseq.BioSequence, opts ...Option) [][]Amplicon`**  
  Batch PCR across multiple sequences.
- **`PCRSliceWorker(opts ...Option) func(int, obiseq.BioSequence) (int, interface{})`**  
  Returns a reusable worker for parallel execution via `obiseq.MakeISliceWorker`.

### Output Format  
Each amplicon includes:
- Coordinates, primer positions/errors/directions
- Flanking extensions (if requested)
- Original sequence metadata preserved

## Predicate Generator: `IsPatternMatchSequence`

Returns a **reusable function** for sequence filtering:
```go
func IsPatternMatchSequence(
  pattern string, errormax int,
  bothStrand bool, allowIndel bool
) obiseq.SequencePredicate
```
- Internally builds `ApatPattern` + reverse complement (if needed).
- Predicate logic:  
  ```go
  func(seq *obiseq.BioSequence) bool {
    return pattern.IsMatching(...) || (!bothStrand && false)
              || rcPattern.IsMatching(...)
  }
  ```
- Ideal for high-throughput read filtering (e.g., barcode detection, primer contamination checks).

## Implementation Highlights

- **C interoperability** via `cgo` with custom memory management (no Go heap copies).
- **Finalizers + manual `Free()`** prevent leaks in long-running pipelines.
- Uses `unsafe.SliceData` for zero-copy sequence access during matching.
- Logging via **Logrus** (errors at `ErrorLevel`, debug amplicon details at `DebugLevel`).
