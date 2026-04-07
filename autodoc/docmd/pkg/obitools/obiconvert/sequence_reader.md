# Semantic Description of `obiconvert` Package Functionality

The `obiconvert` package provides utilities for robust, scalable input handling of biological sequence data in the OBITools4 ecosystem.

- **`ExpandListOfFiles(check_ext, filenames...)`**:  
  Recursively expands file paths into a deduplicated list of eligible files. Supports local directories, symlinks (resolved), and remote URLs (`http(s)://`, `ftp://`).  
  Filters files by extension when `check_ext=true`: accepts `.fasta[.gz]`, `.fastq[.fq][.gz]`, `.seq[.gz]`, `.gb[ gbff|dat ][.gz]`, and `.ecopcr[.gz]`.

- **`CLIReadBioSequences(filenames...)`**:  
  Constructs a streaming iterator (`obiiter.IBioSequence`) over biological sequences from files or stdin.  
  - Adapts parsing strategy based on CLI options: JSON, OBI, or heuristic header parsers.  
  - Configures parallelism (`nworkers ≥ 2`), batch size, memory limits, and quality reading.  
  - Supports full-file batching (for large records) and `U→T` conversion for RNA data.  
  - Handles single/multiple files: uses batched parallel reading when appropriate; supports paired-end input via `PairTo`.  
  - Falls back to format-specific readers (FASTA, FASTQ, GenBank, EMBL, EcoPCR, CSV) or generic fallback.

- **`OpenSequenceDataErrorMessage(args..., err)`**:  
  Formats and logs user-friendly error messages for input failures, then exits with status `1`. Distinguishes stdin-only, single-file, and multi-file error contexts.

Core design principles:  
✅ Lazy evaluation via iterators for memory efficiency.  
✅ Automatic format inference and parallel I/O scaling.  
✅ Symlink resolution, recursive globbing with extension filtering.  
✅ CLI-integrated configuration (header parsing mode, parallel workers, batch settings).
