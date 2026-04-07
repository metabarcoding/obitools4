# `obiconsensus` Package: Semantic Overview

The `obiconsensus` package delivers scalable, graph-based consensus and denoising tools for high-throughput biological sequence data within the OBITools4 ecosystem. It enables error correction, variant clustering, and consensus reconstruction from related amplicon or metagenomic reads—supporting both single-sample and multi-sample workflows.

## Public API Summary

### Core Algorithms & Utilities
- **`BuildConsensus()`**:  
  Constructs a consensus sequence via *de Bruijn graph* assembly of input reads. Automatically selects optimal `k`-mer size (fallback: longest common suffix analysis). Detects graph cycles and incrementally increases `k` until resolved. Optionally persists intermediate graphs (`*.gml`) and FASTA inputs. Output includes metadata: consensus flag, total read weight (summed abundances), `k`-mer size used, and graph statistics.

- **`SampleWeight()`**:  
  Returns a closure that retrieves per-sequence sample abundances (e.g., read counts) from sequence annotations or statistics—enabling weighted graph operations.

- **`SeqBySamples()`**:  
  Groups sequences by sample identifier, using a configurable annotation key (default: `"sample"`). Supports grouping based on either statistical attributes (`StatsOn`) or sequence metadata.

- **`BuildDiffSeqGraph()`**:  
  Builds a *difference graph* where nodes represent unique sequences and edges encode single-nucleotide mutations (position + substitution). Uses `obialign.D1Or0` for exact alignment or approximate LCS-based distance scaling. Supports parallel edge computation and optional progress bar.

- **`MinionDenoise()`**:  
  Denoises sequences by identifying high-degree nodes (potential consensus hubs), building local consensuses via `BuildConsensus()`, and preserving low-degree nodes unchanged. Propagates sample annotations, weights, and metadata.

- **`MinionClusterDenoise()`**:  
  Denoises via *weight-based clustering*: aggregates node weights (self + neighbors), selects local maxima as cluster heads, and builds consensus per neighborhood.

- **`CLIOBIMinion()`**:  
  CLI orchestrator for end-to-end denoising: loads sequences, groups by sample (`--sample`), builds per-sample difference graphs (optional export via `--save-graph`), applies denoising (`MinionDenoise()` or `MinionClusterDenoise()`), optionally deduplicates output (`--unique`), and annotates sequence lengths.

### Configuration & CLI Helpers
- **Clustering Mode**: `--cluster` (`-C`) enables graph-based clustering.
- **Distance Threshold**: `--distance` (`-d`, default: 1) sets max Hamming distance for edge inclusion.
- **K-mer Control**: `--kmer-size` (`SIZE`, default: -1 = auto-selected).
- **Sample Key**: `--sample` (`-s`, default: `"sample"`) defines the annotation field for sample grouping.
- **Filtering Options**:  
  - `--no-singleton`: excludes unique sequences.  
  - `--low-coverage` (default: 0) filters low-abundance sequences.
- **Output Options**:  
  - `--unique` (`-U`) enables deduplication (via `obiuniq`).  
  - `--save-graph DIR` exports graphs in GraphML.  
  - `--save-ratio FILE` writes edge abundance ratios as CSV.
- **Format Integration**: Works with `obiconvert` via unified input/output option sets (`InputOptionSet`, `OutputOptionSet`) for FASTA/FASTQ handling.
- **Getter Functions**: Typed accessors (e.g., `CLIDistStepMax()`, `CLIKmerSize()`) decouple argument parsing from core logic.

## Design Principles
- **Parallelism**: Leverages goroutines and `sync.WaitGroup` for scalable graph construction.
- **Robustness**: Handles edge cases (e.g., single-sequence inputs) gracefully with logging.
- **Extensibility**: Modular architecture allows swapping alignment engines or graph representations.

*Purpose: Accurate, reproducible consensus and denoising for NGS amplicon/metagenomic data at scale.*
