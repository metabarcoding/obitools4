# `obitagpcr` Package Overview

The `obitagpcr` package provides command-line interface (CLI) support for tag-based PCR data processing within the OBITools4 ecosystem. It defines options and utilities to configure orientation handling of sequencing reads in relation to PCR primers.

## Core Functionality

- **`TagPCROptionSet()`**: Adds a `--reorientate` boolean flag to the CLI option parser. When enabled, it reverse-complements reads as needed so all sequences are stored in a consistent orientation relative to the forward and reverse primers.

- **`OptionSet()`**: Aggregates all required option sets for tag-PCR workflows by extending:
  - `obipairing.OptionSet()` — handling paired-end read pairing options,
  - `obimultiplex.MultiplexOptionSet()` — supporting sample demultiplexing,
  - `TagPCROptionSet()` — adding the reorientation flag.

- **`CLIReorientate()`**: Returns a boolean indicating whether read reorientation is enabled, allowing downstream logic to apply reverse-complementation conditionally.

## Semantic Behavior

- **Reorientation semantics**: Ensures uniform strand orientation across samples—critical for downstream alignment, consensus building, or variant calling where primer directionality matters.

- **Modular design**: Leverages existing OBITools4 modules (`obipairing`, `obimultiplex`) to compose a coherent, reusable CLI configuration for tag-PCR pipelines.

## Use Case

Typically used in amplicon sequencing workflows where:
1. Reads originate from both strands due to PCR amplification,
2. Primer positions are known and fixed (forward/reverse),
3. Consistent orientation improves analysis accuracy.

This package ensures that the `--reorientate` option is available and correctly wired into the processing pipeline.
