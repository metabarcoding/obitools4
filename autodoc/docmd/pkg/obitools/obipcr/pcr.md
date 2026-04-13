# `obipcr.CLIPCR`: Amplicon Extraction via In-Silico PCR

The `CLIPCR` function performs *in-silico*PCR on biological sequences to extract amplicons using user-defined primer settings.

## Core Functionality

- **Input**: An iterator over biological sequences (`obiiter.IBioSequence`).
- **Output**: A new sequence iterator yielding batches of amplified fragments (`obiiter.IBioSequence`).
- **Algorithm**: Uses `PCRSliceWorker`, configured via a set of options derived from CLI parameters.

## Primer Configuration

- **Forward/Reverse Primers**: Specified via `CLIForwardPrimer()` and `CLIReversePrimer()`.
- **Mismatch Tolerance**: Controlled by `CLIAllowedMismatch()` for both primers.

## Amplification Constraints

- **Full Extension**: Only full-length amplicons (spanning between primers) are returned if `CLIOnlyFull()` is enabled.
- **Length Filtering**:
  - Minimum length: enforced if `CLIMinLength() > 0`.
  - Maximum length: always applied via `CLIMaxLength()`.

## Optional Features

- **Extension**: If enabled (`CLIWithExtension()`), flanking regions beyond primers are included, using `CLIExtension()`.
- **Circular Genomes**: Supports circular DNA via `CLICircular()`.

## Large Sequence Handling

- Long sequences (>`CLIMaxLength()*1000`) are fragmented into overlapping chunks (`~CLIMaxLength()*100` bp) to improve PCR efficiency.
- Fragmentation parameters are logged for transparency.

## Execution Model

- Memory usage is capped at 50% (`LimitMemory(0.5)`).
- Parallelized processing via `obidefault.ParallelWorkers()`.

## Summary

`CLIPCR` enables flexible, robust *in-silico* PCR with support for mismatches, partial amplification, circular templates, and large-input fragmentation—ideal for metagenomic amplicon processing pipelines.
