# `obimatrix` Package: Semantic Overview

The `obimatrix` package provides core functionality for generating and formatting sequence count matrices in the OBITools4 ecosystem.

## Core Features

- **Matrix Generation**: Converts sequence annotations into tabular count matrices (samples × features).
  
- **Flexible Output Formats**:
  - *Matrix mode*: Standard rectangular format (rows = sequences, columns = samples).
  - *Three-column mode* (`--three-columns`): Long format with `sample`, sequence ID, and value.

- **Configurable Attributes**:
  - Mapping attribute (default: `"merged_sample"`) used to group sequences per sample.
  - Customizable column names for value (`--value-name`, default `"count"`) and sample ID (`--sample-name`, default `"sample"`).
  - NA handling: Assigns a placeholder value (default `"0"`) when the mapping attribute is missing.

- **Transpose Control** (`--transpose`): Allows switching between sequence-centric and sample-centric layouts.

- **Strictness Option** (`--allow-empty`): Controls whether sequences lacking the mapping attribute are excluded (default: strict).

## Integration

- Extends command-line interface via `getoptions`, aggregating options from:
  - CSV handling (`obicsv.CSVOptionSet`)
  - Input parsing (`obiconvert.InputOptionSet`)

- Exposes getter functions (e.g., `CLIMapAttribute()`, `CLIOutFormat()`), enabling downstream tools to retrieve parsed CLI settings programmatically.

## Use Case

Designed for post-processing amplicon sequencing results, transforming annotated reads into quantitative matrices suitable for ecological or bioinformatic analysis (e.g., diversity studies, differential abundance).
