# `obidistribute` Package: Semantic Description of Public Functionalities

The `obidistribute` package enables flexible, scalable distribution of biological sequence data into multiple output files or directories. It supports annotation-based separation (e.g., sample IDs), batch splitting, and hash-sharded distribution — all while integrating with standard NGS formats (FASTA/FASTQ) and the broader `obitools4` ecosystem.

## Core Functionalities

### 1. **Sequence Distribution Strategy Selection**
- `CLIOutputFormat()`  
  Specifies output format: `"fastq"`, `"fasta"`, or generic sequence (e.g., FASTA-like). Controls how sequences are serialized to disk.

- `CLISequenceClassifier()`  
  Selects the annotation key used for classification (e.g., `"sample_id"`, `"taxon"`). Sequences are grouped by the value of this annotation field.

- `CLIDistributeByBatches()` / `-n`  
  Enables round-robin assignment of sequences into *N* fixed batches, regardless of metadata.

- `CLIDistributeByHash()` / `-H`  
  Distributes sequences deterministically into *N* batches using a hash of the sequence ID or annotation — ensures reproducible sharding.

- `CLIDirectoryMode()` / `-d`  
  When used with a classifier, organizes output files into subdirectories named after classification values.

### 2. **Output Naming & File Management**
- `CLIFileNamePattern()` / `-p`  
  Defines a printf-style filename template (e.g., `"sample_%s.fastq"`), where `%s` is replaced by the classifier value or batch index.

- `CLIAppendSequences()` / `-A`  
  Enables appending to existing files instead of overwriting them.

- `CLINAValue()` / `--na-value`  
  Sets the fallback label (default: `"NA"`) for sequences missing a classifier annotation.

### 3. **Processing Configuration**
- `CLIDistributeSequence()`  
  Main entry point: orchestrates input iteration, classification, batching, and parallelized writing. Accepts an `obiiter.IBioSequence` iterator.

- Parallel workers are derived from `obidefault.ParallelWorkers()` (minimum 2), divided by four.

- Batch size and compression settings are inherited from `obidefault`.

### 4. **Header & Format Handling**
- `CLIOutputFastHeaderFormat()`  
  Configures header serialization format: `"json"` (default) or `"obi"`. Controls metadata inclusion in output headers.

- Built on top of `obiconvert.CLIOutputFastHeaderFormat()` and integrates with its header parsing/writing logic.

### 5. **Validation & CLI Integration**
- Uses `getoptions` for option parsing; enforces mutual exclusivity of distribution modes.
- Validates filename pattern syntax and required arguments at startup.

## Semantic Workflow

1. User selects a distribution mode (`classifier`, `batches`, or `hash`) and optional directory nesting.
2. Sequences are read via an iterator; each is classified or assigned to a batch/shard.
3. Sequences are buffered in batches, compressed if configured, and written to output files using the selected format.
4. Filenames are generated dynamically per classification value or batch index, respecting append mode and NA fallbacks.

This module is essential for demultiplexing, batch processing, and scalable data management in high-throughput sequencing pipelines — especially metabarcoding workflows.
