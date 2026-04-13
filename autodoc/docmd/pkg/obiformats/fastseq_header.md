# Semantic Description of `obiformats` Package

The `obiformats` package provides utilities for parsing sequence headers in the OBItools4 framework, supporting two distinct formats:

- **JSON-based format** (e.g., `{"id":"seq1", ...}`): Detected by a leading `{` character.
- **Legacy OBI format** (plain text, e.g., `>seq1 description`): Used when no JSON prefix is present.

## Core Functions

- **`ParseGuessedFastSeqHeader(sequence *obiseq.BioSequence)`**  
  Dynamically routes header parsing based on the first character of the sequence definition:
  - Calls `ParseFastSeqJsonHeader` if JSON-prefixed.
  - Otherwise invokes `ParseFastSeqOBIHeader`.

- **`IParseFastSeqHeaderBatch(iterator, options...) obiiter.IBioSequence`**  
  Applies header parsing to a *batch* of sequences:
  - Takes an iterator over `BioSequence`s.
  - Uses optional configuration (e.g., parallelism, parsing behavior).
  - Wraps the parser in a worker pipeline via `MakeIWorker`, preserving sequence flow.

## Design Principles

- **Format agnosticism**: Automatically detects header type.
- **Iterator-based streaming**: Enables memory-efficient batch processing of large datasets (e.g., FASTQ/FASTA).
- **Extensibility**: Options pattern (`WithOption`) supports runtime customization.

This package serves as a header-decoding layer for downstream analysis in metagenomic or metabarcoding workflows.
