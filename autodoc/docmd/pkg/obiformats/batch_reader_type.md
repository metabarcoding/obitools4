# `obiformats` Package — Semantic Overview

The `obiformats` package provides a standardized interface for **format-agnostic batch reading of biological sequence data** within the OBITools4 ecosystem.

## Core Abstraction

- **`IBatchReader`** is a function type defining the contract for opening and iterating over sequence files:
  ```go
  func(string, ...WithOption) (obiiter.IBioSequence, error)
  ```
- It accepts:
  - A file path (`string`)
  - Optional configuration via variadic `WithOption` arguments (e.g., filtering, parsing rules)
- Returns:
  - An iterator over biological sequences (`obiiter.IBioSequence`)
  - Or an error if the file cannot be opened/parsed

## Semantic Intent

- **Decouples format handling from iteration logic**: Enables uniform consumption of FASTA, FASTQ, SAM/BAM, etc., via a single entry point.
- **Supports extensibility**: New format readers can be registered as `IBatchReader` implementations without altering client code.
- **Enables lazy, streaming access**: Sequences are yielded on-demand via the iterator—memory-efficient for large datasets.

## Typical Usage Pattern

1. Select or compose an `IBatchReader` implementation (e.g., for FASTQ).
2. Call it with a file path and optional options.
3. Iterate over the returned `IBioSequence` to process sequences one-by-one.

## Design Principles

- **Functional, minimal API**: Single responsibility—reading and iteration.
- **Option-based configurability**: Avoids combinatorial function overloading via `With...` patterns.
- **Integration-ready**: Built to work seamlessly with the broader OBITools4 iterator and sequence abstractions.

> *Note: Actual format-specific readers (e.g., `NewFASTQBatchReader`) are expected to conform to this interface but reside outside the core type definition.*
