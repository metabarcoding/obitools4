# `obik super`: Super K-mer Extraction Tool

The `runSuper` function implements the `obik super` subcommand, a bioinformatics utility for extracting *super k-mers* from DNA sequences. Super k-mers are maximal non-overlapping contiguous regions formed by merging overlapping *k*-mers that share a common minimizer—enabling efficient sequence compression and alignment-free analysis.

## Key Features

- **Configurable K-mer & Minimizer Sizes**:  
  Accepts `k` (k-mer length, range: [2–31]) and `m` (minimizer size, range: [1, *k*−1]), validated at runtime.

- **Sequence Input Handling**:  
  Reads biological sequences (FASTA/FASTQ) via `obiconvert.CLIReadBioSequences`, supporting multiple file arguments and standard I/O.

- **Parallel Processing**:  
  Uses a worker-based pipeline (`MakeIWorker`) with configurable parallelism (via `obidefault.ParallelWorkers()`), enabling scalable performance on large datasets.

- **Super K-mer Generation**:  
  Leverages `obikmer.SuperKmerWorker(k, m)` to process each input sequence and emit merged super k-mers—preserving biological context while reducing redundancy.

- **Output Streaming**:  
  Writes results via `obiconvert.CLIWriteBioSequences`, supporting standard output and optional compression; ensures pipeline completion with `obiutils.WaitForLastPipe()`.

- **Logging & Error Handling**:  
  Uses structured logging (Logrus) for operational transparency and robust error reporting with contextual messages.

This tool supports applications in metagenomics, sequence assembly, read correction, and approximate matching—where compact representation of sequencing data is essential.
