# `obisplit` Package Overview

The `obisplit` package provides functionality to split sequencing reads based on user-defined molecular tags (e.g., PCR or sample barcodes), using pattern-matching with configurable error tolerance.

- **Configuration via CSV**: Reads a configuration file (CSV format) mapping `tag` sequences to `pcr_pool` names.
- **Pattern Compilation**: Uses the `obiapat` module to compile tag sequences into fuzzy pattern matchers, allowing mismatches and optionally indels.
- **Reverse Complement Support**: Automatically computes reverse-complemented versions of patterns for dual-indexed or stranded workflows.
- **CLI Integration**: Integrates with `getoptions` to define command-line flags:
  - `-C`, `--config`: Specify the configuration CSV file.
  - `--template`: Output a sample config template to stdout (for quick start).
  - `--pattern-error N`: Set max allowed mismatches in pattern matching (default: 4).
  - `--allows-indels`: Enable indel-aware matching.
- **Error Handling**: Logs fatal errors on invalid config (missing `tag` column), failed pattern compilation, or file access issues.
- **Extensibility**: Extends `obiconvert.OptionSet`, suggesting compatibility with broader OBITools4 conversion pipelines.

The core data structure `SplitSequence` stores parsed tag metadata (name, forward/reverse patterns) for downstream splitting logic.
