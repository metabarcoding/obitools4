# CSV Export Functionality in `obicsv` Package

The `obicsv` package provides utilities to convert biological sequence data into structured CSV format. It supports flexible, configurable output through an `Options` interface.

## Core Functions

- **`CSVSequenceHeader(opt Options)`**:  
  Constructs a CSV header row based on enabled options (e.g., `id`, `count`, `taxid`, `definition`). Additional user-defined attributes are appended, followed by optional `sequence` and `qualities`.

- **`CSVBatchFromSequences(batch BioSequenceBatch, opt Options)`**:  
  Converts a batch of biological sequences into CSV records. Each sequence is processed according to the active options:
  - Sequence ID, count, taxonomic identifier (from `Taxon()` or fallback to raw `taxid`), and definition.
  - Custom attributes retrieved via `GetAttribute(key)`; missing values replaced by a configurable NA value.
  - Nucleotide sequence (as string) and quality scores (converted to ASCII Phred+shifted format or NA if absent).

- **`NewCSVSequenceIterator(iter IBioSequence, options ...WithOption)`**:  
  Wraps a sequence iterator (`IBioSequence`) to produce an asynchronous CSV record stream:
  - Optionally auto-detects and includes all sequence attributes (`CSVAutoColumn`).
  - Launches parallel workers to process batches concurrently.
  - Uses a producer-consumer pattern: one goroutine drives iteration, others write CSV records.

## Key Features

- **Configurable output columns** via option flags (e.g., `CSVId()`, `CSVTaxon()`).
- **Support for quality scores** in standard FASTQ ASCII encoding.
- **NA value handling**: missing fields replaced with a user-defined placeholder (e.g., `"."`).
- **Parallelization**: scalable CSV generation using multiple goroutines.
