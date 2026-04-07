# Functional Overview of the `obicsv` Package  

The `obicsv` package enables efficient, configurable export of biological sequence data (e.g., FASTA/FASTQ) to CSV format. It supports selective column inclusion, parallel batch processing, compression, and seamless CLI integration—ideal for high-throughput NGS pipelines.

## Core Capabilities  

| **Domain** | **Features** |
|-----------|--------------|
| **Column Selection & Formatting** | Toggle output fields (`CSVId`, `CSVSequence`, `CSVTaxon`, etc.); define custom attributes via `CSVKey`/`CSVKeys`; set separator (`CSVSeparator`) and NA placeholder (`CSVNAValue`). |
| **I/O & File Handling** | Write to stdout or file (append/truncate); support gzip compression (`OptionsCompressed`); configure batch size and full-file batching. |
| **Processing Strategy** | Parallel workers (default: `obidefault.ParallelWorkers()`); unordered iteration (`NoOrder`); progress tracking; skip empty sequences. |
| **Metadata Enrichment** | Auto-detect columns (`CSVAutoColumn`); integrate `obipairing`, taxonomic data, and abundance counts; support Phred+shifted quality scores. |
| **CLI Integration** | Command-line flags (`--ids`, `--sequence`, `--taxon`, etc.); extendable via helper functions (`CLIPrintId()`, `CLIHasToBeKeptAttributes()`). |

## Public API Summary  

- **`MakeOptions([]WithOption)`**  
  Builder-style configuration of export behavior. Supported options: `CSVId`, `CSVTaxon`, `OptionsFileName`, `OptionAppendFile`, etc.

- **`NewCSVSequenceIterator(IBioSequence, ...WithOption)`**  
  Wraps a sequence iterator into an async CSV record stream. Launches parallel workers, handles batching, and auto-detects attributes when enabled.

- **`CSVSequenceHeader(Options)`**  
  Generates a CSV header row based on enabled columns and custom keys.

- **`CSVBatchFromSequences(BioSequenceBatch, Options)`**  
  Converts a batch of sequences into `CSVRecord` entries per configured options.

- **`WriteCSV(ICSVRecord, io.WriteCloser)`**  
  Writes CSV data to any writer with compression and parallelization support.

- **`WriteCSVToStdout()`, `WriteCSVToFile()`**  
  Convenience wrappers for common I/O targets.

- **`FormatCVSBatch(CSVRecordBatch, string)`**  
  Renders a batch of records as an in-memory CSV buffer (header prepended only for first chunk).

## Design Principles  

- **Streaming & Laziness**: Uses iterator patterns to avoid full data loading.  
- **Parallelism**: Producer-consumer model with configurable concurrency (min 2 workers).  
- **Resilience**: Graceful handling of missing fields via configurable NA values.  
- **Extensibility**: Supports dynamic attributes (e.g., `obipairing` expands to 8 fields).  

## Usage Example  
```go
opt := MakeOptions([]WithOption{
    OptionFileName("results.csv"),
    CSVId(true), 
    CSVTaxon(false),
    OptionsAppendFile(true),
})
iter := NewCSVSequenceIterator(sourceIter, opt)
WriteCSV(iter, os.Stdout) // or file
```
