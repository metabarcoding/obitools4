# `obicount` Package Functional Overview

The `obicount` package provides command-line interface (CLI) option parsing and internal state management for the `obicount` utility — a tool designed to compute biological sequence metrics from standard input formats (e.g., FASTA, FASTQ). Built on top of `go-getoptions`, it cleanly separates argument handling from core counting logic.

## Core Functionalities

### 1. **Counting Modes**
Three mutually exclusive or combinable counting modes are supported via CLI flags:

| Flag | Long Form      | Semantic Meaning                              |
|------|----------------|-----------------------------------------------|
| `-r` | `--reads`      | Count total number of sequences (i.e., reads) |
| `-v` | `--variants`   | Count unique sequence variants (distinct strings) |
| `-s` | `--symbols`    | Sum of all symbol counts (i.e., total length across reads) |

- **Default behavior**: If *none* of the above flags is provided, all three metrics are computed and reported — ensuring backward-compatible full-report output.

### 2. **State Tracking**
Internal state variables track which metrics are active:

- `__read_count__`
- `__variant_count__`
- `__symbol_count__`

Each is set to `true` when its corresponding flag (`--reads`, etc.) appears on the command line, *or* in default mode (no flags), where all are enabled.

### 3. **Public Query Functions**
Three exported helper functions allow runtime introspection of active metrics:

| Function                        | Returns `true` if…                                      |
|---------------------------------|----------------------------------------------------------|
| `CLIIsPrintingReadCount()`      | Read count is enabled (explicitly requested or default) |
| `CLIIsPrintingVariantCount()`   | Variant count is enabled (explicitly requested or default) |
| `CLIIsPrintingSymbolCount()`    | Symbol count is enabled (explicitly requested or default) |

These functions decouple counting logic from CLI parsing, enabling modular and testable design.

### 4. **Semantic Guarantees**
- All query functions follow *inclusive semantics*: they return `true` both when the option is explicitly set and in default mode.
- This ensures intuitive behavior: no flags → full report; any flag subset → only requested metrics.

### 5. **Separation of Concerns**
- The package handles *only* CLI parsing and state management.
- File I/O, sequence decoding (FASTA/FASTQ), counting algorithms, and output formatting reside in separate modules — promoting maintainability and reuse.

## Usage Example (Conceptual)

```bash
obicount -r input.fasta    # prints only read count  
obicount --variants input.fastq  # prints unique variant count only  
obicount -s              # prints total symbol (length) sum  
obicount input.fasta     # prints all three metrics
```

This design supports extensibility, clarity, and robustness in biological sequence analysis pipelines.
