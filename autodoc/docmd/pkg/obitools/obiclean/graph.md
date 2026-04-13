# Obiclean: Graph-Based Error Correction for PCR Amplified Sequences

Obiclean is a Go package implementing an error-correction pipeline for amplicon sequencing data (e.g., metabarcoding), built around a directed similarity graph. It identifies and filters sequencing errors by leveraging abundance-weighted relationships between sequences.

## Core Data Structures

- `Ratio`: Stores statistical metrics for nucleotide substitutions (e.g., original/mutant counts, positions), used later to estimate empirical transition probabilities.
- `Edge`: Represents a directed link between two sequences (father → son), encoding Hamming distance (`Dist`), position, and nucleotide change.

## Graph Construction

- **One-error graph**: `buildSamplePairs()` compares all sequence pairs within a sample, adding edges only when the father has higher abundance and differs by exactly one mismatch (`obialign.D1Or0`). Parallelized via worker goroutines.
- **Multi-error extension**: `extendSimilarityGraph()` fills in longer-distance edges using a fast LCS-based alignment (`FastLCSScore`) with tolerance up to `maxError`.
- **Sorting**: Sequences are pre-sorted by ascending count (`sortSamples`) to ensure correct parent–child ordering.

## Graph Refinement & Reweighting

- **Reweighting**: `reweightSequences()` redistributes counts upward along edges, mimicking a probabilistic correction model where sons "donate" weight to fathers proportionally to their abundance.
- **Edge filtering**: `FilterGraphOnRatio()` removes edges where the weight ratio violates a power-law decay model (`weight_ratio < ratio^distance`), suppressing spurious long-distance links.

## Output Generation

- **CSV export**: `EmpiricalDistCsv()` writes substitution statistics (e.g., A→C transitions) to a compressed CSV file, grouped by nucleotide pair code (`nucPair`/`intToNucPair`).
- **GML visualization**: `SaveGMLGraphs()` generates per-sample graph files in GML format, with node shapes (circle/rectangle) and colors encoding abundance thresholds (`statThreshold`).

## Status Classification

- `ObicleanStatus()` labels each sequence as:
  - `"s"` (singleton): no incoming or outgoing edges.
  - `"h"` (hub): has sons but no outgoing edges → likely erroneous ancestor of correct variants.
  - `"i"` (internal): has both incoming and outgoing edges → likely intermediate error.

## Statistical Estimation

- `EstimateRatio()` collects substitution events with distance = 1 and sufficient father weight (`minStatRatio`) into `[][]Ratio`, enabling downstream modeling of transition biases.

## Parallelism & UX

- Uses goroutines + channels for scalable pairwise comparisons.
- Integrates `progressbar` and Logrus logging to provide real-time progress feedback during heavy computations.

