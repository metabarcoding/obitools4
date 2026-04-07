# `obicount` Package Functional Overview

The `obicount` package provides CLI option parsing and state management for the `obicount` utility, which counts biological sequence metrics from input files (e.g., FASTA/FASTQ). It leverages `go-getoptions` for argument parsing.

## Core Features

- **Three counting modes**:
  - `--reads` (`-r`) — count total reads (sequences).
  - `--variants` (`-v`) — count unique sequence variants.
  - `--symbols` (`-s`) — sum of all nucleotide/amino-acid symbols (i.e., total length).

- **Default behavior**:  
  If *no* flag is specified, all three counts are printed (i.e., fallback to full report).

- **State variables** (`__read_count__`, `__variant_count__`, `__symbol_count__`) track which metrics are enabled.

- **Helper functions**:
  - `CLIIsPrintingReadCount()` — returns true if read count should be output.
  - `CLIIsPrintingVariantCount()` — same for variant counts.
  - `CLIIsPrintingSymbolCount()` — same for symbol (length) totals.

- **Semantic semantics**:  
  Each function returns `true` if explicitly requested *or* when no flags are set (default mode), ensuring backward compatibility and intuitive CLI behavior.

This package encapsulates only the option-handling logic, keeping concerns separated from file I/O or counting implementation.
