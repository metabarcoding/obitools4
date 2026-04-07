# Semantic Description of `obikmersim` Package

The `obikmersim` package provides tools for **k-mer-based sequence matching and alignment**, designed for high-throughput biological sequence analysis (e.g., amplicon or metagenomic reads). It leverages efficient k-mer indexing and alignment strategies to compare query sequences against a reference set.

## Core Functionalities

1. **K-mer Counting & Matching (`MakeCountMatchWorker`)**  
   - Builds a `KmerMap` from reference sequences.  
   - For each query sequence, retrieves matching references via shared k-mers (filtered by minimum count).  
   - Annotates the query with metadata: match count, k-mer size, and sparsity mode.

2. **K-mer-Guided Alignment (`MakeKmerAlignWorker`)**  
   - Uses k-mers to seed candidate alignments between query and reference sequences.  
   - Performs local alignment with quality-aware consensus building (`ReadAlign`, `BuildQualityConsensus`).  
   - Computes identity, residual similarity (k-mer-aware), alignment length, and orientation.  
   - Filters outputs based on identity threshold (default ≥80%) and alignment length.

3. **CLI Wrappers (`CLILookForSharedKmers`, `CLIAlignSequences`)**  
   - Integrate workers into processing pipelines.  
   - Support self-comparison (`CLISelf()`), batched iteration, and parallel execution.  
   - Configure k-mer size (`CLIKmerSize()`), sparsity, max occurrences, gap/scale parameters.

## Key Features

- **Sparse k-mers**: Optional masking of specific positions (e.g., for degenerate bases).  
- **Fast scoring heuristic**: Preliminary alignment score estimation before full path resolution.  
- **Orientation handling**: Automatically detects reverse-complement matches.  
- **Rich annotation output**: Attributes include alignment statistics, orientation, and quality metrics.

## Use Cases

- Read clustering  
- Reference-based read assignment  
- Error correction via consensus building  
- Similarity screening in large sequence datasets
