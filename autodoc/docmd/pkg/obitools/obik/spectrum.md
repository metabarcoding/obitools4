# `obik spectrum` Command — Semantic Description

The `runSpectrum` function implements the `obik spectrum` subcommand, which computes and exports **k-mer frequency spectra** from indexed k-mer sets.

## Core Functionality

- Opens a pre-built **k-mer index** (a `KmerSetGroup`) from disk using the provided directory path.
- Selects one or more k-mer sets via pattern matching (e.g., `set1`, `group_*`) or defaults to *all* sets if none specified.
- For each selected set, retrieves its **k-mer frequency spectrum**, i.e., a mapping from *frequency* (how many times each k-mer appears across samples) to the count of distinct k-mers at that frequency.

## Output Format

- Generates a **CSV file** (or `stdout` if `-`) with:
  - First column: frequency value (`1`, `2`, ..., up to the maximum observed).
  - Subsequent columns: number of k-mers at that frequency, *per selected set*.
- Only rows where **at least one set has non-zero counts** are written (sparse output).
- Column headers use the actual k-mer set IDs where available; otherwise fall back to `set_N`.

## Design Highlights

- Gracefully handles missing spectrum data (logs warning, uses empty map).
- Efficiently tracks `maxFreq` to avoid unnecessary zero-padding.
- Uses structured logging (`logrus`) for diagnostics (e.g., missing data).
- Compliant with CLI conventions: supports `--output`, pattern-based set selection, and context-aware cancellation (`context.Context`).

## Use Case

Enables comparative analysis of k-mer distributions across multiple sequencing libraries or sample groups—e.g., to assess redundancy, complexity, or contamination levels in metabarcoding data.
