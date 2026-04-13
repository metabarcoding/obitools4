# `obiclean` Package Overview

The `obiclean` package implements a sequence clustering and error-correction module for high-throughput sequencing data, primarily aimed at removing PCR/sequencing errors and detecting chimeric reads.

## Core Functionality

- **Error Correction via Abundance Thresholding**: Sequences below a minimum abundance ratio (`--ratio`) relative to more abundant variants are treated as errors.
- **Distance-Based Clustering**: Sequences differing by ≤ `--distance` nucleotides may be grouped as error variants of a consensus (head) sequence.
- **Sample Filtering**: `--min-sample-count` enforces that a sequence appears in at least *N* samples before inclusion.
- **Head Selection**: The `--head` flag restricts output to sequences marked as "heads" (i.e., representative consensus) in ≥1 sample.
- **Chimera Detection**: Optional `--detect-chimera` flag enables chimera identification using abundance and graph topology heuristics.

## Advanced Features

- **Graph Export**: `--save-graph` writes the underlying DAG-based clustering structure in GraphML format for inspection or debugging.
- **Ratio Logging**: `--save-ratio` exports edge abundance ratios (used for error vs. variant decisions) in CSV format.
- **Mutation Rate Calibration**: `--min-eval-rate` sets the minimum read count required before estimating sequencing error/mutation rates.

## Integration

- Extends `obiconvert` input/output options, supporting standard FASTA/FASTQ formats and metadata handling.
- Uses the `sample` attribute (configurable via `-s`) to associate sequences with biological samples.

## Design Notes

- Clustering mode (`--cluster`, currently commented out) would annotate sequences with true cluster membership.
- Default thresholds prioritize sensitivity: `distance=1`, `ratio=1.0` (i.e., any less-abundant sequence is considered an error), `min-sample=1`.
