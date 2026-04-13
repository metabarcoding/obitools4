# `obiconsensus` Package: Semantic Overview

The `obiconsensus` package implements high-performance consensus and denoising algorithms for biological sequence data within the OBITools4 framework. It supports error correction, variant clustering, and consensus building from related sequencing reads.

## Core Functionality

- **`BuildConsensus()`**: Constructs a consensus sequence from an input set of related sequences using *de Bruijn graph* assembly.  
  - Automatically estimates optimal `k`-mer size if not provided (via longest common suffix analysis).  
  - Detects and resolves cycles in the graph by incrementally increasing `k`.  
  - Optionally saves intermediate graphs (`*.gml`) and input sequences (`*.fasta`).  
  - Annotates output with metadata: consensus flag, total weight (sum of read counts), *k*-mer size used, and graph statistics.

- **`SampleWeight()`**: Returns a function to retrieve per-sequence sample abundances (e.g., read counts) from sequence statistics or attributes.

- **`SeqBySamples()`**: Groups sequences by sample identifiers (retrieved from a specified annotation key), supporting both statistical (`StatsOn`) and attribute-based grouping.

- **`BuildDiffSeqGraph()`**: Builds a *difference graph* between sequences in a sample:  
  - Nodes = unique sequences; edges = pairwise mutations (position + substitution).  
  - Uses fast alignment (`obialign.D1Or0`) or approximate LCS-based distance for scalability.  
  - Supports parallel edge computation and progress bar visualization.

- **`MinionDenoise()`**: Denoises sequences by:  
  - Identifying high-degree nodes (potential consensus candidates).  
  - Building local consensuses for hubs using `BuildConsensus()`.  
  - Preserving low-degree nodes as-is.  
  - Propagating sample annotations and abundance weights.

- **`MinionClusterDenoise()`**: Alternative denoising via *weight-based clustering*:  
  - Computes aggregate weights per node (self + neighbors).  
  - Selects local weight maxima as cluster heads.  
  - Builds consensus for each head’s neighborhood.

- **`CLIOBIMinion()`**: CLI entry point orchestrating full denoising pipeline:  
  - Loads sequences, groups by sample.  
  - Builds and optionally saves difference graphs per sample.  
  - Applies `MinionDenoise()` or `MinionClusterDenoise()`.  
  - Optionally applies deduplication (`obiuniq`) and adds sequence length annotations.

## Design Highlights

- **Parallelism & Progress**: Uses goroutines, `sync.WaitGroup`, and optional progress bars.  
- **Robustness**: Graceful fallbacks (e.g., single-sequence handling, error logging).  
- **Extensibility**: Modular design with pluggable graph and alignment components.  

*Package purpose: Accurate, scalable consensus building for amplicon or metagenomic sequencing data.*
