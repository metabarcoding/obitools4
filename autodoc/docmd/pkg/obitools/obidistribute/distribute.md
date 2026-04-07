# `obidistribute` Package: Sequence Distribution and Output Formatting

This Go module provides functionality to distribute biological sequences across multiple output files based on classification criteria, while applying configurable formatting and parallelization options.

- **Main Function**: `CLIDistributeSequence()` orchestrates the entire process.
- It accepts an iterator (`obiiter.IBioSequence`) of biological sequences as input.

## Key Features

- **Header Format Selection**:  
  Supports JSON or OBI-compliant headers via `obiconvert.CLIOutputFastHeaderFormat()`; defaults to JSON.

- **Parallel Processing**:  
  Automatically configures worker threads (at least 2), derived from `obidefault.ParallelWorkers()` divided by four.

- **Batching & Compression**:  
  Uses configurable batch size and output compression settings from defaults (`obidefault`).

- **Output Format Handling**:  
  Supports FASTQ, FASTA, or generic sequence formats (`WriteSequencesToFile`), selected via `CLIOutputFormat()`.

- **Sequence Classification & Dispatching**:  
  Sequences are classified using `CLISequenceClassifier()`, enabling distribution into multiple files based on metadata (e.g., sample, taxon).

- **File Naming & Appending**:  
  Output filenames follow a pattern (`CLIFileNamePattern()`), with optional append mode via `CLIAppendSequences()`.
