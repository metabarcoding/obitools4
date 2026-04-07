# Obiclean: PCR Amplicon Error Correction & Chimera Detection

Obiclelan is a Go package for cleaning high-throughput amplicon sequencing data. It corrects PCR/sequencing errors by leveraging abundance-weighted sequence relationships and optionally detects chimeric artifacts using graph-based heuristics. Built for scalability, it integrates with OBITools4’s data model and supports IUPAC ambiguity codes.

## Core Concepts

- **`seqPCR`**: Represents a sequence in one sample, with fields for raw count (`Count`) and post-clustering weight (`Weight`), plus graph edges, annotations, and cluster membership.
- **Directed similarity graphs**: Edges point from more abundant (father) to less abundant (son) sequences differing by ≤ *d* nucleotides.
- **Abundance-weighted correction**: Less abundant sequences are penalized unless supported by strong graph evidence.

---

## Public Functionalities

### 1. **Graph Construction**

- `BuildSeqGraph(samples, distance)`: Builds a mutation graph across samples.  
  - Compares all sequence pairs within/between samples.
  - Adds directed edges only if father has higher weight and differs by ≤ `distance` mismatches.
  - Uses parallel workers (`buildSamplePairs`) for one-error edges and `FastLCSScore` for multi-error extensions.

- `FilterGraphOnRatio(samples, ratio)`: Removes spurious edges violating a power-law decay model:  
  `weight_ratio < (ratio)^distance`. Ensures only statistically plausible edges remain.

---

### 2. **Annotation & Status Assignment**

- `annotateOBIClean(samples)`: Populates per-sequence annotations:
  - `"obiclean_head"`: `true` if the sequence has no incoming edges (i.e., is a cluster head).
  - `"obiclean_singletoncount"`, `"internalcount"`, `"headcount"`: Global counts of sequences in each status across all samples.

- `ObicleanStatus(seq) string`: Returns one of:
  - `"s"`: Singleton (no edges).
  - `"h"`: Hub (has outgoing → sons, but no incoming father) — likely erroneous ancestor.
  - `"i"`: Internal (has both parents and children) — intermediate error variant.

- `Status(seq, sample)` / `Weight(seq, sample)`: Get/set per-sample status (`h/i/s`) and weight annotations.

---

### 3. **Clustering & Head Selection**

- `GetCluster(seq, sample)`: Retrieves or initializes cluster membership (e.g., `"cluster_42"`).
- `GetMutation(seq) map[string]int`: Returns mutation counts (e.g., `"A->T@42": 3`).
- `Mutation(samples)`: Populates mutation annotations from graph edges.

---

### 4. **Chimera Detection**

- `AnnotateChimera(samples)`: Flags chimeric sequences per sample:
  - Filters candidates to *head* sequences only.
  - For each candidate `s`, scans more abundant parents for prefix/suffix matches:
    - Uses IUPAC-aware comparisons (`commonPrefix`, `commonSuffix`).
    - Skips near-identical pairs (one edit difference via `oneDifference`).
  - Flags as chimera if:
    ```
    maxPrefixLen + maxSuffixLen ≥ L
      AND not fully contained in one parent (maxSuffix < L)
    ```
  - Annotation format:  
    `"parent_left/parent_right@(overlap)(start)(end)(len)"`.

---

### 5. **Filtering & Output Control**

- CLI-style filters (applied post-processing):
  - `OnlyHead`: Keep only `"obiclean_head"` sequences.
  - `NotAlwaysChimera`: Exclude sequences flagged chimera in *all* samples.
  - `MinSampleCount(n)`: Retain sequences present ≥ *n* times across samples.

- Optional exports:
  - `SaveGMLGraphs(samples)`: Writes per-sample graphs in GML (node shapes/colors encode abundance/status).
  - `EmpiricalDistCsv(samples)`: Exports substitution statistics (e.g., A→C rates at position *i*) to compressed CSV.
  - `EstimateRatio(samples, minStatCount)`: Collects distance-1 substitution events for downstream modeling.

---

## Design Highlights

- **IUPAC-compliant comparisons**: Nucleotide equality via `obiseq.SameIUPACNuc`.
- **Annotation-driven**: No in-place mutation; all metadata stored via `BioSequence.Annotations`.
- **Scalable parallelism**: Uses goroutines + channels for pairwise comparisons; integrates `progressbar`/Logrus.
- **Flexible thresholds**: Configurable via flags (`distance`, `ratio`, `min-sample-count`), defaulting to sensitivity-optimized values.
