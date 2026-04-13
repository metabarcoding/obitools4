# `obigrep.CLIFilterSequence` — Semantic Description

The function `CLIFilterSequence` implements a **command-line-driven sequence filtering pipeline** over an iterator of biological sequences (`IBioSequence`). It selectively retains or discards reads based on user-defined criteria, optionally saving discarded sequences to disk.

## Core Functionality

- **Predicate Construction**:  
  Builds a filtering predicate via `CLISequenceSelectionPredicate()`, which encodes user-specified filters (e.g., min/max length, quality thresholds, primer matches).

- **Paired-End Support**:  
  If input is paired-end (`CLIHasPairedFile()`), the predicate is extended with a `PairedPredicat` configured by `CLIPairedReadMode()` (e.g., strict pairing, orphan handling).

- **Filtering Strategy**:
  - If a predicate exists:  
    → `iterator.FilterOn(...)` applies the filter in parallel/batched mode (configurable via batch size and worker count).  
    → Alternatively, if `CLISaveDiscardedSequences()` is enabled:  
       `iterator.DivideOn(...)` splits sequences into *kept* and *discarded*, with discarded reads asynchronously written to a file (`CLIDiscardedFileName()`).
  - If no predicate is defined:  
    → The original iterator is returned unchanged.

- **Logging & Error Handling**:  
  Uses `logrus` to log discarded-file destination and fatal errors during write operations.

## Semantic Role

This function acts as the **central filtering engine** in CLI tools (e.g., `obigrep`), translating user flags into a type-safe, composable sequence filter—supporting both single- and paired-end data with optional discard logging.
