# `IFragments` Functionality Overview

The `IFragments()` function in the `obiiter` package implements a parallelized sequence fragmentation pipeline for biological sequences. It is designed to split long nucleotide or protein sequences into smaller, overlapping fragments while preserving metadata and enabling concurrent processing.

## Core Parameters
- `minsize`: Minimum sequence length to skip fragmentation.
- `length`: Desired fragment size (in bases/amino acids).
- `overlap`: Number of overlapping residues between consecutive fragments.
- `size`, `nworkers`: Batch size and number of worker goroutines (currently unused in active logic).

## Workflow
1. **Batch Sorting**: Input sequences are batched and sorted for efficient processing.
2. **Parallel Fragmentation**:
   - Each worker processes a subset of batches independently using goroutines.
   - For each sequence longer than `minsize`, it is split into overlapping fragments of length `length` with step size = `length - overlap`.
   - The final fragment is extended to cover the remainder (fusion mode), avoiding tiny trailing pieces.
3. **Resource Management**:
   - Original sequences are recycled (`s.Recycle()`) to optimize memory usage.
   - Fragments are reassembled into batches, sorted by source and order, then rebatched to respect memory/size limits.

## Key Features
- **Overlap handling**: Ensures contiguous coverage without gaps.
- **Memory efficiency**: Uses recycling and batched output.
- **Scalability**: Leverages Go concurrency via `nworkers`.
- **Error safety**: Panics on subsequence errors (e.g., invalid indices).

## Use Case
Ideal for preparing long-read sequencing data (e.g., PacBio, Nanopore) or assembled contigs for downstream analysis requiring fixed-length inputs (e.g., k-mer indexing, ML inference).
