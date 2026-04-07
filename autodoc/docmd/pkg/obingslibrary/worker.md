# PCR Simulation and Barcode Extraction Module

This Go package (`obingslibrary`) provides configuration-driven tools for **PCR simulation and barcode extraction** from NGS libraries.

## Core Concepts

- `Options`: A fluent configuration object for customizing behavior via functional setters.
- Default options are defined in `MakeOptions`, supporting:
  - Error handling (`discardErrors`)
  - Mismatch/indel tolerance via `allowedMismatch` and `allowsIndel`
  - Parallelization (`parallelWorkers`) and batching control (`batchSize`)
  - Optional progress tracking (`withProgressBar`)

## Key Functionalities

- **Barcode Extraction**:
  - `ExtractBarcodeSlice`: Extracts barcodes from a slice of sequences using the NGS library, applying configured error handling and alignment parameters.
  - `ExtractBarcodeSliceWorker`: Returns a reusable worker function for batch processing (e.g., in pipelines or parallel workers).

- **Compilation Step**:
  - `ngslibrary.Compile(...)` prepares internal indexing based on mismatch/indel settings before extraction.

- **Error Handling**:
  - If `discardErrors` is true (default), sequences causing extraction errors are filtered out.
  - Alternatively, error-containing reads can be retained or logged via `OptionUnidentified`.

## Design Highlights

- Uses the *option pattern* for extensibility and clean API.
- Integrates with default settings from `obidefault` (e.g., parallelism, batch size).
- Designed for both direct use and integration into concurrent workflows.
