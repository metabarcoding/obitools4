# `obidemerge` Package Overview

The `obidemerge` package provides functionality to **split biological sequences based on metadata statistics**, commonly used in metabarcoding data processing.

- `MakeDemergeWorker(key string)` returns a `SeqWorker` that processes each sequence as follows:
  - Checks if the sequence contains statistical data associated with `key`.
  - If present, retrieves a map of value → count (e.g., `"speciesA": 12`, `"speciesB": 7`).
  - Creates a new slice of sequences: one copy per unique statistical key, each assigned:
    - The original sequence data (copied),
    - A new attribute `key = <stat_key>`,
    - Count set to the corresponding statistical value.
  - Removes original stats from input sequence after processing.

- If no statistics are found for `key`, the sequence is returned unchanged in a single-element slice.

- `CLIDemergeSequences(iterator)` wraps the worker for CLI use:
  - Uses a default slot name (`CLIDemergeSlot()`, likely `"demerged"` or similar).
  - Applies the worker to an iterator of sequences, optionally in parallel.

**Use case**: Converts aggregated statistics (e.g., from clustering or OTU picking) into discrete, count-annotated sequences — enabling downstream tools to treat each variant as an independent entity with its own abundance.

**Key concept**: *Demerging* = reversing a prior merging step by expanding merged sequences into their constituent statistical components.
