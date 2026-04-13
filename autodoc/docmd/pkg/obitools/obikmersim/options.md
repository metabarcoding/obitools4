# Semantic Description of `obikmersim` Package

The `obikmersim` package provides command-line interface (CLI) configuration and utility functions for k-mer–based sequence similarity analysis, particularly in the context of *in silico* PCR or read-matching workflows.

- **K-mer Counting Options** (`KmerSimCountOptionSet`):  
  Configures parameters for k-mer extraction and comparison: `--kmer-size`, sparse mode (`--sparse`), reference sequences (`--reference`), minimum shared k-mers threshold, and optional self-comparison.

- **K-mer Matching Options** (`KmerSimMatchOptionSet`):  
  Adds alignment-free scoring parameters: `--delta`, mismatch/gap scaling (`--penalty-scale`), gap penalty factor, and a fast absolute scoring mode.

- **Combined Option Sets**:  
  `CountOptionSet` and `MatchOptionSet` integrate k-mer settings with generic conversion options (e.g., input/output format handling via `obiconvert`).

- **CLI Accessors**:  
  Helper functions (e.g., `CLIKmerSize`, `CLIReference`) retrieve parsed values and load reference sequences from files, supporting batched/parallel reading.

- **Core Use Case**:  
  Enables efficient k-mer–based sequence matching (e.g., for taxonomic assignment or PCR primer specificity checks), balancing sensitivity and performance via tunable thresholds, sparse representations, and scalable scoring.
