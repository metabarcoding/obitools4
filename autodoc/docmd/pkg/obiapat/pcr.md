# PCR Simulation Module (`obiapat`)

This Go package implements a **PCR (Polymerase Chain Reaction) simulation algorithm** for biological sequence analysis. It supports flexible primer matching, amplicon extraction with optional flanking extensions, and handles both linear and circular DNA topologies.

## Key Functionalities

- **Primer Matching**: Accepts forward/reverse primers with configurable mismatch tolerance (`OptionForwardPrimer`, `OptionReversePrimer`). Internally builds pattern objects and their reverse complements.
- **Amplicon Extraction**: Identifies valid amplicons bounded by primer pairs, respecting user-defined length constraints (`OptionMinLength`, `OptionMaxLength`).
- **Extension Support**: Optionally adds fixed-length flanking regions (`OptionWithExtension`) — either strict full-extension only or partial trimming allowed.
- **Topology Handling**: Supports linear (`Circular: false`) and circular DNA sequences via `OptionCircular`.
- **Batch & Parallel Processing**: Configurable batch size (`OptionBatchSize`) and parallel workers count (`OptionParallelWorkers`), enabling efficient processing of large datasets.
- **Annotation-Rich Output**: Each amplicon includes detailed annotations (primer sequences, match positions, errors, direction), preserving original sequence metadata.

## Core API

- `PCRSim(sequence, options...)`: Simulates PCR on a single sequence.
- `PCRSlice(sequencesSlice, options...)`: Applies simulation across multiple sequences in a slice.
- `PCRSliceWorker(options...)`: Returns a reusable worker function for parallel execution via `obiseq.MakeISliceWorker`.

## Implementation Details

- Uses pattern-matching (`ApatPattern`) with fuzzy search to locate primers.
- Handles circular topology by wrapping indices around sequence boundaries.
- Reuses internal memory via `MakeApatSequence`/`Free`, supporting efficient GC and large-scale processing.
- Logs critical errors with `logrus`; debug-level details for amplicon generation.

Designed to integrate within the OBITools4 ecosystem, this module enables high-fidelity *in silico* PCR for metabarcoding and NGS data validation workflows.
