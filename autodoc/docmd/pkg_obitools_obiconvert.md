# `obiconvert`: Semantic Overview of Public Functionalities

The `obiconvert` package provides a robust, CLI-driven framework for converting and managing biological sequence data within the OBITools4 ecosystem. It enables format-agnostic input parsing, standardized output generation (FASTA/FASTQ/JSON), and configurable preprocessing—while preserving metadata semantics.

## Input Handling

- **`ExpandListOfFiles(check_ext bool, filenames ...string) []string`**  
  Expands file paths into a deduplicated list of eligible files. Supports local directories, symlinks (resolved), and remote URLs (`http(s)://`, `ftp://`).  
  When `check_ext=true`, filters files by extension: `.fasta[.gz]`, `.fastq[.fq][.gz]`, `.seq[.gz]`, `.gb[| gbff | dat ][.gz]`, and `.ecopcr[.gz]`.

- **`CLIReadBioSequences(filenames ...string) obiiter.IBioSequence`**  
  Returns a lazy, streaming iterator over biological sequences from files or stdin. Automatically selects parsing strategy based on CLI flags:
  - JSON-style (`--input-json-header`)
  - OBI-compliant headers (`--input-OBI-header`, `--input-obi`)
  - Heuristic auto-detection (default).
  
  Configurable via CLI options:
  - Parallel workers (`nworkers ≥ 2`)
  - Batch size and memory limits
  - `U→T` conversion for RNA (`--u-to-t`)
  - Skip empty sequences (`--skip-empty`)
  
  Handles:
  - Single/multiple files (with batched parallel reading)
  - Paired-end input via `--paired-with`
  - Fallback readers: FASTA, FASTQ, GenBank/EMBL, ecoPCR output, CSV

- **`OpenSequenceDataErrorMessage(args ...string, err error)`**  
  Formats and logs user-friendly errors for input failures (stdin-only / single-file / multi-file), then exits with status `1`.

## Output Handling

- **`CLIWriteBioSequences(iter obiiter.IBioSequence, filenames ...string)`**  
  Writes sequences from an `IBioSequence` iterator to stdout or files, based on CLI options:
  - **Format**: FASTQ (if quality scores present), FASTA, JSON (default), or generic sequence.
  - **Header style**: Configured via `CLIOutputFastHeaderFormat()` → `"json"` or `"obi"`.
  - **Compression**: Optional gzip (`--gzip`).
  - **Paired-end output**: Automatically splits into `_R1`, `_R2` files via `BuildPairedFileNames`.
  - **Parallelism**: Uses configurable workers (`WriteParallelWorkers()`).
  
- **`BuildPairedFileNames(filename string) (string, string)`**  
  Generates paired-end filenames: `sample.fastq → sample_R1.fastq`, `sample_R2.fastq`.

## Configuration & Integration

- **`OptionSet(allow_paired bool)`**  
  Centralized CLI option setter. Enables modular setup for paired-end support and shared flags.

- **Taxonomy Integration**:  
  Supports loading taxonomy via `obioptions.LoadTaxonomyOptionSet`.

- **Progress Reporting**:  
  Displays a progress bar unless stderr is redirected or stdout pipes to another process.

## Design Principles

✅ Lazy evaluation via iterators for memory efficiency  
✅ Automatic format inference and parallel I/O scaling  
✅ Symlink resolution, recursive globbing with extension filtering  
✅ CLI-integrated configuration (header parsing mode, workers, batch size)  

All functionality is exposed through public functions and designed for composability with `obiformats`, `obiiter`, and `obidefault`.
