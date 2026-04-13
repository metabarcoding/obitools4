# `obimultiplex` Package Functionalities

The `obimultiplex` package provides command-line and programmatic interfaces for simulating and processing multiplexed PCR amplicon sequencing data, primarily using the `NGSFilter` format.

## Core Features

- **PCR Multiplex Configuration Parsing**: Reads and interprets CSV-based `NGSFilter` files that define experiments, samples, tags (barcodes), and primer sequences.
- **Flexible Primer Matching**:
  - Supports `strict`, `hamming`, and `indel` matching algorithms.
  - Configurable mismatch tolerance (default: ≤2 mismatches).
  - Optional indel allowance during primer alignment.
- **Tag Assignment & Error Handling**:
  - Assigns reads to samples based on tag-primer matching.
  - Outputs unassigned sequences to a dedicated file (if specified).
- **Template & Configuration Support**:
  - Generates and displays an example `NGSFilter` CSV template via CLI.
- **Extensible Annotation**:
  - Allows extra columns in the `NGSFilter` file to annotate sequences with key-value metadata.

## CLI Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--tag-list` / `-s` |  | Path to the NGSFilter CSV file |
| `--allowed-mismatches` / `-e` |  | Max mismatches allowed for primer matching (default: `2`) |
| `--with-indels` |  | Permit indel errors during matching (default: `false`) |
| `--unidentified` / `-u` |  | Output file for reads failing sample assignment |
| `--keep-errors` / `--conserved-error` |  | Retain error information in output (affects annotation) |
| `--template` |  | Print a sample CSV configuration template to stdout |

## Implementation Notes

- Built on top of `obitools4` libraries for formats (`obiformats`) and NGS library handling (`obingslibrary`).  
- Uses `go-getoptions` for CLI argument parsing and Logrus for logging.
- Designed to be composable: integrates with `obiconvert.OptionSet()` via the `OptionSet` wrapper.
