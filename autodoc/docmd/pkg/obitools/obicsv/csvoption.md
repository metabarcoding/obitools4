# Functional Overview of the `obicsv` Package

The `obicsv` package provides a flexible and configurable interface for processing biological sequence data (e.g., FASTA/FASTQ) with support for CSV export and parallelized batch processing.

## Core Concepts

- **Options Pattern**: Uses a builder-style API via `MakeOptions([]WithOption)` to configure behavior.
- **Configurable Processing**: Supports batch size, parallel workers, file I/O mode (append/new), compression handling, and progress tracking.
- **Selective CSV Export**: Fine-grained control over output columns (ID, sequence, quality, taxon, count, definition) and formatting (separator, NA value, custom keys).
- **Default Integration**: Leverages `obidefault` for sensible defaults (e.g., batch size, parallel workers).

## Key Functionalities

| Category | Features |
|---------|----------|
| **I/O Control** | File name, append vs overwrite (`OptionsAppendFile`, `OptionCloseFile`), compression support (`OptionsCompressed`) |
| **Processing Strategy** | Batch size, full-file batch mode (`FullFileBatch`), parallel workers (`ParallelWorkers`), unordered processing (`NoOrder`) |
| **Data Handling** | Skip empty sequences (`SkipEmptySequence`), progress bar display, source tracking |
| **CSV Output Customization** | Toggle columns (`CSVId`, `CSVSequence`, etc.), custom keys via `CSVKey`/`CSVKeys`, separator (`CSVSeparator`) and NA placeholder (`CSVNAValue`), auto-column detection |

## Usage Example

```go
opt := MakeOptions([]WithOption{
  OptionFileName("output.csv"),
  CSVId(true),
  CSVSequence(true),
  CSVTaxon(false),
  OptionsAppendFile(true),
})
```

This package enables efficient, customizable conversion of biological sequence data to structured CSV format with minimal boilerplate.
